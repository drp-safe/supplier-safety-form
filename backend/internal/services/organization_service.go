package services

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/obeyler/supplier-safety-form/internal/models"
)

type OrganizationService struct {
	db *sqlx.DB
}

func NewOrganizationService(db *sqlx.DB) *OrganizationService {
	return &OrganizationService{db: db}
}

// CreateWithOwner creates the organization and its owner membership row in a
// single transaction — a registration is never left with one but not the
// other.
func (s *OrganizationService) CreateWithOwner(name, sectorCode, ownerUserID string) (*models.Organization, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	org := &models.Organization{Name: name, SectorCode: sectorCode}
	err = tx.Get(org, `
		INSERT INTO organizations (name, sector_code) VALUES ($1, $2)
		RETURNING id, name, sector_code, created_at`,
		name, sectorCode)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, 'owner')`,
		org.ID, ownerUserID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return org, nil
}

// GetForUser returns the organization the given user belongs to. Étape 1
// gives every user exactly one organization (created at registration); a
// user with none returns (nil, nil).
func (s *OrganizationService) GetForUser(userID string) (*models.Organization, error) {
	var org models.Organization
	err := s.db.Get(&org, `
		SELECT o.* FROM organizations o
		JOIN organization_members m ON m.organization_id = o.id
		WHERE m.user_id = $1
		LIMIT 1`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &org, nil
}
