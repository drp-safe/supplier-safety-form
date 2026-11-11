package services

import (
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/obeyler/supplier-safety-form/internal/crypto"
	"github.com/obeyler/supplier-safety-form/internal/models"
)

type AdminService struct {
	db *sqlx.DB
}

func NewAdminService(db *sqlx.DB) *AdminService {
	return &AdminService{db: db}
}

func (s *AdminService) GetEmailConfig() (*models.EmailConfig, error) {
	config := &models.EmailConfig{}
	if err := s.db.Get(config, `SELECT * FROM email_config LIMIT 1`); err != nil {
		return nil, err
	}

	var err error
	config.SMTPPassword, err = crypto.Decrypt(config.SMTPPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt smtp_password: %w", err)
	}
	config.SMTPUsername, err = crypto.Decrypt(config.SMTPUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt smtp_username: %w", err)
	}
	return config, nil
}

func (s *AdminService) UpdateEmailConfig(update *models.EmailConfigUpdate) error {
	encUsername, err := crypto.Encrypt(update.SMTPUsername)
	if err != nil {
		return fmt.Errorf("failed to encrypt smtp_username: %w", err)
	}
	encPassword, err := crypto.Encrypt(update.SMTPPassword)
	if err != nil {
		return fmt.Errorf("failed to encrypt smtp_password: %w", err)
	}

	_, err = s.db.Exec(`
		UPDATE email_config SET
			smtp_host = $1, smtp_port = $2, smtp_username = $3, smtp_password = $4,
			smtp_encryption = $5, from_address = $6, from_name = $7, updated_at = $8`,
		update.SMTPHost, update.SMTPPort, encUsername, encPassword,
		update.SMTPEncryption, update.FromAddress, update.FromName, time.Now())
	return err
}

// OAuthConfig is the read-only view of the backend's effective OIDC
// configuration, as loaded at startup from the environment — changing it is
// a k8s ConfigMap edit + redeploy, not a form in this app (decided with the
// user: no hot-reload of the OIDC verifier).
type OAuthConfig struct {
	IssuerURL string `json:"issuer_url"`
	ClientID  string `json:"client_id"`
}

func (s *AdminService) GetOAuthConfig() OAuthConfig {
	return OAuthConfig{
		IssuerURL: os.Getenv("KEYCLOAK_ISSUER_URL"),
		ClientID:  os.Getenv("KEYCLOAK_CLIENT_ID"),
	}
}

// DBInfo is the read-only connection info shown to a superadmin so they can
// connect with psql/an external client — never the password.
type DBInfo struct {
	Host   string `json:"host"`
	Port   string `json:"port"`
	Name   string `json:"name"`
	Driver string `json:"driver"`
}

func (s *AdminService) GetDBInfo() DBInfo {
	return DBInfo{
		Host:   os.Getenv("DB_HOST"),
		Port:   os.Getenv("DB_PORT"),
		Name:   os.Getenv("DB_NAME"),
		Driver: "postgresql",
	}
}
