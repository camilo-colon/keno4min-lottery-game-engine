package dynamodb

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	ticket "github.com/cronos/keno4min-draw-engine/functions/update-tickets/internal/ticket/domain"
	"github.com/cronos/keno4min-draw-engine/functions/update-tickets/internal/ticket/ports"
)

type Repository struct {
	client    *dynamodb.Client
	tableName string
}

var _ ports.Repository = (*Repository)(nil)

func NewRepository(client *dynamodb.Client, tableName string) *Repository {
	return &Repository{client: client, tableName: tableName}
}

func (r *Repository) ListByTIR(ctx context.Context, tir uint64) ([]ticket.Ticket, error) {
	input := &dynamodb.QueryInput{
		TableName:                aws.String(r.tableName),
		IndexName:                aws.String("GSI1"),
		KeyConditionExpression:   aws.String("#pk = :pk"),
		ExpressionAttributeNames: map[string]string{"#pk": "GSI1PK"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "TICKETS#TIR#" + strconv.FormatUint(tir, 10)},
		},
	}
	tickets := make([]ticket.Ticket, 0)
	for {
		page, err := r.client.Query(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("query tickets: %w", err)
		}
		var items []ticketItem
		if err := attributevalue.UnmarshalListOfMaps(page.Items, &items); err != nil {
			return nil, fmt.Errorf("decode tickets: %w", err)
		}
		for _, item := range items {
			tickets = append(tickets, item.ticket())
		}
		if len(page.LastEvaluatedKey) == 0 {
			return tickets, nil
		}
		input.ExclusiveStartKey = page.LastEvaluatedKey
	}
}

// Update guarda el resultado y las balotas solo si el ticket sigue pendiente,
// para no pisar un ticket cancelado o ya liquidado.
func (r *Repository) Update(ctx context.Context, t *ticket.Ticket) error {
	balls, err := attributevalue.Marshal(t.Balls)
	if err != nil {
		return fmt.Errorf("marshal ticket balls: %w", err)
	}
	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:           aws.String(r.tableName),
		Key:                 key(t.ID),
		UpdateExpression:    aws.String("SET #status = :status, #prizeCents = :prizeCents, #balls = :balls"),
		ConditionExpression: aws.String("#status = :pending"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status", "#prizeCents": "prizeCents", "#balls": "balls",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status":     &types.AttributeValueMemberS{Value: string(t.Status)},
			":prizeCents": &types.AttributeValueMemberN{Value: strconv.FormatInt(t.PrizeCents, 10)},
			":balls":      balls,
			":pending":    &types.AttributeValueMemberS{Value: string(ticket.StatusPending)},
		},
	})
	if err != nil {
		return fmt.Errorf("update ticket: %w", err)
	}
	return nil
}
