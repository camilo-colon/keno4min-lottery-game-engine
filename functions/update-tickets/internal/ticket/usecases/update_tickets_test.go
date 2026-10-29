package usecases

import (
	"context"
	"errors"
	"slices"
	"testing"

	ticket "github.com/cronos/keno4min-draw-engine/functions/update-tickets/internal/ticket/domain"
)

type fakeRepository struct {
	tickets   []ticket.Ticket
	listErr   error
	updateErr error
	updated   []ticket.Ticket
}

func (f *fakeRepository) ListByTIR(_ context.Context, _ uint64) ([]ticket.Ticket, error) {
	return f.tickets, f.listErr
}

func (f *fakeRepository) Update(_ context.Context, t *ticket.Ticket) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updated = append(f.updated, *t)
	return nil
}

var balls = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

func TestExecuteSettlesOnlyPendingTickets(t *testing.T) {
	repo := &fakeRepository{tickets: []ticket.Ticket{
		{ID: "win", Status: ticket.StatusPending, Bets: []ticket.Bet{{Nums: []int{1}, ValueCents: 1000}}},
		{ID: "lost", Status: ticket.StatusPending, Bets: []ticket.Bet{{Nums: []int{80}, ValueCents: 1000}}},
		{ID: "cancelled", Status: ticket.StatusCancelled, Bets: []ticket.Bet{{Nums: []int{1}, ValueCents: 1000}}},
	}}
	updated, err := NewUpdateTickets(repo).Execute(t.Context(), 7, balls)
	if err != nil {
		t.Fatal(err)
	}
	if updated != 2 || len(repo.updated) != 2 {
		t.Fatalf("updated %d tickets, saved %d", updated, len(repo.updated))
	}
	if got := repo.updated[0]; got.ID != "win" || got.Status != ticket.StatusWinning || got.PrizeCents != 3500 || !slices.Equal(got.Balls, balls) {
		t.Fatalf("unexpected winning ticket: %+v", got)
	}
	if got := repo.updated[1]; got.ID != "lost" || got.Status != ticket.StatusLost || got.PrizeCents != 0 || !slices.Equal(got.Balls, balls) {
		t.Fatalf("unexpected lost ticket: %+v", got)
	}
}

func TestExecuteRequiresBalls(t *testing.T) {
	repo := &fakeRepository{}
	if _, err := NewUpdateTickets(repo).Execute(t.Context(), 7, nil); !errors.Is(err, ErrNoBalls) {
		t.Fatalf("expected ErrNoBalls, got %v", err)
	}
}

func TestExecutePropagatesErrors(t *testing.T) {
	failure := errors.New("dynamodb unavailable")
	pending := []ticket.Ticket{{ID: "a", Status: ticket.StatusPending}}
	for name, repo := range map[string]*fakeRepository{
		"list":   {listErr: failure},
		"update": {tickets: pending, updateErr: failure},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewUpdateTickets(repo).Execute(t.Context(), 7, balls); !errors.Is(err, failure) {
				t.Fatalf("expected %v, got %v", failure, err)
			}
		})
	}
}
