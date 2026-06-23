package reconcile

import (
	"sort"
	"testing"
)

func sorted(s []string) []string {
	sort.Strings(s)
	return s
}

func TestReconcile_AllMatch(t *testing.T) {
	ledger := []Transaction{{ID: "t1", Amount: 100.00}, {ID: "t2", Amount: 50.00}}
	processor := []Transaction{{ID: "t1", Amount: 100.00}, {ID: "t2", Amount: 50.00}}

	r := Reconcile(ledger, processor)

	if len(r.OnlyInLedger) != 0 || len(r.OnlyInProcessor) != 0 || len(r.AmountMismatch) != 0 || r.TotalDiff != 0 {
		t.Errorf("expected clean reconciliation, got %+v", r)
	}
}

func TestReconcile_OnlyInLedger(t *testing.T) {
	ledger := []Transaction{{ID: "t1", Amount: 100.00}, {ID: "t2", Amount: 50.00}}
	processor := []Transaction{{ID: "t1", Amount: 100.00}}

	r := Reconcile(ledger, processor)

	if len(r.OnlyInLedger) != 1 || r.OnlyInLedger[0] != "t2" {
		t.Errorf("expected t2 only in ledger, got %v", r.OnlyInLedger)
	}
	if r.TotalDiff != 5000 {
		t.Errorf("expected TotalDiff=5000, got %d", r.TotalDiff)
	}
}

func TestReconcile_OnlyInProcessor(t *testing.T) {
	ledger := []Transaction{{ID: "t1", Amount: 100.00}}
	processor := []Transaction{{ID: "t1", Amount: 100.00}, {ID: "t2", Amount: 75.00}}

	r := Reconcile(ledger, processor)

	if len(r.OnlyInProcessor) != 1 || r.OnlyInProcessor[0] != "t2" {
		t.Errorf("expected t2 only in processor, got %v", r.OnlyInProcessor)
	}
	if r.TotalDiff != -7500 {
		t.Errorf("expected TotalDiff=-7500, got %d", r.TotalDiff)
	}
}

func TestReconcile_AmountMismatch_OneCent(t *testing.T) {
	ledger := []Transaction{{ID: "t1", Amount: 200.50}}
	processor := []Transaction{{ID: "t1", Amount: 200.51}}

	r := Reconcile(ledger, processor)

	if len(r.AmountMismatch) != 1 || r.AmountMismatch[0] != "t1" {
		t.Errorf("expected t1 in mismatch, got %v", r.AmountMismatch)
	}
	if r.TotalDiff != -1 {
		t.Errorf("expected TotalDiff=-1 cent, got %d", r.TotalDiff)
	}
}

func TestReconcile_Mixed(t *testing.T) {
	ledger := []Transaction{
		{ID: "t1", Amount: 100.00},
		{ID: "t2", Amount: 200.50},
		{ID: "t3", Amount: 50.00},
	}
	processor := []Transaction{
		{ID: "t1", Amount: 100.00},
		{ID: "t2", Amount: 200.51},
		{ID: "t4", Amount: 30.00},
	}

	r := Reconcile(ledger, processor)

	if got := sorted(r.OnlyInLedger); len(got) != 1 || got[0] != "t3" {
		t.Errorf("expected [t3] only in ledger, got %v", got)
	}
	if got := sorted(r.OnlyInProcessor); len(got) != 1 || got[0] != "t4" {
		t.Errorf("expected [t4] only in processor, got %v", got)
	}
	if got := sorted(r.AmountMismatch); len(got) != 1 || got[0] != "t2" {
		t.Errorf("expected [t2] mismatch, got %v", got)
	}
}

func TestReconcile_FloatPrecision(t *testing.T) {
	// 0.1 + 0.2 != 0.3 in float64 — ensure we handle it correctly via cents
	ledger := []Transaction{{ID: "t1", Amount: 0.10}, {ID: "t2", Amount: 0.20}}
	processor := []Transaction{{ID: "t1", Amount: 0.10}, {ID: "t2", Amount: 0.20}}

	r := Reconcile(ledger, processor)

	if r.TotalDiff != 0 || len(r.AmountMismatch) != 0 {
		t.Errorf("float precision failure: got diff=%d mismatch=%v", r.TotalDiff, r.AmountMismatch)
	}
}

func TestReconcile_Empty(t *testing.T) {
	r := Reconcile(nil, nil)
	if r.TotalDiff != 0 || len(r.OnlyInLedger) != 0 || len(r.OnlyInProcessor) != 0 {
		t.Errorf("expected empty result for empty inputs, got %+v", r)
	}
}
