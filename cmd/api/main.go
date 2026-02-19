package main

import (
	"log"
	"os"

	"github.com/akashtripathi12/TBO_Backend/internal/config"
	"github.com/akashtripathi12/TBO_Backend/internal/handlers"
	"github.com/akashtripathi12/TBO_Backend/internal/routes"
	"github.com/akashtripathi12/TBO_Backend/internal/store"

	"github.com/akashtripathi12/TBO_Backend/internal/queue"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	// Load Config
	cfg := config.Load()

	// Initialize Store
	store.InitDB()
	log.Println("✅ DB Connected. URL:", os.Getenv("DATABASE_URL"))

	// Initialize Redis
	store.InitRedis(cfg)

	// --- Asynq Redis Config ---
	redisAddr := "127.0.0.1:6379"
	if val := os.Getenv("REDIS_URL"); val != "" {
		redisAddr = val
	}
	redisOpt := asynq.RedisClientOpt{Addr: redisAddr}

	// 1. Initialize Asynq Client (Producer)
	client := asynq.NewClient(redisOpt)
	defer client.Close()

	// Initialize Repository with Queue Client
	repo := handlers.NewRepository(cfg, store.DB, client)

	// 2. Initialize Asynq Server (Consumer)
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"default": 10,
			},
		},
	)

	// Register Task Handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TypeEmailDelivery, queue.HandleEmailTask)

	// Run Worker in Background
	go func() {
		log.Println("👷 Asynq Worker Server Starting...")
		if err := srv.Run(mux); err != nil {
			log.Printf("❌ Asynq Server Failed: %v", err)
		}
	}()

	app := fiber.New()

	// Enable CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // For development
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Setup Routes
	routes.SetupRoutes(app, cfg, repo)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("TBO Backend Operational 🚀")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen(":" + port))
}
