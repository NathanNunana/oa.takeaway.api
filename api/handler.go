package api

import (
	"database/sql"
	"os"

	"github.com/gofiber/fiber/v2"

	"github.com/nate/takeway/csvprocessor"
	"github.com/nate/takeway/ratelimit"
	"github.com/nate/takeway/reconcile"
	"github.com/nate/takeway/transfer"
)

// Problem 1: POST /api/transfer
func HandleTransfer(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			FromID         int64   `json:"from_id"`
			ToID           int64   `json:"to_id"`
			Amount         float64 `json:"amount"`
			IdempotencyKey string  `json:"idempotency_key"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		if req.IdempotencyKey == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "idempotency_key required"})
		}

		// Ensure the caller can only transfer from their own account.
		userID := c.Locals("userID").(int64)
		if req.FromID != userID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "cannot transfer from another user's account"})
		}

		if err := transfer.Transfer(db, req.FromID, req.ToID, req.Amount, req.IdempotencyKey); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}

// Problem 2: GET /api/balances
func HandleBalances(csvPath string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		f, err := os.Open(csvPath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not open csv"})
		}
		defer f.Close()

		balances, err := csvprocessor.ProcessBalances(f)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(balances)
	}
}

// Problem 3: POST /api/pay (rate-limited, requires auth)
func HandlePay(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(int64)

		var req struct {
			Amount float64 `json:"amount"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}

		if err := ratelimit.Allow(db, userID); err != nil {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "payment accepted", "user_id": userID, "amount": req.Amount})
	}
}

// Problem 4: GET /api/transactions (IDOR fix — userId query param ignored)
func HandleGetTransactions(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(int64)

		rows, err := db.QueryContext(c.Context(),
			`SELECT id, amount, type, created_at FROM transactions WHERE user_id = $1 ORDER BY created_at DESC`,
			userID,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db error"})
		}
		defer rows.Close()

		type Transaction struct {
			ID        int64   `json:"id"`
			Amount    float64 `json:"amount"`
			Type      string  `json:"type"`
			CreatedAt string  `json:"created_at"`
		}

		txns := make([]Transaction, 0)
		for rows.Next() {
			var t Transaction
			if err := rows.Scan(&t.ID, &t.Amount, &t.Type, &t.CreatedAt); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "scan error"})
			}
			txns = append(txns, t)
		}
		return c.JSON(txns)
	}
}

// GET /api/accounts — shows all accounts and their current balances (useful for demo)
func HandleGetAccounts(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := db.QueryContext(c.Context(),
			`SELECT id, owner_id, balance FROM accounts ORDER BY id`,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db error"})
		}
		defer rows.Close()

		type Account struct {
			ID      int64   `json:"id"`
			OwnerID int64   `json:"owner_id"`
			Balance float64 `json:"balance"`
		}
		accounts := make([]Account, 0)
		for rows.Next() {
			var a Account
			if err := rows.Scan(&a.ID, &a.OwnerID, &a.Balance); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "scan error"})
			}
			accounts = append(accounts, a)
		}
		return c.JSON(accounts)
	}
}
func HandleReconcile() fiber.Handler {
	ledger := []reconcile.Transaction{
		{ID: "t1", Amount: 100.00},
		{ID: "t2", Amount: 200.50},
		{ID: "t3", Amount: 50.00},
		{ID: "t4", Amount: 75.25},
	}
	processor := []reconcile.Transaction{
		{ID: "t1", Amount: 100.00},
		{ID: "t2", Amount: 200.51}, // $0.01 mismatch
		{ID: "t5", Amount: 30.00},  // only in processor
	}

	return func(c *fiber.Ctx) error {
		result := reconcile.Reconcile(ledger, processor)
		return c.JSON(fiber.Map{
			"only_in_ledger":    result.OnlyInLedger,
			"only_in_processor": result.OnlyInProcessor,
			"amount_mismatch":   result.AmountMismatch,
			"total_diff_cents":  result.TotalDiff,
		})
	}
}
