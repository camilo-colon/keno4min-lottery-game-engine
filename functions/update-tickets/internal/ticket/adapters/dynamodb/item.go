package dynamodb

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	ticket "github.com/cronos/keno4min-draw-engine/functions/update-tickets/internal/ticket/domain"
)

type ticketItem struct {
	TicketID string    `dynamodbav:"ticketId"`
	TIR      uint64    `dynamodbav:"tir"`
	Bets     []betItem `dynamodbav:"bets"`
	Status   string    `dynamodbav:"status"`
}

type betItem struct {
	Nums       []int `dynamodbav:"nums"`
	ValueCents int64 `dynamodbav:"valueCents"`
}

func (item ticketItem) ticket() ticket.Ticket {
	bets := make([]ticket.Bet, len(item.Bets))
	for i, bet := range item.Bets {
		bets[i] = ticket.Bet{Nums: bet.Nums, ValueCents: bet.ValueCents}
	}
	return ticket.Ticket{
		ID: item.TicketID, TIR: item.TIR, Bets: bets, Status: ticket.Status(item.Status),
	}
}

func key(ticketID string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: "TICKET#" + ticketID},
		"SK": &types.AttributeValueMemberS{Value: "META"},
	}
}
