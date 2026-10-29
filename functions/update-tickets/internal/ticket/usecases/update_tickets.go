package usecases

import (
	"context"
	"errors"
	"fmt"

	ticket "github.com/cronos/keno4min-draw-engine/functions/update-tickets/internal/ticket/domain"
	"github.com/cronos/keno4min-draw-engine/functions/update-tickets/internal/ticket/ports"
)

var ErrNoBalls = errors.New("draw has no balls")

type UpdateTickets struct {
	tickets ports.Repository
}

func NewUpdateTickets(tickets ports.Repository) *UpdateTickets {
	return &UpdateTickets{tickets: tickets}
}

// Execute liquida los tickets pendientes del sorteo y devuelve cuántos actualizó.
func (u *UpdateTickets) Execute(ctx context.Context, tir uint64, balls []int) (int, error) {
	if len(balls) == 0 {
		return 0, ErrNoBalls
	}
	tickets, err := u.tickets.ListByTIR(ctx, tir)
	if err != nil {
		return 0, fmt.Errorf("list tickets: %w", err)
	}
	updated := 0
	for i := range tickets {
		t := &tickets[i]
		if t.Status != ticket.StatusPending {
			continue
		}
		t.Settle(balls)
		if err := u.tickets.Update(ctx, t); err != nil {
			return updated, fmt.Errorf("update ticket %s: %w", t.ID, err)
		}
		updated++
	}
	return updated, nil
}
