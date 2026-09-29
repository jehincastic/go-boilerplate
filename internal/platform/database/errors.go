package database

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// uniqueViolation is the Postgres SQLSTATE for a unique-constraint failure.
const uniqueViolation = "23505"

// IsUniqueViolation reports whether err is a unique-constraint failure.
// It matches a Postgres 23505 error and GORM's translated duplicate-key error.
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolation
}
