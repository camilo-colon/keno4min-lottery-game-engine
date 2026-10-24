package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	draw "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/draw/domain"
	"github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/draw/ports"
)

type Repository struct {
	client    *dynamodb.Client
	tableName string
}

var _ ports.Repository = (*Repository)(nil)

func NewRepository(client *dynamodb.Client, tableName string) *Repository {
	return &Repository{client: client, tableName: tableName}
}

func (r *Repository) Update(ctx context.Context, d *draw.Draw) error {
	record, err := attributevalue.MarshalMap(toItem(d))
	if err != nil {
		return fmt.Errorf("marshal draw video: %w", err)
	}
	_, err = r.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			r.updateDraw(d.ID, record),
			r.updateRuntime(record),
		},
	})
	if err != nil {
		return fmt.Errorf("update draw video: %w", err)
	}
	return nil
}

func (r *Repository) updateDraw(drawID string, record map[string]types.AttributeValue) types.TransactWriteItem {
	return types.TransactWriteItem{Update: &types.Update{
		TableName:        aws.String(r.tableName),
		Key:              key("DRAW#" + drawID),
		UpdateExpression: aws.String("SET #balls = :balls, #videoId = :videoId, #gsi1pk = :gsi1pk, #gsi1sk = :gsi1sk"),
		ExpressionAttributeNames: map[string]string{
			"#balls": "balls", "#videoId": "videoId",
			"#gsi1pk": "GSI1PK",
			"#gsi1sk": "GSI1SK",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":balls": record["balls"], ":videoId": record["videoId"],
			":gsi1pk": &types.AttributeValueMemberS{Value: "DRAW#DRAWN"},
			":gsi1sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("DRAW#%s", drawID)},
		},
	}}
}

func (r *Repository) updateRuntime(record map[string]types.AttributeValue) types.TransactWriteItem {
	return types.TransactWriteItem{Update: &types.Update{
		TableName:        aws.String(r.tableName),
		Key:              key("DRAW_RUNTIME"),
		UpdateExpression: aws.String("SET #latestResult = :result"),
		ExpressionAttributeNames: map[string]string{
			"#latestResult": "latestResult",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":result": &types.AttributeValueMemberM{Value: record},
		},
	}}
}
