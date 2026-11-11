package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/obeyler/supplier-safety-form/internal/i18n"
	"github.com/obeyler/supplier-safety-form/internal/models"
)

type RegistrationService struct {
	userService   *UserService
	orgService    *OrganizationService
	sectorService *SectorService
}

func NewRegistrationService(userService *UserService, orgService *OrganizationService, sectorService *SectorService) *RegistrationService {
	return &RegistrationService{userService: userService, orgService: orgService, sectorService: sectorService}
}

// RegisterRequest is the public self-registration form: a company picks its
// name and sector, a first owner account is created (no password handled by
// this app — Keycloak sends its own "set your password" email, same pattern
// as pra2's opt-in flow).
type RegisterRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	OrgName    string `json:"org_name"`
	SectorCode string `json:"sector_code"`
	Language   string `json:"language"`
}

// RejectedError is a refusal attributable to the requester (bad input,
// email already taken…), as opposed to an internal failure — the HTTP layer
// answers 400 with Message as-is.
type RejectedError struct {
	Code    string
	Message string
}

func (e *RejectedError) Error() string { return e.Message }

func rejected(code, message string) error {
	return &RejectedError{Code: code, Message: message}
}

func (s *RegistrationService) Register(req RegisterRequest) (*models.Organization, error) {
	lang := i18n.ParseLang(req.Language)

	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.OrgName = strings.TrimSpace(req.OrgName)
	req.SectorCode = strings.TrimSpace(req.SectorCode)

	if req.FirstName == "" || req.LastName == "" || req.Email == "" {
		return nil, rejected("missing_fields", i18n.T(lang, "error.missing_fields"))
	}
	if !strings.Contains(req.Email, "@") {
		return nil, rejected("invalid_email", i18n.T(lang, "error.invalid_email"))
	}
	if req.OrgName == "" {
		return nil, rejected("org_name_required", i18n.T(lang, "error.org_name_required"))
	}
	if req.SectorCode == "" {
		return nil, rejected("sector_required", i18n.T(lang, "error.sector_required"))
	}

	sectors, err := s.sectorService.ListActive()
	if err != nil {
		return nil, fmt.Errorf("lecture des secteurs: %w", err)
	}
	validSector := false
	for _, sec := range sectors {
		if sec.Code == req.SectorCode {
			validSector = true
			break
		}
	}
	if !validSector {
		return nil, rejected("sector_required", i18n.T(lang, "error.sector_required"))
	}

	existing, err := s.userService.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("lecture du compte existant: %w", err)
	}
	if existing != nil {
		return nil, rejected("email_taken", i18n.T(lang, "error.email_taken"))
	}

	adminToken, err := s.getAdminToken()
	if err != nil {
		return nil, err
	}

	keycloakID, err := s.createPasswordlessKeycloakUser(adminToken, req.FirstName, req.LastName, req.Email, lang)
	if err != nil {
		if strings.Contains(err.Error(), "existe déjà") {
			return nil, rejected("email_taken", i18n.T(lang, "error.email_taken"))
		}
		return nil, err
	}

	user := &models.User{
		ID:        keycloakID,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}
	if err := s.userService.Create(user); err != nil {
		return nil, fmt.Errorf("création du compte: %w", err)
	}

	org, err := s.orgService.CreateWithOwner(req.OrgName, req.SectorCode, keycloakID)
	if err != nil {
		return nil, fmt.Errorf("création de l'organisation: %w", err)
	}

	// L'envoi du lien d'activation n'est pas bloquant : le compte et
	// l'organisation existent déjà, et un superadmin peut relancer l'action
	// depuis Keycloak. Échouer ici laisserait croire à l'utilisateur que son
	// inscription a échoué alors que tout est créé.
	if err := s.sendSetPasswordEmail(adminToken, keycloakID); err != nil {
		log.Printf("inscription: envoi du lien d'activation à %s échoué: %v", req.Email, err)
	}

	return org, nil
}

func (s *RegistrationService) keycloakInternalURL() string {
	u := os.Getenv("KEYCLOAK_INTERNAL_URL")
	return strings.TrimRight(u, "/")
}

func (s *RegistrationService) keycloakRealm() string {
	issuer := os.Getenv("KEYCLOAK_ISSUER_URL")
	parts := strings.Split(strings.TrimRight(issuer, "/"), "/")
	return parts[len(parts)-1]
}

func (s *RegistrationService) getAdminToken() (string, error) {
	tokenURL := s.keycloakInternalURL() + "/realms/master/protocol/openid-connect/token"
	adminUser := os.Getenv("KEYCLOAK_ADMIN_USER")
	if adminUser == "" {
		adminUser = "admin"
	}
	resp, err := http.PostForm(tokenURL, url.Values{
		"grant_type": {"password"},
		"client_id":  {"admin-cli"},
		"username":   {adminUser},
		"password":   {os.Getenv("KEYCLOAK_ADMIN_PASSWORD")},
	})
	if err != nil {
		return "", fmt.Errorf("impossible de contacter Keycloak: %w", err)
	}
	defer resp.Body.Close()
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil || tok.AccessToken == "" {
		return "", fmt.Errorf("authentification Keycloak admin échouée")
	}
	return tok.AccessToken, nil
}

// createPasswordlessKeycloakUser crée le compte sans identifiant, en exigeant
// que l'utilisateur pose son mot de passe et vérifie son adresse à la
// première connexion — c'est Keycloak qui envoie ce lien (sendSetPasswordEmail).
func (s *RegistrationService) createPasswordlessKeycloakUser(adminToken, firstName, lastName, email string, lang i18n.Lang) (string, error) {
	usersURL := fmt.Sprintf("%s/admin/realms/%s/users", s.keycloakInternalURL(), s.keycloakRealm())
	payload, _ := json.Marshal(map[string]interface{}{
		"username":        email,
		"email":           email,
		"firstName":       firstName,
		"lastName":        lastName,
		"enabled":         true,
		"emailVerified":   false,
		"attributes":      map[string][]string{"locale": {i18n.KeycloakLocale(lang)}},
		"requiredActions": []string{"UPDATE_PASSWORD", "VERIFY_EMAIL"},
	})
	req, _ := http.NewRequest("POST", usersURL, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erreur réseau Keycloak: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return "", fmt.Errorf("un compte existe déjà avec cet email")
	}
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("création compte échouée (%d): %s", resp.StatusCode, string(body))
	}

	location := resp.Header.Get("Location")
	parts := strings.Split(strings.TrimRight(location, "/"), "/")
	if len(parts) == 0 || parts[len(parts)-1] == "" {
		return "", fmt.Errorf("ID utilisateur absent de la réponse Keycloak")
	}
	id := parts[len(parts)-1]
	if _, err := uuid.Parse(id); err != nil {
		return "", fmt.Errorf("ID utilisateur Keycloak invalide: %q", id)
	}
	return id, nil
}

const setPasswordLifespan = 7 * 24 * 3600 // 7 jours

func (s *RegistrationService) sendSetPasswordEmail(adminToken, keycloakID string) error {
	frontendURL := os.Getenv("FRONTEND_URL")
	clientID := os.Getenv("KEYCLOAK_CLIENT_ID")

	actionsURL := fmt.Sprintf("%s/admin/realms/%s/users/%s/execute-actions-email",
		s.keycloakInternalURL(), s.keycloakRealm(), keycloakID)
	params := url.Values{"lifespan": {fmt.Sprintf("%d", setPasswordLifespan)}}
	if frontendURL != "" && clientID != "" {
		params.Set("client_id", clientID)
		params.Set("redirect_uri", frontendURL)
	}

	payload, _ := json.Marshal([]string{"UPDATE_PASSWORD", "VERIFY_EMAIL"})
	req, _ := http.NewRequest("PUT", actionsURL+"?"+params.Encode(), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erreur réseau Keycloak: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("execute-actions-email a répondu %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
