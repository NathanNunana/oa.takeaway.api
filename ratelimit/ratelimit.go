package ratelimit

import (
	"database/sql"
	"fmt"
	"time"
)

const maxPerMinute = 5

// Allow checks whether userID is under the rate limit, and records the attempt if so.
// The count + insert are wrapped in a transaction to prevent races.
func Allow(db *sql.DB, userID int64) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	window := time.Now().Add(-time.Minute)

	var count int
	err = tx.QueryRow(
		`SELECT COUNT(*) FROM rate_limit_log WHERE user_id = $1 AND created_at > $2`,
		userID, window,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("count requests: %w", err)
	}

	if count >= maxPerMinute {
		return fmt.Errorf("rate limit exceeded: %d/%d requests in the last minute", count, maxPerMinute)
	}

	_, err = tx.Exec(
		`INSERT INTO rate_limit_log (user_id) VALUES ($1)`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("insert log entry: %w", err)
	}

	return tx.Commit()
}
