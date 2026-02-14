package main

import (
	"log"
	"os"

	"github.com/akashtripathi12/TBO_Backend/internal/store"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	// Initialize Store
	store.InitDB()

	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("TBO Backend Operational 🚀")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen(":" + port))
}
