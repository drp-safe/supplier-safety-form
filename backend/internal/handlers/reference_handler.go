package handlers

import (
	"net/http"

	"github.com/obeyler/supplier-safety-form/internal/services"
)

// ReferenceHandler exposes the public, unauthenticated reference lists
// (sectors shown at registration, themes used by the future question
// library) — no need to be logged in to list them.
type ReferenceHandler struct {
	sectorService *services.SectorService
	themeService  *services.ThemeService
}

func NewReferenceHandler(sectorService *services.SectorService, themeService *services.ThemeService) *ReferenceHandler {
	return &ReferenceHandler{sectorService: sectorService, themeService: themeService}
}

// ListSectors GET /api/sectors
func (h *ReferenceHandler) ListSectors(w http.ResponseWriter, r *http.Request) {
	sectors, err := h.sectorService.ListActive()
	if err != nil {
		respondServerError(w, r, "error.failed_list_sectors", err)
		return
	}
	respondJSON(w, sectors, http.StatusOK)
}

// ListThemes GET /api/themes
func (h *ReferenceHandler) ListThemes(w http.ResponseWriter, r *http.Request) {
	themes, err := h.themeService.List()
	if err != nil {
		respondServerError(w, r, "error.failed_list_themes", err)
		return
	}
	respondJSON(w, themes, http.StatusOK)
}
