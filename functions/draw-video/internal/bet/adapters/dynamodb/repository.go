package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	bet "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bet/domain"
	"github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bet/ports"
)

type Repository struct {
	client    *dynamodb.Client
	tableName string
}

var _ ports.Repository = (*Repository)(nil)

func NewRepository(client *dynamodb.Client, tableName string) *Repository {
	return &Repository{client: client, tableName: tableName}
}

func (r *Repository) GetByTIR(ctx context.Context, tir uint64) ([]bet.AdminBets, error) {
	input := &dynamodb.QueryInput{
		TableName:                 aws.String(r.tableName),
		ConsistentRead:            aws.Bool(true),
		KeyConditionExpression:    aws.String("#pk = :pk"),
		ExpressionAttributeNames:  map[string]string{"#pk": "PK"},
		ExpressionAttributeValues: map[string]types.AttributeValue{":pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("BETS#%d", tir)}},
	}
	var betsByAdmin []bet.AdminBets
	for {
		page, err := r.client.Query(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("query bets: %w", err)
		}
		var items []betsItem
		if err := attributevalue.UnmarshalListOfMaps(page.Items, &items); err != nil {
			return nil, fmt.Errorf("decode bets items: %w", err)
		}
		for _, item := range items {
			adminNodeID, err := item.adminNodeID()
			if err != nil {
				return nil, err
			}
			bets, err := decodeBets(item.Bets)
			if err != nil {
				return nil, fmt.Errorf("bets item %q: %w", item.SK, err)
			}
			betsByAdmin = appendAdminBets(betsByAdmin, adminNodeID, bets)
		}
		if len(page.LastEvaluatedKey) == 0 {
			return betsByAdmin, nil
		}
		input.ExclusiveStartKey = page.LastEvaluatedKey
	}
}

func appendAdminBets(betsByAdmin []bet.AdminBets, adminNodeID string, bets []bet.Bet) []bet.AdminBets {
	last := len(betsByAdmin) - 1
	if last >= 0 && betsByAdmin[last].AdminNodeID == adminNodeID {
		betsByAdmin[last].Bets = append(betsByAdmin[last].Bets, bets...)
		return betsByAdmin
	}
	return append(betsByAdmin, bet.AdminBets{AdminNodeID: adminNodeID, Bets: bets})
}
