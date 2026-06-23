package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"
)

var sensitiveFields = map[string]bool{
	"accountNumber": true,
	"ssn":           true,
	"password":      true,
	"cardNumber":    true,
}

func redactJSON(v any) any {
	switch val := v.(type) {
	case map[string]any:
		for k, child := range val {
			if sensitiveFields[k] {
				val[k] = "[REDACTED]"
			} else {
				val[k] = redactJSON(child)
			}
		}
		return val
	case []any:
		for i, item := range val {
			val[i] = redactJSON(item)
		}
		return val
	default:
		return v
	}
}

var (
	colorGreen  = color.New(color.FgGreen, color.Bold).SprintFunc()
	colorYellow = color.New(color.FgYellow, color.Bold).SprintFunc()
	colorRed    = color.New(color.FgRed, color.Bold).SprintFunc()
	colorCyan   = color.New(color.FgCyan).SprintFunc()
	colorGray   = color.New(color.FgWhite).SprintFunc()
)

func statusColor(code int) string {
	s := strconv.Itoa(code)
	switch {
	case code >= 500:
		return colorRed(s)
	case code >= 400:
		return colorYellow(s)
	default:
		return colorGreen(s)
	}
}

func methodColor(method string) string {
	switch method {
	case "GET":
		return colorCyan(method)
	case "POST", "PUT", "PATCH":
		return colorGreen(method)
	case "DELETE":
		return colorRed(method)
	default:
		return colorGray(method)
	}
}

// LoggingMiddleware logs each request with method, path, status, latency, and redacted body.
func LoggingMiddleware(c *fiber.Ctx) error {
	start := time.Now()
	err := c.Next()
	latency := time.Since(start)

	status := c.Response().StatusCode()
	method := c.Method()
	path := c.Path()

	// Format latency cleanly: <1ms shows µs, otherwise ms
	var latStr string
	if latency < time.Millisecond {
		latStr = fmt.Sprintf("%dµs", latency.Microseconds())
	} else {
		latStr = fmt.Sprintf("%dms", latency.Milliseconds())
	}

	// Redact sensitive fields from request body for logging
	var bodyLog string
	if len(c.Body()) > 0 {
		var parsed any
		if json.Unmarshal(c.Body(), &parsed) == nil {
			redacted := redactJSON(parsed)
			if b, e := json.Marshal(redacted); e == nil {
				bodyLog = string(b)
			}
		}
	}

	line := fmt.Sprintf("%s %s %s  %s",
		methodColor(fmt.Sprintf("%-6s", method)),
		colorGray(path),
		statusColor(status),
		colorGray(latStr),
	)
	if bodyLog != "" {
		line += fmt.Sprintf("  body=%s", colorGray(bodyLog))
	}

	fmt.Println(line)
	return err
}

// AuthMiddleware reads X-User-ID and injects it into Locals.
func AuthMiddleware(c *fiber.Ctx) error {
	raw := c.Get("X-User-ID")
	if raw == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing X-User-ID header"})
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid X-User-ID"})
	}
	c.Locals("userID", id)
	return c.Next()
}
