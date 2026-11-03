package dynamodb

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go/middleware"
	ticket "github.com/cronos/keno4min-draw-engine/functions/update-tickets/internal/ticket/domain"
)

func TestUpdateStoresResultAndOutboxInOneTransaction(t *testing.T) {
	calls := 0
	client := dynamodb.New(dynamodb.Options{
		Region: "us-east-1",
		APIOptions: []func(*middleware.Stack) error{
			func(stack *middleware.Stack) error {
				return stack.Initialize.Add(middleware.InitializeMiddlewareFunc("testTicketWrite", func(_ context.Context, input middleware.InitializeInput, _ middleware.InitializeHandler) (middleware.InitializeOutput, middleware.Metadata, error) {
					calls++
					transaction, ok := input.Parameters.(*dynamodb.TransactWriteItemsInput)
					if !ok {
						t.Fatalf("unexpected request: %+v", input.Parameters)
					}
					if len(transaction.TransactItems) != 2 {
						t.Fatalf("got %d transact items, want ticket update and outbox put", len(transaction.TransactItems))
					}

					update := transaction.TransactItems[0].Update
					if update == nil ||
						aws.ToString(update.TableName) != "table" ||
						aws.ToString(update.UpdateExpression) != "SET #status = :status, #prizeCents = :prizeCents, #balls = :balls" ||
						aws.ToString(update.ConditionExpression) != "#status = :pending" ||
						!reflect.DeepEqual(update.Key, key("ticket-1")) {
						t.Fatalf("unexpected update: %+v", update)
					}
					wantBalls := &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberN{Value: "61"}, &types.AttributeValueMemberN{Value: "8"},
					}}
					if update.ExpressionAttributeNames["#balls"] != "balls" ||
						!reflect.DeepEqual(update.ExpressionAttributeValues[":balls"], wantBalls) ||
						!reflect.DeepEqual(update.ExpressionAttributeValues[":status"], &types.AttributeValueMemberS{Value: "lost"}) ||
						!reflect.DeepEqual(update.ExpressionAttributeValues[":prizeCents"], &types.AttributeValueMemberN{Value: "0"}) {
						t.Fatalf("unexpected values: %+v", update.ExpressionAttributeValues)
					}

					put := transaction.TransactItems[1].Put
					if put == nil || aws.ToString(put.TableName) != "table" ||
						aws.ToString(put.ConditionExpression) != "attribute_not_exists(PK)" ||
						!reflect.DeepEqual(put.Item["eventType"], &types.AttributeValueMemberS{Value: "cronos.keno4min.ticket.settled"}) {
						t.Fatalf("unexpected outbox put: %+v", put)
					}
					return middleware.InitializeOutput{Result: &dynamodb.TransactWriteItemsOutput{}}, middleware.Metadata{}, nil
				}), middleware.Before)
			},
		},
	})
	settled := &ticket.Ticket{ID: "ticket-1", Status: ticket.StatusPending, Bets: []ticket.Bet{{Nums: []int{80}, ValueCents: 1000}}}
	settled.Settle([]int{61, 8}, time.Date(2026, 11, 2, 21, 30, 0, 0, time.UTC))

	err := NewRepository(client, "table").Update(t.Context(), settled)
	if err != nil || calls != 1 {
		t.Fatalf("Update() = %v; calls=%d", err, calls)
	}
}
