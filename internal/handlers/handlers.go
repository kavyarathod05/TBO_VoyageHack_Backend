package handlers

import (
	"github.com/akashtripathi12/TBO_Backend/internal/config"
	"gorm.io/gorm"
)

// Repository holds the application configuration and database store
type Repository struct {
	App *config.Config
	DB  *gorm.DB
}

// NewRepository creates a new instance of the repository
func NewRepository(app *config.Config, db *gorm.DB) *Repository {
	return &Repository{
		App: app,
		DB:  db,
	}
}
