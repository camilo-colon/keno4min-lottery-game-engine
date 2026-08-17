package domain

import "testing"

func TestPaymentTableTwoSpot(t *testing.T) {
	factors := PaymentTable[2]
	if len(factors) != 3 {
		t.Fatalf("PaymentTable[2] has %d factors, want 3", len(factors))
	}

	if factors[1] != 100 {
		t.Errorf("one hit factor = %d, want 100 (1x)", factors[1])
	}
	if factors[2] != 1000 {
		t.Errorf("two hit factor = %d, want 1000 (10x)", factors[2])
	}
}
