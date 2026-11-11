package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/obeyler/supplier-safety-form/internal/auth"
	"github.com/obeyler/supplier-safety-form/internal/models"
)

type ctxKey string

const (
	CtxUserEmail ctxKey = "user_email"
	CtxUserName  ctxKey = "user_name"
	CtxUserSub   ctxKey = "user_sub"
	CtxUser      ctxKey = "user"
)

func writeError(w http.ResponseWriter, code string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": code}) //nolint:errcheck
}

// KeycloakAuth validates the Bearer access token and injects the identity into
// the request context.
func KeycloakAuth(verifier *auth.KeycloakVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeError(w, "missing_or_invalid_auth_header", http.StatusUnauthorized)
				return
			}

			claims, err := verifier.VerifyToken(r.Context(), parts[1])
			if err != nil {
				log.Printf("token verification failed: %v", err)
				writeError(w, "invalid_token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), CtxUserEmail, claims.Email)
			ctx = context.WithValue(ctx, CtxUserName, strings.TrimSpace(claims.GivenName+" "+claims.FamilyName))
			ctx = context.WithValue(ctx, CtxUserSub, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserLookup is the minimal dependency LoadUser needs from UserService —
// kept as an interface here so middleware does not import internal/services
// (which itself may import middleware's context keys).
type UserLookup interface {
	GetByKeycloakSub(sub string) (*models.User, error)
}

// LoadUser resolves the authenticated Keycloak subject to the application's
// own user row (created at registration) and injects it into the context.
// Must run after KeycloakAuth. A token with no matching row (should not
// happen: registration creates the DB row and the Keycloak account in the
// same call) is rejected rather than silently treated as anonymous.
func LoadUser(lookup UserLookup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			sub, _ := r.Context().Value(CtxUserSub).(string)
			user, err := lookup.GetByKeycloakSub(sub)
			if err != nil {
				log.Printf("load user failed: %v", err)
				writeError(w, "internal_error", http.StatusInternalServerError)
				return
			}
			if user == nil {
				writeError(w, "user_not_found", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), CtxUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireSuperAdmin only lets through requests whose loaded user (see
// LoadUser) carries the platform-wide is_superadmin flag — independent of any
// organization_members role, same spirit as adminpra-backend's own
// RequireSuperAdmin but backed by a DB column rather than a Keycloak realm
// role (decided with the user: no realm-role management for this app).
func RequireSuperAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(CtxUser).(*models.User)
		if user == nil || !user.IsSuperAdmin {
			writeError(w, "insufficient_permissions", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
