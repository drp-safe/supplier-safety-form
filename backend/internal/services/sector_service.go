package services

import (
	"github.com/jmoiron/sqlx"
	"github.com/obeyler/supplier-safety-form/internal/models"
)

type SectorService struct {
	db *sqlx.DB
}

func NewSectorService(db *sqlx.DB) *SectorService {
	return &SectorService{db: db}
}

// ListActive returns the sectors offered at registration — the public,
// unauthenticated listing.
func (s *SectorService) ListActive() ([]models.Sector, error) {
	var sectors []models.Sector
	err := s.db.Select(&sectors, `
		SELECT code, label, display_order, active FROM sectors
		WHERE active = true ORDER BY display_order ASC, label ASC`)
	if sectors == nil {
		sectors = []models.Sector{}
	}
	return sectors, err
}

// ListAll returns every sector, including inactive ones — used by the
// Administration section.
func (s *SectorService) ListAll() ([]models.Sector, error) {
	var sectors []models.Sector
	err := s.db.Select(&sectors, `
		SELECT code, label, display_order, active FROM sectors
		ORDER BY display_order ASC, label ASC`)
	if sectors == nil {
		sectors = []models.Sector{}
	}
	return sectors, err
}

func (s *SectorService) Create(code, label string, displayOrder int) (*models.Sector, error) {
	sector := &models.Sector{Code: code, Label: label, DisplayOrder: displayOrder, Active: true}
	_, err := s.db.Exec(`
		INSERT INTO sectors (code, label, display_order, active) VALUES ($1, $2, $3, true)`,
		code, label, displayOrder)
	if err != nil {
		return nil, err
	}
	return sector, nil
}

func (s *SectorService) Update(code, label string, displayOrder int, active bool) error {
	_, err := s.db.Exec(`
		UPDATE sectors SET label = $1, display_order = $2, active = $3 WHERE code = $4`,
		label, displayOrder, active, code)
	return err
}

// Delete deactivates a sector rather than removing the row: organizations
// already registered under it must keep a valid foreign key.
func (s *SectorService) Delete(code string) error {
	_, err := s.db.Exec(`UPDATE sectors SET active = false WHERE code = $1`, code)
	return err
}
