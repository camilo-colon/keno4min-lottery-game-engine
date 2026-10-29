package ports

import (
	"context"

	ticket "github.com/cronos/keno4min-draw-engine/functions/update-tickets/internal/ticket/domain"
)

type Repository interface {
	ListByTIR(ctx context.Context, tir uint64) ([]ticket.Ticket, error)
	Update(ctx context.Context, t *ticket.Ticket) error
}
