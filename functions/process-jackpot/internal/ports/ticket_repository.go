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
	// AssignJackpot persiste el premio del pozo del ticket ganador: el monto y el
	// estado, tal como los deja Ticket.AwardJackpot. El estado viaja con el
	// ticket porque ganar el pozo lo vuelve cobrable, y esa derivación es una
	// regla de dominio, no del store.
	AssignJackpot(ctx context.Context, ticket domain.Ticket) error
}
