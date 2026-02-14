package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// User represents a registered user (Head Guest or Agent)
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ClerkID      string    `gorm:"uniqueIndex;not null"` // External Auth ID
	Email        string    `gorm:"uniqueIndex;not null"`
	Role         string    `gorm:"default:'head_guest'"` // 'agent' or 'head_guest'
	Name         string
	Phone        string
	AgentProfile AgentProfile `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type AgentProfile struct {
	UserID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	AgencyName    string
	AgencyCode    string
	Location      string
	BusinessPhone string
}

type Guest struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	EventID       uuid.UUID `gorm:"type:uuid;index;not null"`
	Name          string    `gorm:"not null"`
	Age           int
	Type          string `gorm:"default:'adult'"` // 'adult' or 'child'
	ArrivalDate   time.Time
	DepartureDate time.Time
	Phone         string
	Email         string
}

func (g *Guest) BeforeSave(tx *gorm.DB) (err error) {
	if g.Age >= 12 {
		g.Type = "Adult"
	} else if g.Age < 12 {
		g.Type = "Child"
	}
	return
}

type Event struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AgentID        uuid.UUID      `gorm:"type:uuid;index"`
	HeadGuestID    uuid.UUID      `gorm:"type:uuid;index"`
	HotelID        uuid.UUID      `gorm:"index"`
	RoomsInventory datatypes.JSON `gorm:"type:jsonb"` // Stores the [type, count] array
	Status         string         `gorm:"default:'draft'"`
	StartDate      time.Time
	EndDate        time.Time
	Location       string
}
