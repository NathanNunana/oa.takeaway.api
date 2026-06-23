package csvprocessor

import (
	"strings"
	"testing"
)

func TestProcessBalances_Basic(t *testing.T) {
	input := strings.NewReader(
		"u1,2024-01-01T00:00:00Z,100.00,credit\n" +
			"u1,2024-01-01T00:01:00Z,30.00,debit\n" +
			"u2,2024-01-01T00:02:00Z,200.00,credit\n",
	)

	balances, err := ProcessBalances(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balances["u1"] != 70.00 {
		t.Errorf("u1: expected 70.00, got %.2f", balances["u1"])
	}
	if balances["u2"] != 200.00 {
		t.Errorf("u2: expected 200.00, got %.2f", balances["u2"])
	}
}

func TestProcessBalances_MultipleUsers(t *testing.T) {
	input := strings.NewReader(
		"u1,2024-01-01T00:00:00Z,500.00,credit\n" +
			"u2,2024-01-01T00:01:00Z,300.00,credit\n" +
			"u1,2024-01-01T00:02:00Z,200.00,debit\n" +
			"u3,2024-01-01T00:03:00Z,100.00,credit\n" +
			"u2,2024-01-01T00:04:00Z,50.00,debit\n",
	)

	balances, err := ProcessBalances(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cases := map[string]float64{"u1": 300.00, "u2": 250.00, "u3": 100.00}
	for user, want := range cases {
		if balances[user] != want {
			t.Errorf("%s: expected %.2f, got %.2f", user, want, balances[user])
		}
	}
}

func TestProcessBalances_Empty(t *testing.T) {
	balances, err := ProcessBalances(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error on empty input: %v", err)
	}
	if len(balances) != 0 {
		t.Errorf("expected empty map, got %v", balances)
	}
}

func TestProcessBalances_InvalidType(t *testing.T) {
	input := strings.NewReader("u1,2024-01-01T00:00:00Z,100.00,refund\n")
	_, err := ProcessBalances(input)
	if err == nil {
		t.Error("expected error for unknown transaction type, got nil")
	}
}

func TestProcessBalances_InvalidAmount(t *testing.T) {
	input := strings.NewReader("u1,2024-01-01T00:00:00Z,abc,credit\n")
	_, err := ProcessBalances(input)
	if err == nil {
		t.Error("expected error for invalid amount, got nil")
	}
}

func TestProcessBalances_MissingColumns(t *testing.T) {
	input := strings.NewReader("u1,2024-01-01T00:00:00Z\n")
	_, err := ProcessBalances(input)
	if err == nil {
		t.Error("expected error for missing columns, got nil")
	}
}
