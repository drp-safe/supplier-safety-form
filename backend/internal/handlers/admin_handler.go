package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/obeyler/supplier-safety-form/internal/email"
	"github.com/obeyler/supplier-safety-form/internal/models"
	"github.com/obeyler/supplier-safety-form/internal/services"
)

// AdminHandler serves the Administration section, gated by
// middleware.RequireSuperAdmin at the router level — never called directly
// by an ordinary organization member.
type AdminHandler struct {
	adminService  *services.AdminService
	sectorService *services.SectorService
}

func NewAdminHandler(adminService *services.AdminService, sectorService *services.SectorService) *AdminHandler {
	return &AdminHandler{adminService: adminService, sectorService: sectorService}
}

// GetEmailConfig GET /api/admin/email-config
func (h *AdminHandler) GetEmailConfig(w http.ResponseWriter, r *http.Request) {
	config, err := h.adminService.GetEmailConfig()
	if err != nil {
		respondServerError(w, r, "error.failed_get_email_config", err)
		return
	}
	respondJSON(w, config, http.StatusOK)
}

// UpdateEmailConfig PUT /api/admin/email-config
func (h *AdminHandler) UpdateEmailConfig(w http.ResponseWriter, r *http.Request) {
	var update models.EmailConfigUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		respondError(w, r, "error.invalid_request", http.StatusBadRequest)
		return
	}
	if update.SMTPHost == "" || update.FromAddress == "" {
		respondError(w, r, "error.smtp_host_required", http.StatusBadRequest)
		return
	}
	if err := h.adminService.UpdateEmailConfig(&update); err != nil {
		respondServerError(w, r, "error.failed_update_email_config", err)
		return
	}
	respondSuccess(w, r, "success.email_config_updated", http.StatusOK)
}

// TestEmailConfig POST /api/admin/email-config/test
func (h *AdminHandler) TestEmailConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TestEmail string `json:"test_email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TestEmail == "" {
		respondError(w, r, "error.test_email_required", http.StatusBadRequest)
		return
	}

	config, err := h.adminService.GetEmailConfig()
	if err != nil {
		respondServerError(w, r, "error.failed_load_email_config", err)
		return
	}

	mailer := email.NewEmailService(config)
	if err := mailer.SendEmail(req.TestEmail, "Supplier Safety Form — test SMTP", "<p>Configuration SMTP valide.</p>"); err != nil {
		respondServerError(w, r, "error.test_email_failed", err)
		return
	}
	respondSuccess(w, r, "success.test_email_sent", http.StatusOK)
}

// GetOAuthConfig GET /api/admin/oauth-config — lecture seule.
func (h *AdminHandler) GetOAuthConfig(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, h.adminService.GetOAuthConfig(), http.StatusOK)
}

// GetDBInfo GET /api/admin/db-info — lecture seule, jamais le mot de passe.
func (h *AdminHandler) GetDBInfo(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, h.adminService.GetDBInfo(), http.StatusOK)
}

// ListAllSectors GET /api/admin/sectors (y compris les secteurs désactivés).
func (h *AdminHandler) ListAllSectors(w http.ResponseWriter, r *http.Request) {
	sectors, err := h.sectorService.ListAll()
	if err != nil {
		respondServerError(w, r, "error.failed_list_sectors", err)
		return
	}
	respondJSON(w, sectors, http.StatusOK)
}

type sectorRequest struct {
	Code         string `json:"code"`
	Label        string `json:"label"`
	DisplayOrder int    `json:"display_order"`
	Active       bool   `json:"active"`
}

// CreateSector POST /api/admin/sectors
func (h *AdminHandler) CreateSector(w http.ResponseWriter, r *http.Request) {
	var req sectorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		respondError(w, r, "error.invalid_request", http.StatusBadRequest)
		return
	}
	if req.Label == "" {
		respondError(w, r, "error.sector_label_required", http.StatusBadRequest)
		return
	}
	if _, err := h.sectorService.Create(req.Code, req.Label, req.DisplayOrder); err != nil {
		respondServerError(w, r, "error.failed_create_sector", err)
		return
	}
	respondSuccess(w, r, "success.sector_created", http.StatusCreated)
}

// UpdateSector PUT /api/admin/sectors/{code}
func (h *AdminHandler) UpdateSector(w http.ResponseWriter, r *http.Request) {
	code := mux.Vars(r)["code"]
	var req sectorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "error.invalid_request", http.StatusBadRequest)
		return
	}
	if req.Label == "" {
		respondError(w, r, "error.sector_label_required", http.StatusBadRequest)
		return
	}
	if err := h.sectorService.Update(code, req.Label, req.DisplayOrder, req.Active); err != nil {
		respondServerError(w, r, "error.failed_update_sector", err)
		return
	}
	respondSuccess(w, r, "success.sector_updated", http.StatusOK)
}

// DeleteSector DELETE /api/admin/sectors/{code} — désactivation, pas de
// suppression physique (des organisations existantes y réfèrent déjà).
func (h *AdminHandler) DeleteSector(w http.ResponseWriter, r *http.Request) {
	code := mux.Vars(r)["code"]
	if err := h.sectorService.Delete(code); err != nil {
		respondServerError(w, r, "error.failed_delete_sector", err)
		return
	}
	respondSuccess(w, r, "success.sector_deleted", http.StatusOK)
}
