package dynamodb

import video "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/video/domain"

type videoItem struct {
	PK    string `dynamodbav:"PK"`
	SK    string `dynamodbav:"SK"`
	ID    string `dynamodbav:"id"`
	Balls []int  `dynamodbav:"balls"`
}

func (item videoItem) video() video.Video {
	return video.Video{ID: item.ID, Balls: item.Balls}
}
