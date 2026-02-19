package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransferBooking represents a cab/transfer booking for a specific event
type TransferBooking struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TransferID     uuid.UUID `gorm:"type:uuid;not null;index" json:"transfer_id"`
	Transfer       Transfer  `gorm:"foreignKey:TransferID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"transfer_details,omitempty"`
	EventID        uuid.UUID `gorm:"type:uuid;not null;index" json:"event_id"`
	Event          Event     `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	CabsBooked     int       `gorm:"not null" json:"cabs_booked"`                     // Number of cabs booked
	PriceLocked    float64   `gorm:"type:decimal(10,2);not null" json:"price_locked"` // Price locked at booking time
	PickupLocation string    `gorm:"size:200" json:"pickup_location"`                 // Pickup address
	DropLocation   string    `gorm:"size:200" json:"drop_location"`                   // Drop address
	PickupTime     time.Time `json:"pickup_time"`                                     // Scheduled pickup time
	Status         string    `gorm:"size:20;default:'pending';index" json:"status"`   // 'pending', 'confirmed', 'completed', 'cancelled'
	BookedBy       uuid.UUID `gorm:"type:uuid" json:"booked_by"`                      // User who made the booking
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// BeforeSave validates the transfer booking before saving
func (tb *TransferBooking) BeforeSave(tx *gorm.DB) error {
	// Validate status
	validStatuses := map[string]bool{
		"pending":   true,
		"confirmed": true,
		"completed": true,
		"cancelled": true,
	}
	if !validStatuses[tb.Status] {
		return gorm.ErrInvalidData
	}

	// Validate cabs booked
	if tb.CabsBooked <= 0 {
		return gorm.ErrInvalidData
	}

	return nil
}
