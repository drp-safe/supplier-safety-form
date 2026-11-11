package services

import (
	"github.com/jmoiron/sqlx"
	"github.com/obeyler/supplier-safety-form/internal/models"
)

type ThemeService struct {
	db *sqlx.DB
}

func NewThemeService(db *sqlx.DB) *ThemeService {
	return &ThemeService{db: db}
}

func (s *ThemeService) List() ([]models.Theme, error) {
	var themes []models.Theme
	err := s.db.Select(&themes, `
		SELECT code, label, display_order FROM themes ORDER BY display_order ASC, label ASC`)
	if themes == nil {
		themes = []models.Theme{}
	}
	return themes, err
}
