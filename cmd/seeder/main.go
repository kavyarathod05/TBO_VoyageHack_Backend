package main

import (
	"log"

	"github.com/akashtripathi12/TBO_Backend/internal/store"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	store.InitDB()

	log.Println("⚠️  STARTING DATABASE RESET...")

	store.DB.Exec("TRUNCATE TABLE users CASCADE")
	store.DB.Exec("TRUNCATE TABLE guests CASCADE")
	store.DB.Exec("TRUNCATE TABLE agent_profiles CASCADE")
	store.DB.Exec("TRUNCATE TABLE events CASCADE")

	log.Println("✅ Database Cleared.")

	log.Println("🌱 Seeding new data...")

	// populating logic here

	log.Println("🎉 Database Reset & Populated Successfully!")
}
