package dynamodb

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	draw "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/draw/domain"
)

type drawItem struct {
	DrawID  string `dynamodbav:"drawId"`
	TIR     uint64 `dynamodbav:"tir"`
	Balls   []int  `dynamodbav:"balls"`
	VideoID string `dynamodbav:"videoId"`
}

func toItem(d *draw.Draw) drawItem {
	balls := d.Balls
	if balls == nil {
		balls = []int{}
	}
	return drawItem{
		DrawID: d.ID, TIR: d.TIR, Balls: balls, VideoID: d.VideoID,
	}
}

func key(pk string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: pk},
		"SK": &types.AttributeValueMemberS{Value: "META"},
	}
}
