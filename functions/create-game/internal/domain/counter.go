package domain

type Counter struct {
	Id    string `bson:"_id" json:"id"`
	Count uint64 `bson:"value" json:"value"`
}

func NewCounter(id string, count uint64) *Counter {
	return &Counter{
		Id:    id,
		Count: count,
	}
}
