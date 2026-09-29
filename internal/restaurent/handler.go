package restaurent

import (
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jehincastic/go-boilerplate/internal/booking"
	"github.com/jehincastic/go-boilerplate/internal/platform/httpserver"
	"github.com/jehincastic/go-boilerplate/internal/slot"
	"github.com/jehincastic/go-boilerplate/internal/table"
)

type ServiceAPI interface{}

type Cache struct {
	mu    sync.Mutex
	locks map[string]string
}

func (c *Cache) getLock(key string, userId string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.locks[key]
	if !ok {
		c.locks[key] = userId
		return true
	}
	return false
}

func (c *Cache) releaseLock(key string, userId string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	existingUser, ok := c.locks[key]
	if !ok {
		return false
	}
	if existingUser == userId {
		delete(c.locks, key)
		return true
	}
	return false
}

type Handler struct {
	svc ServiceAPI
	log *slog.Logger
}

type AvaliabilitySlot struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	StartTime          string    `json:"startTime"`
	EndTime            string    `json:"endTime"`
	AvailabileCapacity int       `json:"availableCapacity"`
}

type Availability struct {
	ID    uuid.UUID          `gorm:"type:uuid;primaryKey" json:"id"`
	Name  string             `json:"name"`
	Slots []AvaliabilitySlot `json:"slot"`
}

// NewHandler returns a user HTTP handler.
func NewHandler(log *slog.Logger) *Handler {
	return &Handler{log: log}
}

func isTimeWithinRange(lowerLimit, upperLimit, timeToCheck string) bool {
	return true
}

func (h *Handler) GetAvailableSlots(w http.ResponseWriter, r *http.Request) {
	restaurent := Restaurent{
		ID:        uuid.New(),
		Name:      "ABC",
		StartTime: "10:00",
		EndTime:   "22:00",
	}
	slots := []slot.Slot{
		{
			ID:         uuid.New(),
			StartTime:  "10:00",
			EndTime:    "11:00",
			Restaurent: restaurent.ID,
		},
		{
			ID:         uuid.New(),
			StartTime:  "11:00",
			EndTime:    "12:00",
			Restaurent: restaurent.ID,
		},
	}
	tables := []table.Table{
		{
			ID:         uuid.New(),
			Restaurent: restaurent.ID,
			Name:       "T1",
			Capacity:   5,
		},
		{
			ID:         uuid.New(),
			Restaurent: restaurent.ID,
			Name:       "T2",
			Capacity:   15,
		},
		{
			ID:         uuid.New(),
			Restaurent: restaurent.ID,
			Name:       "T3",
			Capacity:   7,
		},
	}
	bookings := []booking.Booking{
		{
			ID:              uuid.New(),
			TableID:         tables[0].ID,
			RestaurentID:    restaurent.ID,
			SlotID:          slots[1].ID,
			BookingCapacity: 4,
			BookedOn:        time.Now(),
		},
		{
			ID:              uuid.New(),
			TableID:         tables[2].ID,
			RestaurentID:    restaurent.ID,
			SlotID:          slots[1].ID,
			BookingCapacity: 7,
			BookedOn:        time.Now(),
		},
	}
	result := []Availability{}

	for _, tab := range tables {
		availableSlots := []AvaliabilitySlot{}
		for _, slt := range slots {
			isSlotTimingValid := isTimeWithinRange(restaurent.StartTime, restaurent.StartTime, slt.StartTime) && isTimeWithinRange(restaurent.StartTime, restaurent.StartTime, slt.EndTime)
			if isSlotTimingValid {
				availableCapacity := tab.Capacity
				for _, bk := range bookings {
					if bk.SlotID == slt.ID && bk.TableID == tab.ID {
						availableCapacity = availableCapacity - bk.BookingCapacity
					}
				}
				if availableCapacity > 0 {
					availableSlots = append(availableSlots, AvaliabilitySlot{
						ID:                 slt.ID,
						StartTime:          slt.StartTime,
						EndTime:            slt.EndTime,
						AvailabileCapacity: availableCapacity,
					})
				}
			}
		}
		result = append(result, Availability{
			ID:    tab.ID,
			Name:  tab.Name,
			Slots: availableSlots,
		})
	}
	if err := httpserver.WriteJSON(w, http.StatusOK, httpserver.Envelope{"avaiableTables": result}, nil); err != nil {
		httpserver.ServerError(w, r, h.log, err)
	}
}

var cache = Cache{locks: map[string]string{}}

func (h *Handler) BookSlot(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TableId    uuid.UUID `json:"tableId"`
		SlotTime   string    `json:"slotTime"`
		Capacity   int       `json:"capacity"`
		BookingFor time.Time `json:"bookingFor"`
	}
	if err := httpserver.ReadJSON(w, r, &input); err != nil {
		httpserver.BadRequest(w, r, err)
		return
	}
	// var restaurent Restaurent
	// var slot slot.Slot
	// var table table.Table
	// var bookings []booking.Booking
	restaurent := Restaurent{
		ID:        uuid.New(),
		Name:      "ABC",
		StartTime: "10:00",
		EndTime:   "22:00",
	}
	slot := slot.Slot{
		ID:         uuid.New(),
		StartTime:  "10:00",
		EndTime:    "11:00",
		Restaurent: restaurent.ID,
	}
	table := table.Table{
		ID:         input.TableId,
		Restaurent: restaurent.ID,
		Name:       "T1",
		Capacity:   5,
	}
	bookings := []booking.Booking{
		{
			ID:              uuid.New(),
			TableID:         table.ID,
			RestaurentID:    restaurent.ID,
			SlotID:          slot.ID,
			BookingCapacity: 4,
			BookedOn:        time.Now(),
		},
	}
	key := input.TableId.String() + "restaurent.ID.String()" + "slot.ID.String()"
	isLockAquiredByMe := cache.getLock(key, "user-id")
	if !isLockAquiredByMe {
		httpserver.BadRequest(w, r, errors.New("unable to get lock"))
		return
	}
	defer cache.releaseLock(key, "user-id")
	time.Sleep(time.Second * 10)
	availablCapacity := table.Capacity
	for _, bk := range bookings {
		availablCapacity = availablCapacity - bk.BookingCapacity
	}

	if availablCapacity < input.Capacity {
		httpserver.BadRequest(w, r, errors.New("max capacity"))
		return
	}

	newBook := booking.Booking{
		ID:              uuid.New(),
		TableID:         input.TableId,
		RestaurentID:    restaurent.ID,
		SlotID:          slot.ID,
		BookingCapacity: input.Capacity,
		BookedOn:        input.BookingFor,
	}
	if err := httpserver.WriteJSON(w, http.StatusOK, httpserver.Envelope{"booking": newBook}, nil); err != nil {
		httpserver.ServerError(w, r, h.log, err)
	}
}
