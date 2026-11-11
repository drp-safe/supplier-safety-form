package db

import (
	"fmt"
	"hash/crc32"
	"io/fs"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// RunMigrations applies all pending SQL migrations in order.
// On an existing database (schema_migrations empty but core tables exist),
// all current migrations are baselined as already applied.
// Uses a PostgreSQL advisory lock to prevent concurrent runs in multi-pod
// deployments. The lock number is specific to this application — it must
// never collide with another service's migrator sharing the same cluster.
const migrationAdvisoryLock = 5508312399

func RunMigrations(database *sqlx.DB, migrationsFS fs.FS) error {
	if _, err := database.Exec(fmt.Sprintf("SELECT pg_advisory_lock(%d)", migrationAdvisoryLock)); err != nil {
		return fmt.Errorf("failed to acquire migration lock: %w", err)
	}
	defer database.Exec(fmt.Sprintf("SELECT pg_advisory_unlock(%d)", migrationAdvisoryLock)) //nolint:errcheck

	_, err := database.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version      VARCHAR(10)  NOT NULL,
			description  VARCHAR(200) NOT NULL,
			script       VARCHAR(300) NOT NULL,
			checksum     BIGINT,
			applied_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
			execution_ms INTEGER,
			success      BOOLEAN      NOT NULL,
			error_msg    TEXT,
			PRIMARY KEY (version)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	entries, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	// Baseline detection: tracking table empty but DB already has tables → existing install.
	var count int
	if err := database.Get(&count, "SELECT COUNT(*) FROM schema_migrations"); err != nil {
		return fmt.Errorf("failed to count migrations: %w", err)
	}
	if count == 0 {
		var exists bool
		database.Get(&exists, "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='organizations')") //nolint:errcheck
		if exists {
			log.Println("[migrations] Existing database detected — baselining all migrations as already applied")
			if err := baselineMigrations(database, migrationsFS, files); err != nil {
				return fmt.Errorf("baseline failed: %w", err)
			}
			log.Printf("[migrations] Baselined %d migration(s)", len(files))
			return nil
		}
	}

	applied := 0
	for _, filename := range files {
		version, description := parseMigrationName(filename)

		var success bool
		err := database.Get(&success, "SELECT success FROM schema_migrations WHERE version = $1", version)
		if err == nil && success {
			continue
		}

		content, err := fs.ReadFile(migrationsFS, filename)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}
		checksum := int64(crc32.ChecksumIEEE(content))

		log.Printf("[migrations] Applying %s: %s", version, description)
		start := time.Now()

		tx, err := database.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", filename, err)
		}

		_, execErr := tx.Exec(string(content))
		execMs := time.Since(start).Milliseconds()

		if execErr != nil {
			tx.Rollback()  //nolint:errcheck
			database.Exec( //nolint:errcheck
				`INSERT INTO schema_migrations (version, description, script, checksum, execution_ms, success, error_msg)
				VALUES ($1, $2, $3, $4, $5, false, $6)
				ON CONFLICT (version) DO UPDATE SET success=false, error_msg=$6, applied_at=CURRENT_TIMESTAMP`,
				version, description, filename, checksum, execMs, execErr.Error(),
			)
			return fmt.Errorf("migration %s failed: %w", filename, execErr)
		}

		tx.Commit()    //nolint:errcheck
		database.Exec( //nolint:errcheck
			`INSERT INTO schema_migrations (version, description, script, checksum, execution_ms, success)
			VALUES ($1, $2, $3, $4, $5, true)
			ON CONFLICT (version) DO UPDATE SET success=true, error_msg=NULL, execution_ms=$5, applied_at=CURRENT_TIMESTAMP`,
			version, description, filename, checksum, execMs,
		)
		log.Printf("[migrations] Applied %s in %dms", version, execMs)
		applied++
	}

	if applied > 0 {
		log.Printf("[migrations] %d migration(s) applied successfully", applied)
	} else {
		log.Println("[migrations] Database is up to date")
	}
	return nil
}

func baselineMigrations(database *sqlx.DB, migrationsFS fs.FS, files []string) error {
	for _, filename := range files {
		version, description := parseMigrationName(filename)
		content, err := fs.ReadFile(migrationsFS, filename)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", filename, err)
		}
		checksum := int64(crc32.ChecksumIEEE(content))
		_, err = database.Exec(
			`INSERT INTO schema_migrations (version, description, script, checksum, execution_ms, success)
			VALUES ($1, $2, $3, $4, 0, true)
			ON CONFLICT (version) DO NOTHING`,
			version, description, filename, checksum,
		)
		if err != nil {
			return fmt.Errorf("failed to baseline %s: %w", filename, err)
		}
	}
	return nil
}

func parseMigrationName(filename string) (version, description string) {
	base := strings.TrimSuffix(filename, ".sql")
	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 {
		return base, base
	}
	version = parts[0]
	description = strings.ReplaceAll(parts[1], "_", " ")
	return
}
