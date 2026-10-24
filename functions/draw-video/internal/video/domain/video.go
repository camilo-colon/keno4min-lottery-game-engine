package domain

type Video struct {
	ID    string `json:"id"`
	Balls []int  `json:"balls"`
}

func New(id string, balls []int) *Video {
	return &Video{ID: id, Balls: balls}
}
