package database

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func TestIsUniqueViolation(t *testing.T) {
	pgErr := &pgconn.PgError{Code: uniqueViolation}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "postgres unique", err: pgErr, want: true},
		{name: "wrapped postgres unique", err: fmt.Errorf("insert user: %w", pgErr), want: true},
		{name: "foreign key", err: &pgconn.PgError{Code: "23503"}, want: false},
		{name: "gorm duplicated key", err: gorm.ErrDuplicatedKey, want: true},
		{name: "wrapped gorm duplicated key", err: fmt.Errorf("insert user: %w", gorm.ErrDuplicatedKey), want: true},
		{name: "other", err: errors.New("connection reset"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUniqueViolation(tt.err); got != tt.want {
				t.Fatalf("IsUniqueViolation() = %v, want %v", got, tt.want)
			}
		})
	}
}
