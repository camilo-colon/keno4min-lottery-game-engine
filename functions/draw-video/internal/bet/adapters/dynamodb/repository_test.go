package dynamodb

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go/middleware"
	bet "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bet/domain"
	"github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bitmask"
)

const (
	firstAdmin  = "0b6f1c0e-5a43-4c1d-9a7e-1f2d3c4b5a69"
	secondAdmin = "7d2e9f4a-8b1c-4e6f-a3d5-9c8b7a6f5e4d"
)

func newBetsTestClient(query func(context.Context, *dynamodb.QueryInput) (*dynamodb.QueryOutput, error)) *dynamodb.Client {
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

func betsRecordItem(sortKey string, record []byte) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK":  &types.AttributeValueMemberS{Value: "BETS#42"},
		"SK":  &types.AttributeValueMemberS{Value: sortKey},
		"b":   &types.AttributeValueMemberB{Value: record},
		"ttl": &types.AttributeValueMemberN{Value: "1795000000"},
	}
}

func TestGetByTIRReadsEveryPageAndGroupsBetsByAdmin(t *testing.T) {
	type contextKey struct{}
	ctx := context.WithValue(t.Context(), contextKey{}, "request")
	twoNumbers := bet.Bet{Numbers: bitmask.MaskOf([]int{7, 70}), StakeCents: 500}
	threeNumbers := bet.Bet{Numbers: bitmask.MaskOf([]int{1, 2, 3}), StakeCents: 1000}
	fourNumbers := bet.Bet{Numbers: bitmask.MaskOf([]int{10, 20, 30, 40}), StakeCents: 200}
	oneNumber := bet.Bet{Numbers: bitmask.MaskOf([]int{80}), StakeCents: 300}
	items := []map[string]types.AttributeValue{
		betsRecordItem(firstAdmin+"#ticket-1", encodeBets([]bet.Bet{twoNumbers, threeNumbers})),
		betsRecordItem(firstAdmin+"#ticket-2", encodeBets([]bet.Bet{fourNumbers})),
		betsRecordItem(secondAdmin+"#ticket-3", encodeBets([]bet.Bet{oneNumber})),
	}
	want := []bet.AdminBets{
		{AdminNodeID: firstAdmin, Bets: []bet.Bet{twoNumbers, threeNumbers, fourNumbers}},
		{AdminNodeID: secondAdmin, Bets: []bet.Bet{oneNumber}},
	}
	partitionKey := &types.AttributeValueMemberS{Value: "BETS#42"}
	pageKey := map[string]types.AttributeValue{"PK": items[0]["PK"], "SK": items[0]["SK"]}
	calls := 0
	client := newBetsTestClient(func(gotCtx context.Context, input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
		if gotCtx.Value(contextKey{}) != "request" || aws.ToString(input.TableName) != "table" ||
			aws.ToString(input.KeyConditionExpression) != "#pk = :pk" || input.ExpressionAttributeNames["#pk"] != "PK" ||
			!reflect.DeepEqual(input.ExpressionAttributeValues[":pk"], partitionKey) ||
			!aws.ToBool(input.ConsistentRead) || input.IndexName != nil {
			t.Fatalf("unexpected query: %+v", input)
		}
		calls++
		switch calls {
		case 1:
			if len(input.ExclusiveStartKey) != 0 {
				t.Fatal("first page must start without a key")
			}
			return &dynamodb.QueryOutput{Items: items[:1], LastEvaluatedKey: pageKey}, nil
		case 2:
			if !reflect.DeepEqual(input.ExclusiveStartKey, pageKey) {
				t.Fatal("missing pagination key")
			}
			return &dynamodb.QueryOutput{Items: items[1:]}, nil
		default:
			t.Fatal("unexpected extra page")
			return nil, nil
		}
	})
	got, err := NewRepository(client, "table").GetByTIR(ctx, 42)
	if err != nil || calls != 2 || !reflect.DeepEqual(got, want) {
		t.Fatalf("GetByTIR() = %+v, %v; calls=%d", got, err, calls)
	}
}

func TestGetByTIRReturnsNoBetsForARoundWithoutSales(t *testing.T) {
	r := NewRepository(newBetsTestClient(func(context.Context, *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
		return &dynamodb.QueryOutput{}, nil
	}), "table")
	got, err := r.GetByTIR(t.Context(), 42)
	if err != nil || len(got) != 0 {
		t.Fatalf("GetByTIR() = %+v, %v", got, err)
	}
}

func TestGetByTIRPropagatesQueryErrorWithoutPartialResults(t *testing.T) {
	failure := errors.New("query failed")
	valid := encodeBets([]bet.Bet{{Numbers: bitmask.MaskOf([]int{7}), StakeCents: 500}})
	calls := 0
	r := NewRepository(newBetsTestClient(func(context.Context, *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
		calls++
		if calls == 1 {
			item := betsRecordItem(firstAdmin+"#ticket-1", valid)
			return &dynamodb.QueryOutput{
				Items:            []map[string]types.AttributeValue{item},
				LastEvaluatedKey: map[string]types.AttributeValue{"PK": item["PK"], "SK": item["SK"]},
			}, nil
		}
		return nil, failure
	}), "table")
	got, err := r.GetByTIR(t.Context(), 42)
	if got != nil || !errors.Is(err, failure) || calls != 2 {
		t.Fatalf("GetByTIR() = %+v, %v; calls=%d", got, err, calls)
	}
}

func TestGetByTIRRejectsMalformedItems(t *testing.T) {
	valid := encodeBets([]bet.Bet{{Numbers: bitmask.MaskOf([]int{7}), StakeCents: 500}})
	withoutBets := betsRecordItem(firstAdmin+"#ticket-1", valid)
	delete(withoutBets, "b")
	tests := []struct {
		name    string
		item    map[string]types.AttributeValue
		wantErr string
	}{
		{name: "sort key without separator", item: betsRecordItem(firstAdmin, valid), wantErr: "has no admin node id"},
		{name: "sort key without admin", item: betsRecordItem("#ticket-1", valid), wantErr: "has no admin node id"},
		{name: "truncated bets record", item: betsRecordItem(firstAdmin+"#ticket-1", valid[:23]), wantErr: "bets record has 23 bytes"},
		{name: "missing bets record", item: withoutBets, wantErr: "bets record has 0 bytes"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRepository(newBetsTestClient(func(context.Context, *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
				return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{tt.item}}, nil
			}), "table")
			got, err := r.GetByTIR(t.Context(), 42)
			if got != nil || err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("GetByTIR() = %+v, %v; want error containing %q", got, err, tt.wantErr)
			}
		})
	}
}
