package draw

import "time"

const Duration = 157 * time.Second

type Draw struct {
	ID       string    `json:"id"`
	TIR      uint64    `json:"tir"`
	Status   string    `json:"status"`
	StartsAt time.Time `json:"startsAt"`
	ClosesAt time.Time `json:"closesAt"`
	Balls    []int     `json:"balls,omitempty"`
	VideoID  string    `json:"videoId,omitempty"`
}

func New(id string, tir uint64, startsAt time.Time) *Draw {
	startsAt = startsAt.UTC()
	return &Draw{
		ID: id, TIR: tir, Status: "open",
		StartsAt: startsAt, ClosesAt: startsAt.Add(Duration),
	}
}
