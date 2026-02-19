package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Transfer represents a global cab/transfer option available for booking
type Transfer struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CabType         string    `gorm:"size:20;not null;index" json:"cab_type"`                // 'hatchback', 'sedan', 'suv'
	CarModel        string    `gorm:"size:100;not null" json:"car_model"`                    // e.g., "Toyota Innova"
	TotalCount      int       `gorm:"not null;default:10" json:"total_count"`                // Total cabs available
	AvailableCount  int       `gorm:"not null;default:10" json:"available_count"`            // Remaining cabs
	BasePricePerCab float64   `gorm:"type:decimal(10,2);not null" json:"base_price_per_cab"` // Base price per cab
	Status          string    `gorm:"size:20;default:'active';index" json:"status"`          // 'active', 'inactive'
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// BeforeSave validates the transfer before saving
func (t *Transfer) BeforeSave(tx *gorm.DB) error {
	// Validate cab type
	validCabTypes := map[string]bool{
		"hatchback": true,
		"sedan":     true,
		"suv":       true,
	}
	if !validCabTypes[t.CabType] {
		return gorm.ErrInvalidData
	}

	// Validate status
	validStatuses := map[string]bool{
		"active":   true,
		"inactive": true,
	}
	if !validStatuses[t.Status] {
		return gorm.ErrInvalidData
	}

	// Validate count
	if t.AvailableCount > t.TotalCount {
		return gorm.ErrInvalidData
	}

	return nil
}

