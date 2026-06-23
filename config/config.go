package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	CSVPath     string
	Port        string
}

// Load reads .env (if present) then pulls values from environment.
// Real env vars always take precedence over .env file values.
func Load() (*Config, error) {
	// Silently ignore missing .env — production may use real env vars.
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	csvPath := os.Getenv("CSV_PATH")
	if csvPath == "" {
		csvPath = "data/transactions.csv"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		DatabaseURL: dbURL,
		CSVPath:     csvPath,
		Port:        port,
	}, nil
}
