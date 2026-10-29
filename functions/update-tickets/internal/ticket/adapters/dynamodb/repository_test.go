package dynamodb

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go/middleware"
	ticket "github.com/cronos/keno4min-draw-engine/functions/update-tickets/internal/ticket/domain"
)

func TestUpdateStoresResultAndBallsWhilePending(t *testing.T) {
	calls := 0
	client := dynamodb.New(dynamodb.Options{
		Region: "us-east-1",
		APIOptions: []func(*middleware.Stack) error{
			func(stack *middleware.Stack) error {
				return stack.Initialize.Add(middleware.InitializeMiddlewareFunc("testTicketWrite", func(_ context.Context, input middleware.InitializeInput, _ middleware.InitializeHandler) (middleware.InitializeOutput, middleware.Metadata, error) {
					calls++
					update, ok := input.Parameters.(*dynamodb.UpdateItemInput)
					if !ok {
						t.Fatalf("unexpected request: %+v", input.Parameters)
					}
					if aws.ToString(update.UpdateExpression) != "SET #status = :status, #prizeCents = :prizeCents, #balls = :balls" ||
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
					return middleware.InitializeOutput{Result: &dynamodb.UpdateItemOutput{}}, middleware.Metadata{}, nil
				}), middleware.Before)
			},
		},
	})
	err := NewRepository(client, "table").Update(t.Context(), &ticket.Ticket{
		ID: "ticket-1", Status: ticket.StatusLost, Balls: []int{61, 8},
	})
	if err != nil || calls != 1 {
		t.Fatalf("Update() = %v; calls=%d", err, calls)
	}
}
