package handler

import (
	"context"

	"github.com/cronos/keno4min-lottery-game-engine/functions/draw-balls/internal/domain"
)

// GameRepository define las operaciones necesarias sobre la colección de juegos.
type GameRepository interface {
	FindByID(ctx context.Context, gameID string) (*domain.Game, error)
	UpdateGame(ctx context.Context, game *domain.Game) error
	GetHistoryGame(ctx context.Context, limit int64) ([]domain.Game, error)
}

// DrawRepository define las operaciones necesarias sobre la colección de draws.
type DrawRepository interface {
	GetRandomKeno4MinDraw(ctx context.Context) (*domain.Draws, error)
}

type TicketRepository interface {
	GetStats(ctx context.Context) (*domain.Stats, error)
	GetTicketsByGame(ctx context.Context, gameID string) ([]domain.Ticket, error)
}
