package handlers

import (
	"github.com/akashtripathi12/TBO_Backend/internal/config"
)

// Repository holds the application configuration and database store
type Repository struct {
	App *config.Config
}

// NewRepository creates a new instance of the repository
func NewRepository(app *config.Config) *Repository {
	return &Repository{
		App: app,
	}
}
