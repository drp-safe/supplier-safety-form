package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

// KeycloakConfig holds Keycloak connection settings.
type KeycloakConfig struct {
	IssuerURL   string // e.g. "https://auth-dev.drp-safe.com/realms/supplier-safety-form"
	ClientID    string // e.g. "supplier-safety-form-frontend"
	InternalURL string // optional in-cluster URL, e.g. "http://keycloak.keycloak-dev.svc.cluster.local:8080"
}

// KeycloakVerifier validates Keycloak-issued JWT access tokens.
type KeycloakVerifier struct {
	verifier       *oidc.IDTokenVerifier
	clientID       string
	trustedClients map[string]struct{}
}

// KeycloakClaims holds the claims extracted from a Keycloak token.
type KeycloakClaims struct {
	Subject    string
	Email      string
	GivenName  string
	FamilyName string
	Roles      []string
}

// internalRewriteTransport rewrites requests from the external Keycloak URL to
// the internal K8s service URL, solving hairpin NAT issues.
type internalRewriteTransport struct {
	base        http.RoundTripper
	externalURL *url.URL
	internalURL *url.URL
}

func (t *internalRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == t.externalURL.Host {
		req = req.Clone(req.Context())
		originalScheme := req.URL.Scheme
		req.URL.Scheme = t.internalURL.Scheme
		req.URL.Host = t.internalURL.Host
		req.Host = t.externalURL.Host
		req.Header.Set("X-Forwarded-Proto", originalScheme)
		req.Header.Set("X-Forwarded-Host", t.externalURL.Host)
	}
	return t.base.RoundTrip(req)
}

// NewKeycloakVerifier creates a verifier that fetches JWKS from Keycloak.
func NewKeycloakVerifier(cfg KeycloakConfig) (*KeycloakVerifier, error) {
	ctx := context.Background()

	if cfg.InternalURL != "" {
		extURL, err := url.Parse(cfg.IssuerURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse issuer URL: %w", err)
		}
		intURL, err := url.Parse(cfg.InternalURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse internal URL: %w", err)
		}
		httpClient := &http.Client{Transport: &internalRewriteTransport{
			base:        http.DefaultTransport,
			externalURL: extURL,
			internalURL: intURL,
		}}
		ctx = oidc.ClientContext(ctx, httpClient)
	}

	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get OIDC provider: %w", err)
	}

	// Keycloak access tokens have aud:"account", not the client ID; verify azp manually.
	verifier := provider.Verifier(&oidc.Config{SkipClientIDCheck: true})

	trusted := map[string]struct{}{cfg.ClientID: {}}
	if extra := os.Getenv("KEYCLOAK_TRUSTED_AZPS"); extra != "" {
		for _, id := range splitTrimmed(extra) {
			trusted[id] = struct{}{}
		}
	}

	return &KeycloakVerifier{verifier: verifier, clientID: cfg.ClientID, trustedClients: trusted}, nil
}

func splitTrimmed(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// VerifyToken validates a Keycloak access token and extracts claims.
func (kv *KeycloakVerifier) VerifyToken(ctx context.Context, rawToken string) (*KeycloakClaims, error) {
	idToken, err := kv.verifier.Verify(ctx, rawToken)
	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	var all map[string]interface{}
	if err := idToken.Claims(&all); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	if len(kv.trustedClients) > 0 {
		azp, _ := all["azp"].(string)
		if _, ok := kv.trustedClients[azp]; !ok {
			return nil, fmt.Errorf("token azp %q is not a trusted client", azp)
		}
	}

	claims := &KeycloakClaims{Subject: idToken.Subject}
	if email, ok := all["email"].(string); ok {
		claims.Email = email
	}
	if name, ok := all["given_name"].(string); ok {
		claims.GivenName = name
	}
	if name, ok := all["family_name"].(string); ok {
		claims.FamilyName = name
	}
	if realmAccess, ok := all["realm_access"].(map[string]interface{}); ok {
		if roles, ok := realmAccess["roles"].([]interface{}); ok {
			for _, r := range roles {
				if role, ok := r.(string); ok {
					claims.Roles = append(claims.Roles, role)
				}
			}
		}
	}
	if resourceAccess, ok := all["resource_access"].(map[string]interface{}); ok {
		if clientAccess, ok := resourceAccess[kv.clientID].(map[string]interface{}); ok {
			if roles, ok := clientAccess["roles"].([]interface{}); ok {
				for _, r := range roles {
					if role, ok := r.(string); ok {
						claims.Roles = append(claims.Roles, role)
					}
				}
			}
		}
	}

	return claims, nil
}
