package dynamodb

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go/middleware"
	draw "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/draw/domain"
)

func newDrawTestClient(call func(context.Context, any) (any, error)) *dynamodb.Client {
	return dynamodb.New(dynamodb.Options{
		Region: "us-east-1",
		APIOptions: []func(*middleware.Stack) error{
			func(stack *middleware.Stack) error {
				return stack.Initialize.Add(middleware.InitializeMiddlewareFunc("testDrawWrite", func(ctx context.Context, input middleware.InitializeInput, _ middleware.InitializeHandler) (middleware.InitializeOutput, middleware.Metadata, error) {
					result, err := call(ctx, input.Parameters)
					return middleware.InitializeOutput{Result: result}, middleware.Metadata{}, err
				}), middleware.Before)
			},
		},
	})
}

func TestUpdateWritesDrawAndRuntimeTogether(t *testing.T) {
	type contextKey struct{}
	ctx := context.WithValue(t.Context(), contextKey{}, "request")
	balls := []int{63, 80, 37, 31, 68, 57, 70, 79, 60, 33, 47, 66, 58, 55, 18, 64, 74, 59, 22, 75}
	calls := 0
	client := newDrawTestClient(func(gotCtx context.Context, input any) (any, error) {
		calls++
		transaction, ok := input.(*dynamodb.TransactWriteItemsInput)
		if !ok || len(transaction.TransactItems) != 2 || gotCtx.Value(contextKey{}) != "request" {
			t.Fatalf("expected a two-item transaction: %+v", input)
		}
		drawUpdate := transaction.TransactItems[0].Update
		runtimeUpdate := transaction.TransactItems[1].Update
		if aws.ToString(drawUpdate.TableName) != "table" || aws.ToString(runtimeUpdate.TableName) != "table" ||
			!reflect.DeepEqual(drawUpdate.Key, drawKey("DRAW#draw-1")) || !reflect.DeepEqual(runtimeUpdate.Key, drawKey("DRAW_RUNTIME")) {
			t.Fatalf("unexpected transaction targets: %+v", transaction)
		}
		if aws.ToString(drawUpdate.UpdateExpression) != "SET #balls = :balls, #videoId = :videoId, #status = :status, #gsi1pk = :gsi1pk, #gsi1sk = :gsi1sk" ||
			drawUpdate.ConditionExpression != nil ||
			!reflect.DeepEqual(drawUpdate.ExpressionAttributeNames, map[string]string{
				"#balls": "balls", "#videoId": "videoId", "#status": "status", "#gsi1pk": "GSI1PK", "#gsi1sk": "GSI1SK",
			}) ||
			len(drawUpdate.ExpressionAttributeValues) != 5 {
			t.Fatalf("unsafe draw update: %+v", drawUpdate)
		}
		if !reflect.DeepEqual(drawUpdate.ExpressionAttributeValues[":status"], &types.AttributeValueMemberS{Value: draw.StatusDrawn}) {
			t.Fatalf("unexpected draw status: %+v", drawUpdate.ExpressionAttributeValues)
		}
		if !reflect.DeepEqual(drawUpdate.ExpressionAttributeValues[":gsi1pk"], &types.AttributeValueMemberS{Value: "DRAW#DRAWN"}) ||
			!reflect.DeepEqual(drawUpdate.ExpressionAttributeValues[":gsi1sk"], &types.AttributeValueMemberS{Value: "DRAW#draw-1"}) {
			t.Fatalf("unexpected drawn index keys: %+v", drawUpdate.ExpressionAttributeValues)
		}
		if aws.ToString(runtimeUpdate.UpdateExpression) != "SET #latestResult = :result" ||
			runtimeUpdate.ConditionExpression != nil ||
			!reflect.DeepEqual(runtimeUpdate.ExpressionAttributeNames, map[string]string{"#latestResult": "latestResult"}) ||
			len(runtimeUpdate.ExpressionAttributeValues) != 1 {
			t.Fatalf("unsafe runtime update: %+v", runtimeUpdate)
		}
		result := runtimeUpdate.ExpressionAttributeValues[":result"].(*types.AttributeValueMemberM).Value
		var decoded struct {
			DrawID  string `dynamodbav:"drawId"`
			TIR     uint64 `dynamodbav:"tir"`
			Balls   []int  `dynamodbav:"balls"`
			VideoID string `dynamodbav:"videoId"`
		}
		if err := attributevalue.UnmarshalMap(result, &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded.DrawID != "draw-1" || decoded.TIR != math.MaxUint64 || decoded.VideoID != "0007" || !reflect.DeepEqual(decoded.Balls, balls) || len(result) != 4 {
			t.Fatalf("unexpected latestResult: %+v", decoded)
		}
		for _, field := range []string{"balls", "videoId"} {
			if !reflect.DeepEqual(drawUpdate.ExpressionAttributeValues[":"+field], result[field]) {
				t.Fatalf("draw and runtime disagree on %s", field)
			}
		}
		return &dynamodb.TransactWriteItemsOutput{}, nil
	})
	if err := NewRepository(client, "table").Update(ctx, &draw.Draw{ID: "draw-1", TIR: math.MaxUint64, Status: draw.StatusDrawn, Balls: balls, VideoID: "0007"}); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("expected one transaction, got %d calls", calls)
	}
}

func drawKey(pk string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{"PK": &types.AttributeValueMemberS{Value: pk}, "SK": &types.AttributeValueMemberS{Value: "META"}}
}

func canceledWrite(drawCode, runtimeCode string) *types.TransactionCanceledException {
	return &types.TransactionCanceledException{CancellationReasons: []types.CancellationReason{
		{Code: aws.String(drawCode)}, {Code: aws.String(runtimeCode)},
	}}
}

func TestUpdatePropagatesErrorsWithoutAnotherWrite(t *testing.T) {
	for _, failure := range []error{
		errors.New("DynamoDB unavailable"), context.Canceled,
		canceledWrite("TransactionConflict", "None"),
		canceledWrite("None", "ProvisionedThroughputExceeded"),
	} {
		calls := 0
		client := newDrawTestClient(func(_ context.Context, input any) (any, error) {
			calls++
			if _, ok := input.(*dynamodb.TransactWriteItemsInput); !ok {
				t.Fatalf("unexpected additional write: %+v", input)
			}
			return nil, failure
		})
		err := NewRepository(client, "table").Update(t.Context(), &draw.Draw{ID: "draw-1", TIR: 42, Balls: []int{63, 80}, VideoID: "3573"})
		if calls != 1 || !errors.Is(err, failure) {
			t.Fatalf("lost failure %v: %v; calls=%d", failure, err, calls)
		}
	}
}
