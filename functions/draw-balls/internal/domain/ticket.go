package domain

type Bet struct {
	Money uint    `bson:"money" json:"money"`
	Mask  BitMask `bson:"bitmask" json:"bitmask"`
}

type Ticket struct {
	Id      string `bson:"_id,omitempty" json:"id"`
	GameId  string `bson:"game_id" json:"gameId"`
	ClubId  string `bson:"club_id" json:"clubId"`
	AdminId string `bson:"admin_id" json:"adminId"`
	Bets    []Bet  `bson:"bets" json:"bets"`
	Total   uint   `bson:"total" json:"total"`
}

type Stats struct {
	TotalIncome uint    `bson:"total_income" json:"totalIncome"`
	TotalPaid   uint    `bson:"total_paid" json:"totalPaid"`
	Rtp         float64 `bson:"rtp" json:"rtp"`
}
