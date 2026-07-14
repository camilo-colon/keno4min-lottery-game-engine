package domain

// Club representa un club con su configuración, balance y jackpots.
type Club struct {
	ID         string     `bson:"_id,omitempty" json:"id"`
	Name       string     `bson:"name" json:"name"`
	Address    string     `bson:"address" json:"address"`
	Balance    int64      `bson:"balance" json:"balance"`
	Room       int64      `bson:"room" json:"room"`
	Active     bool       `bson:"active" json:"active"`
	CashiersID []string   `bson:"cashiers_id" json:"cashiersId"`
	Config     ClubConfig `bson:"config" json:"config"`
	AdminID    string     `bson:"admin_id" json:"adminId"`
	Version    int64      `bson:"version" json:"version"`
	// JP1 es la instancia viva del jackpot #1 embebida en el club.
	JP1 Jackpot `bson:"jp1" json:"jp1"`
}

// ClubConfig agrupa la configuración del club.
type ClubConfig struct {
	Bet     BetConfig     `bson:"bet" json:"bet"`
	Region  RegionConfig  `bson:"region" json:"region"`
	Jackpot JackpotConfig `bson:"jackpot" json:"jackpot"`
	Values  []int64       `bson:"values" json:"values"`
}

// BetConfig define los límites de apuesta del club.
type BetConfig struct {
	Max int64 `bson:"max" json:"max"`
	Min int64 `bson:"min" json:"min"`
}

// RegionConfig define la localización del club.
type RegionConfig struct {
	Timezone string `bson:"timezone" json:"timezone"`
	Currency string `bson:"currency" json:"currency"`
	Locale   string `bson:"locale" json:"locale"`
}

// JackpotConfig define las reglas por defecto del jackpot del club.
// Nota: son las reglas/defaults, no la instancia viva (esa es Club.JP1).
type JackpotConfig struct {
	Min     int64 `bson:"min" json:"min"`
	Max     int64 `bson:"max" json:"max"`
	Percent int64 `bson:"percent" json:"percent"`
	Seed    int64 `bson:"seed" json:"seed"`
}
