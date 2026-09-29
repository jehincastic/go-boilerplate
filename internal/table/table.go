package table

import (
	"github.com/google/uuid"
)

type Table struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Restaurent uuid.UUID `json:"restaurentId"`
	Name       string    `json:"name"`
	Capacity   int       `json:"capacity"`
}

// TableName maps User to the users table created by SQL migrations.
func (Table) TableName() string {
	return "tables"
}
