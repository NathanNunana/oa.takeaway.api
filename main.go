package main

import (
	"database/sql"
	"log"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"

	"github.com/nate/takeway/api"
	"github.com/nate/takeway/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	app := fiber.New()
	app.Use(api.LoggingMiddleware) // Problem 5: logging + redaction on all routes

	// Public
	app.Get("/health", api.HandleHealth(db))                  // Health check
	app.Get("/api/balances", api.HandleBalances(cfg.CSVPath)) // Problem 2
	app.Get("/api/reconcile", api.HandleReconcile())          // Problem 6
	app.Get("/api/accounts", api.HandleGetAccounts(db))       // Demo: watch balances change

	// Protected (require X-User-ID header)
	protected := app.Group("/api", api.AuthMiddleware)
	protected.Post("/transfer", api.HandleTransfer(db))           // Problem 1
	protected.Post("/pay", api.HandlePay(db))                     // Problem 3
	protected.Get("/transactions", api.HandleGetTransactions(db)) // Problem 4

	log.Printf("listening on :%s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
