package handlers

import (
	"net/http"

	"github.com/jmoiron/sqlx"
)

type MigrationsHandler struct {
	db *sqlx.DB
}

func NewMigrationsHandler(db *sqlx.DB) *MigrationsHandler {
	return &MigrationsHandler{db: db}
}

type MigrationRecord struct {
	Version     string  `db:"version"      json:"version"`
	Description string  `db:"description"  json:"description"`
	Script      string  `db:"script"       json:"script"`
	Checksum    *int64  `db:"checksum"     json:"checksum"`
	AppliedAt   string  `db:"applied_at"   json:"applied_at"`
	ExecutionMs *int64  `db:"execution_ms" json:"execution_ms"`
	Success     bool    `db:"success"      json:"success"`
	ErrorMsg    *string `db:"error_msg"    json:"error_msg"`
}

// ListMigrations GET /api/admin/migrations — historique en lecture seule.
func (h *MigrationsHandler) ListMigrations(w http.ResponseWriter, r *http.Request) {
	var migrations []MigrationRecord
	err := h.db.Select(&migrations, `
		SELECT version, description, script, checksum,
		       TO_CHAR(applied_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS applied_at,
		       execution_ms, success, error_msg
		FROM schema_migrations
		ORDER BY version ASC
	`)
	if err != nil {
		respondServerError(w, r, "error.failed_list_migrations", err)
		return
	}
	if migrations == nil {
		migrations = []MigrationRecord{}
	}
	respondJSON(w, migrations, http.StatusOK)
}
