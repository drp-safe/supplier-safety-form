package models

import "time"

// Sector is a reference entry in the (admin-editable, not hardcoded) list of
// business sectors an organization picks from at registration.
type Sector struct {
	Code         string `db:"code" json:"code"`
	Label        string `db:"label" json:"label"`
	DisplayOrder int    `db:"display_order" json:"display_order"`
	Active       bool   `db:"active" json:"active"`
}

// Theme classifies questions in the (future) public question library.
type Theme struct {
	Code         string `db:"code" json:"code"`
	Label        string `db:"label" json:"label"`
	DisplayOrder int    `db:"display_order" json:"display_order"`
}

// Organization is a registered company — it can both send forms to its own
// suppliers and answer forms sent to it by other organizations (single
// account model, no buyer/supplier distinction).
type Organization struct {
	ID         string    `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	SectorCode string    `db:"sector_code" json:"sector_code"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// User mirrors a Keycloak account. ID is the Keycloak subject's own UUID
// (exactly the JWT "sub" claim) — there is no separate local identifier to
// keep in sync.
type User struct {
	ID           string    `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	FirstName    string    `db:"first_name" json:"first_name"`
	LastName     string    `db:"last_name" json:"last_name"`
	IsSuperAdmin bool      `db:"is_superadmin" json:"is_superadmin"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// OrganizationMember links a user to an organization with an applicative
// role — "owner" (set once at registration) or "member" (reserved for a
// later stage: inviting teammates is not part of Étape 1).
type OrganizationMember struct {
	ID             string    `db:"id" json:"id"`
	OrganizationID string    `db:"organization_id" json:"organization_id"`
	UserID         string    `db:"user_id" json:"user_id"`
	Role           string    `db:"role" json:"role"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

// EmailConfig is the single-row SMTP configuration, editable from the
// Administration section (superadmin only). SMTPPassword is encrypted at
// rest and never serialized back to the client.
type EmailConfig struct {
	ID             string    `db:"id" json:"id"`
	SMTPHost       string    `db:"smtp_host" json:"smtp_host"`
	SMTPPort       int       `db:"smtp_port" json:"smtp_port"`
	SMTPUsername   string    `db:"smtp_username" json:"smtp_username"`
	SMTPPassword   string    `db:"smtp_password" json:"-"`
	SMTPEncryption string    `db:"smtp_encryption" json:"smtp_encryption"`
	FromAddress    string    `db:"from_address" json:"from_address"`
	FromName       string    `db:"from_name" json:"from_name"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// EmailConfigUpdate is the request body for PUT /api/admin/email-config.
type EmailConfigUpdate struct {
	SMTPHost       string `json:"smtp_host"`
	SMTPPort       int    `json:"smtp_port"`
	SMTPUsername   string `json:"smtp_username"`
	SMTPPassword   string `json:"smtp_password"`
	SMTPEncryption string `json:"smtp_encryption"`
	FromAddress    string `json:"from_address"`
	FromName       string `json:"from_name"`
}
