package ports

import (
	"context"
	"errors"
)

// ErrAlreadyProcessed indica que el jackpot de un (game, club) ya fue procesado
// en una corrida previa; el llamador debe saltarlo (skip idempotente).
var ErrAlreadyProcessed = errors.New("jackpot already processed for game and club")

// RunRepository registra el procesamiento del jackpot por (game, club) para
// garantizar idempotencia ante reintentos.
type RunRepository interface {
	// Mark registra el par (gameID, clubID). Devuelve ErrAlreadyProcessed si ya
	// existía, actuando como guardián de idempotencia.
	Mark(ctx context.Context, gameID, clubID string) error
}
