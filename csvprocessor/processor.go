package csvprocessor

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

// ProcessBalances streams a CSV (userID, timestamp, amount, type) and returns
// each user's final balance. Rows are read one at a time — O(users) memory, not O(rows).
func ProcessBalances(r io.Reader) (map[string]float64, error) {
	reader := csv.NewReader(r)
	reader.ReuseRecord = true // reuse the backing array to reduce allocations

	balances := make(map[string]float64)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read csv: %w", err)
		}
		if len(record) < 4 {
			return nil, fmt.Errorf("expected 4 columns, got %d", len(record))
		}

		userID := record[0]
		amount, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return nil, fmt.Errorf("parse amount %q: %w", record[2], err)
		}
		txType := record[3]

		switch txType {
		case "credit":
			balances[userID] += amount
		case "debit":
			balances[userID] -= amount
		default:
			return nil, fmt.Errorf("unknown transaction type %q", txType)
		}
	}

	return balances, nil
}
