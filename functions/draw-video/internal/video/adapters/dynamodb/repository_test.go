package dynamodb

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go/middleware"
	video "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/video/domain"
)

func newVideoTestClient(query func(context.Context, *dynamodb.QueryInput) (*dynamodb.QueryOutput, error)) *dynamodb.Client {
	return dynamodb.New(dynamodb.Options{
		Region: "us-east-1",
		APIOptions: []func(*middleware.Stack) error{
			func(stack *middleware.Stack) error {
				return stack.Initialize.Add(middleware.InitializeMiddlewareFunc("testQuery", func(ctx context.Context, input middleware.InitializeInput, _ middleware.InitializeHandler) (middleware.InitializeOutput, middleware.Metadata, error) {
					result, err := query(ctx, input.Parameters.(*dynamodb.QueryInput))
					return middleware.InitializeOutput{Result: result}, middleware.Metadata{}, err
				}), middleware.Before)
			},
		},
	})
}

func TestGetAllReadsEveryPageAndPreservesBallOrder(t *testing.T) {
	type contextKey struct{}
	ctx := context.WithValue(t.Context(), contextKey{}, "request")
	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: "VIDEO"},
		"SK": &types.AttributeValueMemberS{Value: "0007"},
	}
	want := []video.Video{
		{ID: "0007", Balls: []int{63, 80, 37, 31, 68, 57, 70, 79, 60, 33, 47, 66, 58, 55, 18, 64, 74, 59, 22, 75}},
		{ID: "3573", Balls: []int{80, 1}},
	}
	items := make([]map[string]types.AttributeValue, len(want))
	for i, v := range want {
		item, err := attributevalue.MarshalMap(map[string]any{
			"PK": "VIDEO", "SK": v.ID, "id": v.ID, "balls": v.Balls,
		})
		if err != nil {
			t.Fatal(err)
		}
		items[i] = item
	}
	calls := 0
	client := newVideoTestClient(func(gotCtx context.Context, input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
		if gotCtx.Value(contextKey{}) != "request" || aws.ToString(input.TableName) != "table" ||
			aws.ToString(input.KeyConditionExpression) != "#pk = :pk" || input.ExpressionAttributeNames["#pk"] != "PK" ||
			!reflect.DeepEqual(input.ExpressionAttributeValues[":pk"], key["PK"]) || input.IndexName != nil {
			t.Fatalf("unexpected query: %+v", input)
		}
		calls++
		switch calls {
		case 1:
			if len(input.ExclusiveStartKey) != 0 {
				t.Fatal("first page must start without a key")
			}
			return &dynamodb.QueryOutput{Items: items[:1], LastEvaluatedKey: key}, nil
		case 2:
			if !reflect.DeepEqual(input.ExclusiveStartKey, key) {
				t.Fatal("missing pagination key")
			}
			return &dynamodb.QueryOutput{Items: items[1:], LastEvaluatedKey: map[string]types.AttributeValue{}}, nil
		default:
			t.Fatal("unexpected extra page")
			return nil, nil
		}
	})
	got, err := NewRepository(client, "table").GetAll(ctx)
	if err != nil || calls != 2 || !reflect.DeepEqual(got, want) {
		t.Fatalf("GetAll() = %+v, %v; calls=%d", got, err, calls)
	}
}

func TestGetAllReturnsEmptySlice(t *testing.T) {
	r := NewRepository(newVideoTestClient(func(context.Context, *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
		return &dynamodb.QueryOutput{}, nil
	}), "table")
	videos, err := r.GetAll(t.Context())
	if err != nil || videos == nil || len(videos) != 0 {
		t.Fatalf("GetAll() = %+v, %v", videos, err)
	}
}

func TestGetAllPropagatesQueryErrorWithoutPartialResults(t *testing.T) {
	failure := errors.New("query failed")
	calls := 0
	r := NewRepository(newVideoTestClient(func(context.Context, *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
		calls++
		if calls == 1 {
			return &dynamodb.QueryOutput{
				Items:            []map[string]types.AttributeValue{{"id": &types.AttributeValueMemberS{Value: "0001"}}},
				LastEvaluatedKey: map[string]types.AttributeValue{"PK": &types.AttributeValueMemberS{Value: "VIDEO"}, "SK": &types.AttributeValueMemberS{Value: "0001"}},
			}, nil
		}
		return nil, failure
	}), "table")
	videos, err := r.GetAll(t.Context())
	if videos != nil || !errors.Is(err, failure) || calls != 2 {
		t.Fatalf("GetAll() = %+v, %v; calls=%d", videos, err, calls)
	}
}

func TestGetAllRejectsInvalidBalls(t *testing.T) {
	r := NewRepository(newVideoTestClient(func(context.Context, *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
		return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{{
			"id": &types.AttributeValueMemberS{Value: "3573"}, "balls": &types.AttributeValueMemberS{Value: "invalid"},
		}}}, nil
	}), "table")
	videos, err := r.GetAll(t.Context())
	if videos != nil || err == nil {
		t.Fatalf("GetAll() = %+v, %v", videos, err)
	}
}
