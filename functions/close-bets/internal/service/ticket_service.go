package service

import (
	"context"
	"fmt"

	"github.com/cronos/keno4min-lottery-game-engine/functions/close-bets/internal/domain"
)

const DefaultBatchSize int64 = 500

type TicketService struct {
	ticketRepo domain.TicketRepository
	batchSize  int64
}

func NewTicketService(ticketRepo domain.TicketRepository, batchSize int64) *TicketService {
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}

	return &TicketService{
		ticketRepo: ticketRepo,
		batchSize:  batchSize,
	}
}

func (s *TicketService) MovePendingToDrawing(ctx context.Context, gameID string) (int, error) {
	var nextCursor *string
	totalUpdated := 0

	for {
		tickets, cursor, err := s.ticketRepo.GetTicketsByGameID(ctx, gameID, nextCursor, s.batchSize)
		if err != nil {
			return 0, fmt.Errorf("error fetching pending tickets: %w", err)
		}

		if len(tickets) == 0 {
			break
		}

		for i := range tickets {
			tickets[i].State = domain.TicketStateDrawing
		}

		if err := s.ticketRepo.UpdateTickets(ctx, tickets); err != nil {
			return 0, fmt.Errorf("error updating tickets to DRAWING: %w", err)
		}

		totalUpdated += len(tickets)
		nextCursor = cursor
	}

	return totalUpdated, nil
}
