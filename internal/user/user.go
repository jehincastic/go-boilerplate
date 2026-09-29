package user

import (
	"time"

	"github.com/google/uuid"
)

// User is an account. The id is assigned by Postgres with uuidv7().
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	Name         string    `json:"name"`
	Email        string    `gorm:"type:citext" json:"email"`
	PasswordHash string    `gorm:"column:password_hash" json:"-"`
}

// TableName maps User to the users table created by SQL migrations.
func (User) TableName() string {
	return "users"
}
