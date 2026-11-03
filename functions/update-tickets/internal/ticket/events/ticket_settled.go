package events

import "github.com/cronos/keno4min-draw-engine/functions/update-tickets/internal/common"

var _ common.DomainEvent = (*TicketSettled)(nil)

const TicketSettledType = "cronos.keno4min.ticket.settled"

// TicketSettled does not import the ticket domain, so the aggregate can record it without a cycle.
type TicketSettled struct {
	common.BaseEvent

	TicketID   string `json:"ticketId"`
	Game       string `json:"game"`
	Status     string `json:"status"`
	PrizeCents int64  `json:"prizeCents"`
}

func (t TicketSettled) Type() string {
	return TicketSettledType
}
