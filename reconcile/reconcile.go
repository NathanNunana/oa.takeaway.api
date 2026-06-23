package reconcile

import "math"

// Transaction represents a record from either the ledger or the payment processor.
type Transaction struct {
	ID     string
	Amount float64 // dollars; converted to cents internally
}

// Result holds the reconciliation outcome.
type Result struct {
	OnlyInLedger    []string // IDs in ledger but not processor
	OnlyInProcessor []string // IDs in processor but not ledger
	AmountMismatch  []string // IDs in both but with different amounts
	TotalDiff       int64    // net difference in cents (ledger - processor)
}

func toCents(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

// Reconcile compares ledger vs processor transactions.
// Common $0.01 discrepancies arise from: float rounding, currency conversion,
// fee deductions, partial refunds, or duplicate processing on one side.
func Reconcile(ledger, processor []Transaction) Result {
	ledgerMap := make(map[string]int64, len(ledger))
	for _, t := range ledger {
		ledgerMap[t.ID] = toCents(t.Amount)
	}

	processorMap := make(map[string]int64, len(processor))
	for _, t := range processor {
		processorMap[t.ID] = toCents(t.Amount)
	}

	var result Result

	// Check every ledger entry against processor.
	for id, lAmt := range ledgerMap {
		pAmt, found := processorMap[id]
		if !found {
			result.OnlyInLedger = append(result.OnlyInLedger, id)
			result.TotalDiff += lAmt
		} else if lAmt != pAmt {
			result.AmountMismatch = append(result.AmountMismatch, id)
			result.TotalDiff += lAmt - pAmt
		}
	}

	// Find processor entries absent from ledger.
	for id, pAmt := range processorMap {
		if _, found := ledgerMap[id]; !found {
			result.OnlyInProcessor = append(result.OnlyInProcessor, id)
			result.TotalDiff -= pAmt
		}
	}

	return result
}
