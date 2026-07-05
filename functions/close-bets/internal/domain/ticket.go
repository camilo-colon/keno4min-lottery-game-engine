package domain

import "context"

type TicketState string

const (
	TicketStatePending TicketState = "PENDING"
	TicketStateDrawing TicketState = "DRAWING"
)

type Ticket struct {
	Id     string      `bson:"_id,omitempty"`
	GameId string      `bson:"game_id"`
	State  TicketState `bson:"state"`
}

type TicketRepository interface {
	GetTicketsByGameID(ctx context.Context, gameID string, cursor *string, limit int64) ([]Ticket, *string, error)
	UpdateTickets(ctx context.Context, tickets []Ticket) error
}
