package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cronos/keno4min-draw-engine/functions/create-draw/internal/draw"
)

type client interface {
	GetItem(context.Context, *dynamodb.GetItemInput, ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(context.Context, *dynamodb.UpdateItemInput, ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	TransactWriteItems(context.Context, *dynamodb.TransactWriteItemsInput, ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
}

type Repository struct {
	client    client
	tableName string
}

var _ draw.Repository = (*Repository)(nil)

func NewRepository(client client, tableName string) *Repository {
	return &Repository{client: client, tableName: tableName}
}

func (r *Repository) NextTIR(ctx context.Context) (uint64, error) {
	result, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                aws.String(r.tableName),
		Key:                      key("COUNTER#DRAW_TIR"),
		UpdateExpression:         aws.String("ADD #value :one"),
		ConditionExpression:      aws.String("attribute_not_exists(PK) OR (#value >= :zero AND #value < :max)"),
		ExpressionAttributeNames: map[string]string{"#value": "value"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":zero": &types.AttributeValueMemberN{Value: "0"},
			":one":  &types.AttributeValueMemberN{Value: "1"},
			":max":  &types.AttributeValueMemberN{Value: strconv.FormatUint(math.MaxUint64, 10)},
		},
		ReturnValues: types.ReturnValueUpdatedNew,
	})
	if err != nil {
		return 0, fmt.Errorf("increment draw counter: %w", err)
	}

	value, ok := result.Attributes["value"].(*types.AttributeValueMemberN)
	if !ok {
		return 0, errors.New("draw counter value is missing or invalid")
	}
	tir, err := strconv.ParseUint(value.Value, 10, 64)
	if err != nil || tir == 0 {
		return 0, errors.New("draw counter is outside the allowed range")
	}
	return tir, nil
}

func (r *Repository) FindByRequestID(ctx context.Context, requestID string) (*draw.Draw, error) {
	var request struct {
		ID string `dynamodbav:"id"`
	}
	found, err := r.get(ctx, "DRAW_REQUEST#"+requestID, &request)
	if err != nil || !found {
		return nil, err
	}
	if request.ID == "" {
		return nil, errors.New("request references a missing draw ID")
	}
	var item drawItem
	found, err = r.get(ctx, "DRAW#"+request.ID, &item)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.New("request references a missing draw")
	}
	return item.draw(), nil
}

func (r *Repository) Save(ctx context.Context, requestID string, d *draw.Draw) error {
	record, err := attributevalue.MarshalMap(toItem(d))
	if err != nil {
		return fmt.Errorf("marshal draw: %w", err)
	}
	tir := strconv.FormatUint(d.TIR, 10)
	lookup := map[string]types.AttributeValue{"id": record["id"]}
	request := map[string]types.AttributeValue{"id": record["id"], "tir": record["tir"]}

	_, err = r.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			r.insert("DRAW#"+d.ID, record),
			r.insert("DRAW_TIR#"+tir, lookup),
			r.updateView(record),
			r.insert("DRAW_REQUEST#"+requestID, request),
		},
	})
	if isConflict(err) {
		return &draw.DrawConflictError{Cause: err}
	}
	return err
}

func (r *Repository) get(ctx context.Context, pk string, target any) (bool, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName), Key: key(pk), ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return false, err
	}
	if len(result.Item) == 0 {
		return false, nil
	}
	if err := attributevalue.UnmarshalMap(result.Item, target); err != nil {
		return false, fmt.Errorf("decode %s: %w", pk, err)
	}
	return true, nil
}

func (r *Repository) insert(pk string, attributes map[string]types.AttributeValue) types.TransactWriteItem {
	item := key(pk)
	maps.Copy(item, attributes)
	return types.TransactWriteItem{Put: &types.Put{
		TableName: aws.String(r.tableName), Item: item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	}}
}

func (r *Repository) updateView(record map[string]types.AttributeValue) types.TransactWriteItem {
	return types.TransactWriteItem{Update: &types.Update{
		TableName:                aws.String(r.tableName),
		Key:                      key("DRAW_RUNTIME"),
		UpdateExpression:         aws.String("SET #current = :current"),
		ConditionExpression:      aws.String("attribute_not_exists(#current) OR #current.#tir < :tir"),
		ExpressionAttributeNames: map[string]string{"#current": "current", "#tir": "tir"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":current": &types.AttributeValueMemberM{Value: record},
			":tir":     record["tir"],
		},
	}}
}

func isConflict(err error) bool {
	if _, ok := errors.AsType[*types.TransactionConflictException](err); ok {
		return true
	}
	canceled, ok := errors.AsType[*types.TransactionCanceledException](err)
	if !ok {
		return false
	}
	failed := false
	for _, reason := range canceled.CancellationReasons {
		switch aws.ToString(reason.Code) {
		case "", "None":
		case "ConditionalCheckFailed", "TransactionConflict":
			failed = true
		default:
			return false
		}
	}
	return failed
}
