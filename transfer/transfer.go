package transfer

import (
	"database/sql"
	"errors"
	"fmt"
)

// Transfer moves amount from fromID to toID idempotently.
// If the idempotency key has already been used, it returns nil (already processed).
func Transfer(db *sql.DB, fromID, toID int64, amount float64, idempotencyKey string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert idempotency record; skip if already exists.
	res, err := tx.Exec(
		`INSERT INTO transfers (idempotency_key, from_id, to_id, amount)
		 VALUES ($1, $2, $3, $4) ON CONFLICT (idempotency_key) DO NOTHING`,
		idempotencyKey, fromID, toID, amount,
	)
	if err != nil {
		return fmt.Errorf("insert transfer record: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		// Already processed — idempotent no-op.
		return nil
	}

	// Lock sender row to prevent concurrent overdrafts.
	var currentBalance float64
	err = tx.QueryRow(
		`SELECT balance FROM accounts WHERE id = $1 FOR UPDATE`,
		fromID,
	).Scan(&currentBalance)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("sender account %d not found", fromID)
	}
	if err != nil {
		return fmt.Errorf("fetch sender balance: %w", err)
	}
	if currentBalance < amount {
		return fmt.Errorf("insufficient funds: have %.4f, need %.4f", currentBalance, amount)
	}

	if _, err = tx.Exec(
		`UPDATE accounts SET balance = balance - $1 WHERE id = $2`,
		amount, fromID,
	); err != nil {
		return fmt.Errorf("deduct from sender: %w", err)
	}

	res2, err := tx.Exec(
		`UPDATE accounts SET balance = balance + $1 WHERE id = $2`,
		amount, toID,
	)
	if err != nil {
		return fmt.Errorf("credit receiver: %w", err)
	}
	if affected, _ := res2.RowsAffected(); affected == 0 {
		return fmt.Errorf("receiver account %d not found", toID)
	}

	// Write debit + credit rows so transaction history reflects the transfer.
	if _, err = tx.Exec(
		`INSERT INTO transactions (user_id, amount, type) VALUES ($1, $2, 'debit'), ($3, $2, 'credit')`,
		fromID, amount, toID,
	); err != nil {
		return fmt.Errorf("insert transaction records: %w", err)
	}

	return tx.Commit()
}
