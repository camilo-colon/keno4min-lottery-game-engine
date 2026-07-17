package ports

import (
	"context"

	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets-keno4min/internal/domain"
)

// TicketRepository define las operaciones necesarias sobre la colección de tickets.
//
// Regla de negocio del puerto: las consultas SOLO deben traer tickets en estado
// PENDING, nunca "todos los que no estén CANCELED". Ya no existe close-bets que
// mueva PENDING -> DRAWING antes del sorteo: el cierre de apuestas es temporal
// (closesAt), así que PENDING es exactamente "apuesta cerrada, pendiente de
// resolver". Filtrar por PENDING hace esta Lambda idempotente por descarte: si
// el Step Functions reintenta después de que un cajero ya marcó el ticket como
// PAYED (o fue CANCELED), un segundo run no lo encuentra y no lo toca. Filtrar
// por "!= CANCELED" sería un BUG: revertiría tickets PAYED a WINNING en un
// reintento.
type TicketRepository interface {
	// FindPendingByGame devuelve una página de tickets en estado PENDING del
	// juego, ordenados por _id, para paginar juegos con muchos tickets. cursor
	// nil trae la primera página; el cursor devuelto es nil cuando no hay más
	// páginas.
	FindPendingByGame(ctx context.Context, gameID string, cursor *string, limit int64) ([]domain.Ticket, *string, error)
	// UpdateTickets persiste win, state y balls de los tickets recibidos en un
	// solo BulkWrite.
	UpdateTickets(ctx context.Context, tickets []domain.Ticket) error
}
