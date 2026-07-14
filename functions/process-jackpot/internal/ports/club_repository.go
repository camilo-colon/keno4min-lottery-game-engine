package ports

import (
	"context"

	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/domain"
)

// ClubRepository define las operaciones necesarias sobre la colección de clubes.
// El club es el aggregate root; el jackpot (jp1) vive embebido en él.
type ClubRepository interface {
	FindByID(ctx context.Context, clubID string) (*domain.Club, error)
	// IncrementJackpot suma de forma atómica un monto al pozo jp1 del club y
	// devuelve el estado del jackpot ya incrementado.
	IncrementJackpot(ctx context.Context, clubID string, amount int64) (*domain.Jackpot, error)
	// ResetJackpot reemplaza el pozo jp1 del club con un jackpot fresco.
	ResetJackpot(ctx context.Context, clubID string, jackpot domain.Jackpot) error
}
