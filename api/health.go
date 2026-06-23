package api

import (
	"database/sql"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
)

var startTime = time.Now()

func HandleHealth(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// DB connectivity
		dbStatus := "ok"
		dbLatency := time.Duration(0)
		t := time.Now()
		if err := db.PingContext(c.Context()); err != nil {
			dbStatus = "unreachable: " + err.Error()
		} else {
			dbLatency = time.Since(t)
		}

		// Runtime memory
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		status := fiber.StatusOK
		if dbStatus != "ok" {
			status = fiber.StatusServiceUnavailable
		}

		return c.Status(status).JSON(fiber.Map{
			"status": map[string]any{
				"api": "ok",
				"db":  dbStatus,
			},
			"uptime":        time.Since(startTime).Round(time.Second).String(),
			"db_latency_ms": dbLatency.Milliseconds(),
			"memory": fiber.Map{
				"alloc_mb":   mem.Alloc / 1024 / 1024,
				"sys_mb":     mem.Sys / 1024 / 1024,
				"num_gc":     mem.NumGC,
				"goroutines": runtime.NumGoroutine(),
			},
		})
	}
}
