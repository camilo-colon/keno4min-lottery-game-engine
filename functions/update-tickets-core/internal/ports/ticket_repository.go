package ports

import (
	"context"

	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets-core/internal/domain"
)

// TicketRepository define las operaciones necesarias sobre la colección de
// tickets del microservicio core.
//
// Regla de negocio del puerto: las consultas SOLO deben traer tickets en estado
// PENDING del juego KENO4MIN. A diferencia del engine, el core NO tiene un gate
// (no hay close-bets que mueva PENDING -> DRAWING), así que PENDING es el estado
// pre-sorteo. Filtrar por PENDING hace esta Lambda idempotente por descarte: si
// Step Functions reintenta después de que un ticket ya fue resuelto (WINNING /
// LOSS) o cobrado (PAYED) o cancelado (CANCELED), un segundo run no lo encuentra
// y no lo toca. Filtrar por "!= CANCELED" sería un BUG: revertiría tickets PAYED
// a WINNING en un reintento.
type TicketRepository interface {
	// FindPendingByGame devuelve una página de tickets PENDING de KENO4MIN del
	// juego, ordenados por _id, para paginar juegos con muchos tickets. cursor
	// nil trae la primera página; el cursor devuelto es nil cuando no hay más
	// páginas.
	FindPendingByGame(ctx context.Context, gameID string, cursor *string, limit int64) ([]domain.Ticket, *string, error)
	// UpdateTickets persiste win y state de los tickets recibidos en un solo
	// BulkWrite. El ticket del core no almacena balotas, así que no se tocan.
	UpdateTickets(ctx context.Context, tickets []domain.Ticket) error
}
