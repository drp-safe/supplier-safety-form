package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/obeyler/supplier-safety-form/internal/i18n"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// respondError sends a JSON error response with a translated message.
func respondError(w http.ResponseWriter, r *http.Request, code string, status int) {
	lang := i18n.LangFromRequest(r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: code, Message: i18n.T(lang, code)}) //nolint:errcheck
}

// respondServerError logs the error and sends a generic 500 response. Use
// this for all unexpected internal errors (DB failures, service errors).
func respondServerError(w http.ResponseWriter, r *http.Request, code string, err error) {
	log.Printf("server error [%s]: %v", code, err)
	respondError(w, r, code, http.StatusInternalServerError)
}

// respondSuccess sends a JSON success response with a translated message.
func respondSuccess(w http.ResponseWriter, r *http.Request, msgKey string, status int) {
	lang := i18n.LangFromRequest(r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"message": i18n.T(lang, msgKey)}) //nolint:errcheck
}

// respondJSON sends a JSON response with the given data.
func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}
