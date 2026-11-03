package dynamodb

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	draw "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/draw/domain"
)

func TestDynamoDBIntegration(t *testing.T) {
	if os.Getenv("DYNAMODB_INTEGRATION") != "1" {
		t.Skip("set DYNAMODB_INTEGRATION=1 with DynamoDB Local at 127.0.0.1:18074")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	client := dynamodb.New(dynamodb.Options{
		Region: "us-east-1", BaseEndpoint: aws.String("http://127.0.0.1:18074"),
		Credentials: aws.CredentialsProviderFunc(func(context.Context) (aws.Credentials, error) {
			return aws.Credentials{AccessKeyID: "fake", SecretAccessKey: "fake"}, nil
		}),
	})
	tableName := fmt.Sprintf("draw-video-test-%d", time.Now().UnixNano())
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName), BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("PK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("SK"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("PK"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("SK"), KeyType: types.KeyTypeRange},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := client.DeleteTable(cleanupCtx, &dynamodb.DeleteTableInput{TableName: aws.String(tableName)}); err != nil {
			t.Errorf("cleanup: %v", err)
		}
	})
	if err := dynamodb.NewTableExistsWaiter(client).Wait(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(tableName)}, 5*time.Second); err != nil {
		t.Fatal(err)
	}
	put := func(values map[string]any) {
		t.Helper()
		item, err := attributevalue.MarshalMap(values)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(tableName), Item: item}); err != nil {
			t.Fatal(err)
		}
	}
	get := func(pk string) map[string]types.AttributeValue {
		t.Helper()
		result, err := client.GetItem(ctx, &dynamodb.GetItemInput{TableName: aws.String(tableName), Key: drawKey(pk), ConsistentRead: aws.Bool(true)})
		if err != nil {
			t.Fatal(err)
		}
		return result.Item
	}
	put(map[string]any{
		"PK": "DRAW_RUNTIME", "SK": "META", "marker": "preserved",
		"current": map[string]any{"id": "future-draw", "tir": uint64(100), "status": "open", "startsAt": "2026-10-24T12:00:00Z", "closesAt": "2026-10-24T12:02:37Z"},
	})
	originalCurrent := get("DRAW_RUNTIME")["current"]
	for tir := 1; tir <= 3; tir++ {
		put(map[string]any{"PK": "DRAW#draw-" + strconv.Itoa(tir), "SK": "META", "id": "draw-" + strconv.Itoa(tir), "tir": tir, "status": "open", "startsAt": "2026-10-24T12:00:00Z"})
	}
	repository := NewRepository(client, tableName)
	checkResult := func(drawID string, tir uint64, balls []int, videoID string) {
		t.Helper()
		runtime := get("DRAW_RUNTIME")
		if !reflect.DeepEqual(runtime["current"], originalCurrent) || !reflect.DeepEqual(runtime["marker"], &types.AttributeValueMemberS{Value: "preserved"}) {
			t.Fatalf("current or runtime metadata changed: %+v", runtime)
		}
		var result struct {
			DrawID  string `dynamodbav:"drawId"`
			TIR     uint64 `dynamodbav:"tir"`
			Balls   []int  `dynamodbav:"balls"`
			VideoID string `dynamodbav:"videoId"`
		}
		if err := attributevalue.Unmarshal(runtime["latestResult"], &result); err != nil {
			t.Fatal(err)
		}
		if result.DrawID != drawID || result.TIR != tir || result.VideoID != videoID || !reflect.DeepEqual(result.Balls, balls) {
			t.Fatalf("unexpected latestResult: %+v", result)
		}
		canonical := get("DRAW#" + drawID)
		if !reflect.DeepEqual(canonical["GSI1PK"], &types.AttributeValueMemberS{Value: "DRAW#DRAWN"}) ||
			!reflect.DeepEqual(canonical["GSI1SK"], &types.AttributeValueMemberS{Value: "DRAW#" + drawID}) {
			t.Fatalf("unexpected drawn index keys: %+v", canonical)
		}
		latest := runtime["latestResult"].(*types.AttributeValueMemberM).Value
		for _, field := range []string{"tir", "balls", "videoId"} {
			if !reflect.DeepEqual(canonical[field], latest[field]) {
				t.Fatalf("draw and runtime disagree on %s", field)
			}
		}
		if !reflect.DeepEqual(canonical["status"], &types.AttributeValueMemberS{Value: draw.StatusDrawn}) {
			t.Fatal("draw status was not marked as drawn")
		}
	}
	if err := repository.Update(ctx, &draw.Draw{ID: "draw-2", TIR: 2, Status: draw.StatusDrawn, Balls: []int{63, 80}, VideoID: "0007"}); err != nil {
		t.Fatal(err)
	}
	checkResult("draw-2", 2, []int{63, 80}, "0007")
	if err := repository.Update(ctx, &draw.Draw{ID: "draw-3", TIR: 3, Status: draw.StatusDrawn, Balls: []int{37, 31}, VideoID: "3573"}); err != nil {
		t.Fatal(err)
	}
	checkResult("draw-3", 3, []int{37, 31}, "3573")
	// Runtime publication is unconditional, including when the TIR is older.
	if err := repository.Update(ctx, &draw.Draw{ID: "draw-1", TIR: 1, Status: draw.StatusDrawn, Balls: []int{68, 57}, VideoID: "0001"}); err != nil {
		t.Fatal(err)
	}
	checkResult("draw-1", 1, []int{68, 57}, "0001")
	// Replaying the current TIR keeps both records consistent.
	if err := repository.Update(ctx, &draw.Draw{ID: "draw-3", TIR: 3, Status: draw.StatusDrawn, Balls: []int{70, 79}, VideoID: "0008"}); err != nil {
		t.Fatal(err)
	}
	checkResult("draw-3", 3, []int{70, 79}, "0008")
	if err := repository.Update(ctx, &draw.Draw{ID: "missing", TIR: 4, Status: draw.StatusDrawn, Balls: []int{1, 2}, VideoID: "0009"}); err != nil {
		t.Fatal(err)
	}
	wantCreated, err := attributevalue.MarshalMap(map[string]any{
		"PK": "DRAW#missing", "SK": "META", "balls": []int{1, 2}, "videoId": "0009", "status": draw.StatusDrawn,
		"GSI1PK": "DRAW#DRAWN", "GSI1SK": "DRAW#missing",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(get("DRAW#missing"), wantCreated) {
		t.Fatal("missing draw was not created with balls, videoId and drawn index keys")
	}
	beforeDraw := get("DRAW#draw-2")
	if err := repository.Update(ctx, &draw.Draw{ID: "draw-2", TIR: 4, Status: draw.StatusDrawn, Balls: []int{1, 2}, VideoID: "0009"}); err != nil {
		t.Fatal(err)
	}
	afterDraw := get("DRAW#draw-2")
	runtime := get("DRAW_RUNTIME")
	latestWithDifferentTIR := runtime["latestResult"].(*types.AttributeValueMemberM).Value
	if !reflect.DeepEqual(beforeDraw["tir"], afterDraw["tir"]) ||
		!reflect.DeepEqual(latestWithDifferentTIR["tir"], &types.AttributeValueMemberN{Value: "4"}) ||
		!reflect.DeepEqual(runtime["current"], originalCurrent) {
		t.Fatal("update must preserve the draw TIR and publish the received TIR")
	}
	for _, field := range []string{"balls", "videoId"} {
		if !reflect.DeepEqual(afterDraw[field], latestWithDifferentTIR[field]) {
			t.Fatalf("draw and runtime disagree on %s", field)
		}
	}
	if _, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{TableName: aws.String(tableName), Key: drawKey("DRAW_RUNTIME")}); err != nil {
		t.Fatal(err)
	}
	if err := repository.Update(ctx, &draw.Draw{ID: "draw-2", TIR: 2, Status: draw.StatusDrawn, Balls: []int{1, 2}, VideoID: "0009"}); err != nil {
		t.Fatal(err)
	}
	recreated := get("DRAW_RUNTIME")
	latest := recreated["latestResult"].(*types.AttributeValueMemberM).Value
	updatedDraw := get("DRAW#draw-2")
	for _, field := range []string{"tir", "balls", "videoId"} {
		if !reflect.DeepEqual(updatedDraw[field], latest[field]) {
			t.Fatalf("recreated runtime and draw disagree on %s", field)
		}
	}
	if !reflect.DeepEqual(latest["videoId"], &types.AttributeValueMemberS{Value: "0009"}) {
		t.Fatal("missing runtime was not recreated with the new result")
	}
}
