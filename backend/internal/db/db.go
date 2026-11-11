package db

import (
	"fmt"
	"net"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Connect opens a PostgreSQL connection from the DB_* environment variables.
func Connect() (*sqlx.DB, error) {
	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		// Secure by default in production; plaintext only for local/dev.
		if os.Getenv("APP_ENV") == "production" {
			sslmode = "require"
		} else {
			sslmode = "disable"
		}
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")

	// Resolve hostname to IPv4 to avoid IPv6 connectivity issues in-cluster.
	resolvedHost, err := resolveIPv4(host)
	if err != nil {
		resolvedHost = host // fallback to original hostname
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		resolvedHost,
		port,
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		sslmode,
	)

	database, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := database.Ping(); err != nil {
		return nil, err
	}

	return database, nil
}

// resolveIPv4 resolves a hostname to its first IPv4 address.
func resolveIPv4(host string) (string, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return "", err
	}
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String(), nil
		}
	}
	return "", fmt.Errorf("no IPv4 address found for %s", host)
}
