package ports

import (
	"context"

	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/domain"
)

// TicketRepository define las operaciones necesarias sobre la colección de tickets.
//
// Regla de negocio del puerto: los tickets CANCELED no participan del jackpot, así
// que las consultas de este puerto NUNCA los devuelven (ni para el incremento ni
// para elegir ganador).
type TicketRepository interface {
	// FindClubIDsByGame devuelve los IDs de los clubes que participaron en el
	// juego (clubes con al menos un ticket NO cancelado para ese game_id).
	FindClubIDsByGame(ctx context.Context, gameID string) ([]string, error)
	// FindByClubAndGame devuelve los tickets NO cancelados de un club en un juego.
	FindByClubAndGame(ctx context.Context, clubID, gameID string) ([]domain.Ticket, error)
	// AssignJackpot asigna el premio del pozo al ticket ganador.
	AssignJackpot(ctx context.Context, ticketID string, amount int64) error
}
