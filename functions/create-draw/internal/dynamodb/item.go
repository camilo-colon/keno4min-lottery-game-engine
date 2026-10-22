package dynamodb

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cronos/keno4min-draw-engine/functions/create-draw/internal/draw"
)

// Storage names match the existing TypeScript function.
type drawItem struct {
	PK       string    `dynamodbav:"PK,omitempty"`
	SK       string    `dynamodbav:"SK,omitempty"`
	ID       string    `dynamodbav:"id"`
	TIR      uint64    `dynamodbav:"tir"`
	Status   string    `dynamodbav:"status"`
	StartsAt time.Time `dynamodbav:"startsAt"`
	ClosesAt time.Time `dynamodbav:"closesAt"`
	Balls    []int     `dynamodbav:"balls,omitempty"`
	VideoID  string    `dynamodbav:"videoId,omitempty"`
}

func toItem(d *draw.Draw) drawItem {
	return drawItem{
		ID: d.ID, TIR: d.TIR, Status: d.Status,
		StartsAt: d.StartsAt.UTC(), ClosesAt: d.ClosesAt.UTC(),
		Balls: d.Balls, VideoID: d.VideoID,
	}
}

func (item drawItem) draw() *draw.Draw {
	return &draw.Draw{
		ID: item.ID, TIR: item.TIR, Status: item.Status,
		StartsAt: item.StartsAt.UTC(), ClosesAt: item.ClosesAt.UTC(),
		Balls: item.Balls, VideoID: item.VideoID,
	}
}

func key(pk string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: pk},
		"SK": &types.AttributeValueMemberS{Value: "META"},
	}
}
