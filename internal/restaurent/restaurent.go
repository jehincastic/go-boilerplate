package restaurent

import (
	"github.com/google/uuid"
)

type Restaurent struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `json:"name"`
	StartTime string    `json:"startTime"`
	EndTime   string    `json:"endTime"`
}

// TableName maps User to the users table created by SQL migrations.
func (Restaurent) TableName() string {
	return "restaurents"
}
