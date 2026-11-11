package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/obeyler/supplier-safety-form/internal/i18n"
	"github.com/obeyler/supplier-safety-form/internal/services"
)

type RegistrationHandler struct {
	service *services.RegistrationService
}

func NewRegistrationHandler(service *services.RegistrationService) *RegistrationHandler {
	return &RegistrationHandler{service: service}
}

// Register POST /api/register (public, unauthenticated).
func (h *RegistrationHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req services.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "error.invalid_request", http.StatusBadRequest)
		return
	}

	org, err := h.service.Register(req)
	if err != nil {
		if rejected, ok := err.(*services.RejectedError); ok {
			lang := i18n.ParseLang(req.Language)
			respondJSON(w, ErrorResponse{Error: rejected.Code, Message: i18n.T(lang, "error."+rejected.Code)}, http.StatusBadRequest)
			return
		}
		respondServerError(w, r, "error.registration_failed", err)
		return
	}

	respondJSON(w, map[string]interface{}{
		"organization_id": org.ID,
		"message":         i18n.T(i18n.ParseLang(req.Language), "success.registration_completed"),
	}, http.StatusCreated)
}
