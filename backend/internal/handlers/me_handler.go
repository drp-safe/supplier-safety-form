package handlers

import (
	"net/http"

	"github.com/obeyler/supplier-safety-form/internal/middleware"
	"github.com/obeyler/supplier-safety-form/internal/models"
	"github.com/obeyler/supplier-safety-form/internal/services"
)

type MeHandler struct {
	orgService *services.OrganizationService
}

func NewMeHandler(orgService *services.OrganizationService) *MeHandler {
	return &MeHandler{orgService: orgService}
}

type meResponse struct {
	*models.User
	Organization *models.Organization `json:"organization"`
}

// GetMe GET /api/me
func (h *MeHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(middleware.CtxUser).(*models.User)
	org, err := h.orgService.GetForUser(user.ID)
	if err != nil {
		respondServerError(w, r, "error.internal_error", err)
		return
	}
	respondJSON(w, meResponse{User: user, Organization: org}, http.StatusOK)
}
