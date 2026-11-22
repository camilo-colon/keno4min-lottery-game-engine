package domain

import (
	"slices"
	"testing"
)

var updateTicketsPaymentTable = map[int][]int64{
	1:  {0, 350},
	2:  {0, 100, 1000},
	3:  {0, 0, 200, 5000},
	4:  {0, 0, 100, 1000, 10000},
	5:  {0, 0, 100, 300, 2000, 15000},
	6:  {0, 0, 0, 200, 1500, 6000, 50000},
	7:  {100, 0, 0, 200, 400, 2000, 8000, 100000},
	8:  {100, 0, 0, 0, 500, 1500, 5000, 20000, 200000},
	9:  {200, 0, 0, 0, 100, 1000, 2500, 12500, 100000, 500000},
	10: {200, 0, 0, 0, 0, 500, 3000, 10000, 30000, 200000, 1000000},
}

func TestPaymentTableMatchesUpdateTickets(t *testing.T) {
	if len(paymentTable) != 11 {
		t.Fatalf("table has %d rows, want 11", len(paymentTable))
	}
	for spots, got := range paymentTable {
		want := updateTicketsPaymentTable[spots]
		if !slices.Equal(got, want) {
			t.Errorf("spots %d: row = %v, want %v", spots, got, want)
		}
	}
}
