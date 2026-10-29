package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cronos/keno4min-draw-engine/functions/create-draw/internal/draw"
)

type stubClient struct {
	get    func(context.Context, *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	update func(context.Context, *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error)
	write  func(context.Context, *dynamodb.TransactWriteItemsInput) (*dynamodb.TransactWriteItemsOutput, error)
}

func (s stubClient) GetItem(ctx context.Context, in *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return s.get(ctx, in)
}
func (s stubClient) UpdateItem(ctx context.Context, in *dynamodb.UpdateItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return s.update(ctx, in)
}
func (s stubClient) TransactWriteItems(ctx context.Context, in *dynamodb.TransactWriteItemsInput, _ ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error) {
	return s.write(ctx, in)
}

func TestNextTIRValidatesReturnedCounter(t *testing.T) {
	for _, tc := range []struct {
		value string
		valid bool
	}{
		{"1", true},
		{"42", true},
		{"9007199254740992", true},
		{"18446744073709551615", true},
		{"18446744073709551616", false},
		{"", false},
		{"0", false},
		{"-1", false},
		{"1.5", false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			client := stubClient{update: func(_ context.Context, in *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
				if aws.ToString(in.TableName) != "table" || !reflect.DeepEqual(in.Key, key("COUNTER#DRAW_TIR")) ||
					aws.ToString(in.UpdateExpression) != "ADD #value :one" || in.ReturnValues != types.ReturnValueUpdatedNew || in.ConditionExpression == nil {
					t.Fatalf("unsafe counter update: %+v", in)
				}
				if in.ExpressionAttributeValues[":max"].(*types.AttributeValueMemberN).Value != strconv.FormatUint(math.MaxUint64, 10) {
					t.Fatal("counter must allow the full uint64 range")
				}
				attributes := map[string]types.AttributeValue{}
				if tc.value != "" {
					attributes["value"] = &types.AttributeValueMemberN{Value: tc.value}
				}
				return &dynamodb.UpdateItemOutput{Attributes: attributes}, nil
			}}
			tir, err := NewRepository(client, "table").NextTIR(context.Background())
			want, _ := strconv.ParseUint(tc.value, 10, 64)
			if tc.valid && (err != nil || tir != want) || !tc.valid && (err == nil || tir != 0) {
				t.Fatalf("tir=%d, err=%v", tir, err)
			}
		})
	}
	failure := errors.New("counter unavailable")
	r := NewRepository(stubClient{update: func(context.Context, *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
		return nil, failure
	}}, "table")
	_, err := r.NextTIR(context.Background())
	if !errors.Is(err, failure) {
		t.Fatalf("lost error: %v", err)
	}
}

func TestFindByRequestReadsExistingTypeScriptItems(t *testing.T) {
	calls := 0
	client := stubClient{get: func(_ context.Context, in *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
		calls++
		if !aws.ToBool(in.ConsistentRead) {
			t.Fatal("idempotency requires consistent reads")
		}
		if calls == 1 {
			if !reflect.DeepEqual(in.Key, key("DRAW_REQUEST#request")) {
				t.Fatalf("unexpected key: %v", in.Key)
			}
			return &dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "draw-1"}}}, nil
		}
		if !reflect.DeepEqual(in.Key, key("DRAW#draw-1")) {
			t.Fatalf("unexpected key: %v", in.Key)
		}
		item, err := attributevalue.MarshalMap(map[string]any{
			"id": "draw-1", "tir": 12, "status": "drawn",
			"startsAt": "2026-09-12T15:00:00.123Z", "closesAt": "2026-09-12T15:02:37.123Z",
			"balls": []int{1, 2}, "videoId": "video",
		})
		return &dynamodb.GetItemOutput{Item: item}, err
	}}
	d, err := NewRepository(client, "table").FindByRequestID(context.Background(), "request")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || d.ID != "draw-1" || d.TIR != 12 || d.StartsAt.Nanosecond() != 123000000 || len(d.Balls) != 2 || d.VideoID != "video" {
		t.Fatalf("unexpected draw: %+v", d)
	}
}

func TestFindByRequestRejectsDanglingReference(t *testing.T) {
	calls := 0
	client := stubClient{get: func(context.Context, *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
		calls++
		if calls == 1 {
			return &dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "missing"}}}, nil
		}
		return &dynamodb.GetItemOutput{}, nil
	}}
	d, err := NewRepository(client, "table").FindByRequestID(context.Background(), "request")
	if d != nil || err == nil {
		t.Fatalf("dangling request silently accepted: %+v, %v", d, err)
	}
}

func TestSaveUsesOneTransactionAndPreservesStoredShape(t *testing.T) {
	d := draw.New("draw-1", 42, time.Date(2026, 10, 21, 15, 0, 0, 123000000, time.UTC))
	client := stubClient{write: func(_ context.Context, in *dynamodb.TransactWriteItemsInput) (*dynamodb.TransactWriteItemsOutput, error) {
		if len(in.TransactItems) != 4 {
			t.Fatalf("expected four atomic writes, got %d", len(in.TransactItems))
		}
		canonical := in.TransactItems[0].Put
		for _, field := range []string{"balls", "videoId", "GSI1PK", "GSI1SK"} {
			if _, ok := canonical.Item[field]; ok {
				t.Fatalf("unexpected field %s", field)
			}
		}
		var item drawItem
		if err := attributevalue.UnmarshalMap(canonical.Item, &item); err != nil {
			t.Fatal(err)
		}
		if item.PK != "DRAW#draw-1" || item.SK != "META" || !reflect.DeepEqual(item.draw(), d) {
			t.Fatalf("unexpected canonical: %+v", item)
		}
		lookup := in.TransactItems[1].Put
		if lookup.Item["PK"].(*types.AttributeValueMemberS).Value != "DRAW_TIR#42" {
			t.Fatalf("unexpected lookup: %+v", lookup)
		}
		runtime := in.TransactItems[2].Update
		if aws.ToString(runtime.UpdateExpression) != "SET #current = :current" || aws.ToString(runtime.ConditionExpression) != "attribute_not_exists(#current) OR #current.#tir < :tir" {
			t.Fatalf("unsafe runtime update: %+v", runtime)
		}
		current := runtime.ExpressionAttributeValues[":current"].(*types.AttributeValueMemberM).Value
		if _, exists := current["PK"]; exists {
			t.Fatal("runtime contains storage keys")
		}
		for _, write := range in.TransactItems {
			if write.Put != nil && aws.ToString(write.Put.ConditionExpression) != "attribute_not_exists(PK)" {
				t.Fatal("write can overwrite existing data")
			}
		}
		return &dynamodb.TransactWriteItemsOutput{}, nil
	}}
	if err := NewRepository(client, "table").Save(context.Background(), "request", d); err != nil {
		t.Fatal(err)
	}
}

func TestSaveOnlyMapsConcurrencyErrorsToRetryableType(t *testing.T) {
	for _, tc := range []struct {
		name     string
		err      error
		conflict bool
	}{
		{"success", nil, false},
		{"conflict", &types.TransactionConflictException{}, true},
		{"conditional", cancellation("None", "ConditionalCheckFailed"), true},
		{"wrapped conflict", fmt.Errorf("SDK: %w", cancellation("TransactionConflict")), true},
		{"capacity", cancellation("ProvisionedThroughputExceeded"), false},
		{"mixed", cancellation("ConditionalCheckFailed", "ValidationError"), false},
		{"no reasons", cancellation(), false},
		{"other", errors.New("network"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRepository(stubClient{write: func(context.Context, *dynamodb.TransactWriteItemsInput) (*dynamodb.TransactWriteItemsOutput, error) {
				return nil, tc.err
			}}, "table")
			err := r.Save(context.Background(), "request", draw.New("id", 1, time.Now()))
			var conflict *draw.DrawConflictError
			if errors.As(err, &conflict) != tc.conflict {
				t.Fatalf("unexpected classification: %T %v", err, err)
			}
			if tc.conflict && reflect.TypeOf(err).Elem().Name() != "DrawConflictError" {
				t.Fatalf("Step Functions will not retry: %T", err)
			}
			if !tc.conflict && err != tc.err {
				t.Fatalf("non-conflict error changed: %v", err)
			}
		})
	}
}

func cancellation(codes ...string) error {
	reasons := make([]types.CancellationReason, len(codes))
	for i, code := range codes {
		reasons[i].Code = aws.String(code)
	}
	return &types.TransactionCanceledException{CancellationReasons: reasons}
}
