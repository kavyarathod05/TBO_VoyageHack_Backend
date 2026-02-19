package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FlightBooking represents a flight booking for a specific event
type FlightBooking struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FlightID    uuid.UUID `gorm:"type:uuid;not null;index" json:"flight_id"`
	Flight      Flight    `gorm:"foreignKey:FlightID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"flight_details,omitempty"`
	EventID     uuid.UUID `gorm:"type:uuid;not null;index" json:"event_id"`
	Event       Event     `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	SeatsBooked int       `gorm:"not null" json:"seats_booked"`                    // Number of seats booked
	PriceLocked float64   `gorm:"type:decimal(10,2);not null" json:"price_locked"` // Price locked at booking time
	Status      string    `gorm:"size:20;default:'pending';index" json:"status"`   // 'pending', 'confirmed', 'cancelled'
	BookedBy    uuid.UUID `gorm:"type:uuid" json:"booked_by"`                      // User who made the booking
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BeforeSave validates the flight booking before saving
func (fb *FlightBooking) BeforeSave(tx *gorm.DB) error {
	// Validate status
	validStatuses := map[string]bool{
		"pending":   true,
		"confirmed": true,
		"cancelled": true,
	}
	if !validStatuses[fb.Status] {
		return gorm.ErrInvalidData
	}

	// Validate seats booked
	if fb.SeatsBooked <= 0 {
		return gorm.ErrInvalidData
	}

	return nil
}
