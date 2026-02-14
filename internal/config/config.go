package config

import (
	"os"
	"time"
)

type Config struct {
	Port           string
	Env            string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	BodyLimit      int
	AllowedOrigins []string
	TrustedProxies []string
	EnableLogger   bool
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	return &Config{
		Port:           ":" + port,
		Env:            env,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		BodyLimit:      4 * 1024 * 1024, // 4MB
		AllowedOrigins: []string{"*"},   // TODO: Restrict in production
		TrustedProxies: []string{},
		EnableLogger:   true,
	}
}
