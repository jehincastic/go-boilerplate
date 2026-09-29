package booking

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TableID         uuid.UUID `json:"tableId"`
	RestaurentID    uuid.UUID `json:"restaurentId"`
	SlotID          uuid.UUID `json:"slotId"`
	BookingCapacity int       `json:"bookingCapacity"`
	BookedOn        time.Time `json:"bookedOn"`
}

// TableName maps User to the users table created by SQL migrations.
func (Booking) TableName() string {
	return "bookings"
}
