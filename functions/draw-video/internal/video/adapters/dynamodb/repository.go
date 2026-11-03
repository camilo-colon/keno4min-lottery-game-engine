package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	video "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/video/domain"
	"github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/video/ports"
)

type Repository struct {
	client    *dynamodb.Client
	tableName string
}

var _ ports.Repository = (*Repository)(nil)

func NewRepository(client *dynamodb.Client, tableName string) *Repository {
	return &Repository{client: client, tableName: tableName}
}

func (r *Repository) GetAll(ctx context.Context) ([]video.Video, error) {
	input := &dynamodb.QueryInput{
		TableName:                 aws.String(r.tableName),
		KeyConditionExpression:    aws.String("#pk = :pk"),
		ExpressionAttributeNames:  map[string]string{"#pk": "PK"},
		ExpressionAttributeValues: map[string]types.AttributeValue{":pk": &types.AttributeValueMemberS{Value: "VIDEO"}},
	}
	videos := make([]video.Video, 0)
	for {
		page, err := r.client.Query(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("query videos: %w", err)
		}
		var items []videoItem
		if err := attributevalue.UnmarshalListOfMaps(page.Items, &items); err != nil {
			return nil, fmt.Errorf("decode videos: %w", err)
		}
		for _, item := range items {
			videos = append(videos, item.video())
		}
		if len(page.LastEvaluatedKey) == 0 {
			return videos, nil
		}
		input.ExclusiveStartKey = page.LastEvaluatedKey
	}
}
