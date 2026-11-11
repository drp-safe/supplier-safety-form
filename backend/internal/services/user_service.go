package services

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/obeyler/supplier-safety-form/internal/models"
)

type UserService struct {
	db *sqlx.DB
}

func NewUserService(db *sqlx.DB) *UserService {
	return &UserService{db: db}
}

// GetByKeycloakSub implements middleware.UserLookup. The JWT "sub" claim is
// exactly the user's id (see models.User), set at creation from the Keycloak
// admin API's own response — no separate mapping column.
func (s *UserService) GetByKeycloakSub(sub string) (*models.User, error) {
	var user models.User
	err := s.db.Get(&user, `SELECT * FROM users WHERE id = $1`, sub)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := s.db.Get(&user, `SELECT * FROM users WHERE email = $1`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) Create(user *models.User) error {
	_, err := s.db.Exec(`
		INSERT INTO users (id, email, first_name, last_name, is_superadmin)
		VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.Email, user.FirstName, user.LastName, user.IsSuperAdmin)
	return err
}
