package user

import (
	"context"
	"errors"

	"github.com/jehincastic/go-boilerplate/internal/platform/database"
	"github.com/jehincastic/go-boilerplate/internal/platform/errs"
	"gorm.io/gorm"
)

// Repository stores users with GORM.
type Repository struct {
	db *gorm.DB
}

// NewRepository returns a user repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Insert stores a user. Postgres assigns the uuidv7 id.
func (r *Repository) Insert(ctx context.Context, user *User) error {
	ctx, cancel := database.WithTimeout(ctx)
	defer cancel()

	err := database.FromContext(ctx, r.db).Raw(
		`INSERT INTO users (name, email, password_hash)
		 VALUES (?, ?, ?)
		 RETURNING id, created_at, name, email, password_hash`,
		user.Name, user.Email, user.PasswordHash,
	).Scan(user).Error
	if database.IsUniqueViolation(err) {
		return errs.ErrDuplicateEmail
	}
	return err
}

// GetByEmail loads the user with the given email address.
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := database.WithTimeout(ctx)
	defer cancel()

	var user User
	err := database.FromContext(ctx, r.db).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
