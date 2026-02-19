package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Flight represents a global flight available for booking
type Flight struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FlightNumber   string    `gorm:"size:20;not null;index" json:"flight_number"`   // e.g., "AI 234"
	AirlineName    string    `gorm:"size:100;not null" json:"airline_name"`         // e.g., "Air India"
	DepartureTime  time.Time `gorm:"not null;index" json:"departure_time"`          // Departure datetime
	ArrivalTime    time.Time `gorm:"not null" json:"arrival_time"`                  // Arrival datetime
	DepartureCode  string    `gorm:"size:3;not null;index" json:"departure_code"`   // Airport code (e.g., "BOM")
	ArrivalCode    string    `gorm:"size:3;not null;index" json:"arrival_code"`     // Airport code (e.g., "DEL")
	TotalSeats     int       `gorm:"not null;default:180" json:"total_seats"`       // Total capacity
	AvailableSeats int       `gorm:"not null;default:180" json:"available_seats"`   // Remaining seats
	BasePrice      float64   `gorm:"type:decimal(10,2);not null" json:"base_price"` // Base price per seat
	Status         string    `gorm:"size:20;default:'active';index" json:"status"`  // 'active', 'cancelled'
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// BeforeSave validates the flight before saving
func (f *Flight) BeforeSave(tx *gorm.DB) error {
	// Validate status
	validStatuses := map[string]bool{
		"active":    true,
		"cancelled": true,
	}
	if !validStatuses[f.Status] {
		return gorm.ErrInvalidData
	}

	// Validate times
	if !f.ArrivalTime.After(f.DepartureTime) {
		return gorm.ErrInvalidData
	}

	// Validate seats
	if f.AvailableSeats > f.TotalSeats {
		return gorm.ErrInvalidData
	}

	return nil
}
