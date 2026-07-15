package ports

import (
	"context"

	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets/internal/domain"
)

// TicketRepository define las operaciones necesarias sobre la colección de tickets.
//
// Regla de negocio del puerto: las consultas SOLO deben traer tickets en estado
// DRAWING, nunca "todos los que no estén CANCELED". close-bets mueve
// PENDING -> DRAWING justo antes del sorteo, así que DRAWING es exactamente
// "sorteado, pendiente de resolver". Filtrar por DRAWING hace esta Lambda
// idempotente por descarte: si el Step Functions reintenta después de que un
// cajero ya marcó el ticket como PAYED (o fue CANCELED), un segundo run no lo
// encuentra y no lo toca. Filtrar por "!= CANCELED" sería un BUG: revertiría
// tickets PAYED a WINNING en un reintento.
type TicketRepository interface {
	// FindDrawingByGame devuelve una página de tickets en estado DRAWING del
	// juego, ordenados por _id, para paginar juegos con muchos tickets. cursor
	// nil trae la primera página; el cursor devuelto es nil cuando no hay más
	// páginas.
	FindDrawingByGame(ctx context.Context, gameID string, cursor *string, limit int64) ([]domain.Ticket, *string, error)
	// UpdateTickets persiste win, state y balls de los tickets recibidos en un
	// solo BulkWrite.
	UpdateTickets(ctx context.Context, tickets []domain.Ticket) error
}
