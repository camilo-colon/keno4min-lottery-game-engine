package handler_test

import (
	"context"

	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets/internal/domain"
)

// fakeTickets simula el repositorio de tickets paginando en memoria por páginas
// preconfiguradas y registrando lo que se escribió en cada BulkWrite.
type fakeTickets struct {
	pages      [][]domain.Ticket
	findErr    error
	updateErr  error
	updateCall [][]domain.Ticket
	findCalls  int
}

func (f *fakeTickets) FindDrawingByGame(ctx context.Context, gameID string, cursor *string, limit int64) ([]domain.Ticket, *string, error) {
	idx := f.findCalls
	f.findCalls++

	if f.findErr != nil {
		return nil, nil, f.findErr
	}
	if idx >= len(f.pages) {
		return nil, nil, nil
	}

	page := f.pages[idx]
	if idx == len(f.pages)-1 {
		return page, nil, nil
	}
	next := "cursor"
	return page, &next, nil
}

func (f *fakeTickets) UpdateTickets(ctx context.Context, tickets []domain.Ticket) error {
	// copiamos para que las mutaciones posteriores del slice del handler no
	// afecten lo ya registrado.
	snapshot := make([]domain.Ticket, len(tickets))
	copy(snapshot, tickets)
	f.updateCall = append(f.updateCall, snapshot)
	return f.updateErr
}

// drawnGame construye un game DRAWN con balotas y sus números.
func drawnGame(id string, nums []int64, balls domain.Bitmask) domain.Game {
	return domain.Game{ID: id, Balls: &domain.GameBalls{Nums: nums, Mask: balls}}
}
