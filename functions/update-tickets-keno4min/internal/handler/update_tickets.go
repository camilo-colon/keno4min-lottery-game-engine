package handler

import (
	"context"
	"fmt"

	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets-keno4min/internal/domain"
	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets-keno4min/internal/ports"
)

// DefaultBatchSize es el tamaño de página usado para paginar los tickets de un
// juego al resolverlos, igual que close-bets pagina el pase PENDING -> DRAWING.
const DefaultBatchSize int64 = 500

// UpdateTicketsHandler resuelve los tickets DRAWING de un juego contra las
// balotas sorteadas por DrawBalls.
type UpdateTicketsHandler struct {
	tickets   ports.TicketRepository
	batchSize int64
}

// NewUpdateTicketsHandler crea un nuevo handler para resolver tickets.
func NewUpdateTicketsHandler(tickets ports.TicketRepository, batchSize int64) *UpdateTicketsHandler {
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}
	return &UpdateTicketsHandler{tickets: tickets, batchSize: batchSize}
}

// Handle resuelve todos los tickets DRAWING del juego: calcula win y state
// contra game.Balls y les copia las balotas sorteadas. Devuelve la cantidad de
// tickets resueltos.
func (h *UpdateTicketsHandler) Handle(ctx context.Context, game domain.Game) (int, error) {
	if game.Balls == nil {
		return 0, fmt.Errorf("game %s has no drawn balls", game.ID)
	}
	balls := *game.Balls

	var cursor *string
	total := 0
	for {
		tickets, next, err := h.tickets.FindDrawingByGame(ctx, game.ID, cursor, h.batchSize)
		if err != nil {
			return 0, err
		}
		if len(tickets) == 0 {
			break
		}

		for i := range tickets {
			tickets[i].Resolve(balls)
		}

		if err := h.tickets.UpdateTickets(ctx, tickets); err != nil {
			return 0, err
		}

		total += len(tickets)
		cursor = next
	}

	return total, nil
}
