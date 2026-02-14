package store

import (
	"log"
	"os"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := os.Getenv("DATABASE_URL")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Connected to Supabase")

	// Sync schema to DB
	db.AutoMigrate(
		&models.User{},
		&models.AgentProfile{},
		&models.Event{},
		&models.Guest{},
	)

	DB = db
	log.Println("✅ schema synced")
}
