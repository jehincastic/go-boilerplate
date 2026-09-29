package slot

import (
	"github.com/google/uuid"
)

type Slot struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Restaurent uuid.UUID `json:"restaurentId"`
	StartTime  string    `json:"startTime"`
	EndTime    string    `json:"endTime"`
}

// TableName maps User to the users table created by SQL migrations.
func (Slot) TableName() string {
	return "slots"
}
