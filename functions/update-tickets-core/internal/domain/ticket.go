package domain

import "time"

// TicketState representa el estado de un ticket en el core.
type TicketState string

var (
	PENDING  TicketState = "PENDING"
	DRAWING  TicketState = "DRAWING"
	LOSS     TicketState = "LOSS"
	PAYED    TicketState = "PAYED"
	WINNING  TicketState = "WINNING"
	CANCELED TicketState = "CANCELED"
)

// GameKeno4Min es el discriminador de juego en el core. Los tickets del core son
// polimórficos (una colección para todos los juegos), así que esta Lambda SOLO
// resuelve los de KENO4MIN.
const GameKeno4Min = "KENO4MIN"

// Ticket representa una jugada tal como la persiste el microservicio core. A
// diferencia del ticket del engine, el core NO guarda el bitmask de cada apuesta
// (solo los números en metadata.bets[].nums) ni las balotas sorteadas, así que
// esta Lambda reconstruye el bitmask al vuelo para calcular el payout.
type Ticket struct {
	ID        string      `bson:"_id,omitempty" json:"id"`
	Game      string      `bson:"game" json:"game"`
	GameID    string      `bson:"game_id" json:"gameId"`
	ClubID    string      `bson:"club_id" json:"clubId"`
	AdminID   string      `bson:"admin_id" json:"adminId"`
	Cupon     string      `bson:"cupon" json:"cupon"`
	State     TicketState `bson:"state" json:"state"`
	Win       int64       `bson:"win" json:"win"`
	Jackpot   int64       `bson:"jackpot" json:"jackpot"`
	Total     int64       `bson:"total" json:"total"`
	CreatedBy string      `bson:"created_by" json:"createdBy"`
	CreatedAt time.Time   `bson:"created_at" json:"createdAt"`
	Metadata  Metadata    `bson:"metadata" json:"metadata"`
	Version   int64       `bson:"version" json:"version"`
}

// Metadata agrupa los datos variables del ticket. En el core las apuestas viven
// bajo metadata.bets, no en la raíz del documento. Balls guarda las balotas
// sorteadas de la partida para que cada ticket sea auto-contenido (auditoría sin
// ir a buscar el game); se llena al resolver.
type Metadata struct {
	Bets  []Bet   `bson:"bets" json:"bets"`
	Balls []int64 `bson:"balls,omitempty" json:"balls,omitempty"`
}

// Bet representa una apuesta individual del core: solo números y monto. El
// bitmask se deriva de Nums en tiempo de resolución.
type Bet struct {
	Nums  []int64 `bson:"nums" json:"nums"`
	Money int64   `bson:"money" json:"money"`
}

// Payout calcula lo que paga esta apuesta contra las balotas sorteadas: el
// factor de PaymentTable[picks][aciertos] aplicado al monto apostado. picks y
// aciertos salen del bitmask reconstruido desde Nums.
func (b Bet) Payout(balls Bitmask) int64 {
	betMask := NewBitmask(b.Nums)
	picks := betMask.Count()
	hits := betMask.Matches(balls)

	factors, ok := PaymentTable[picks]
	if !ok || hits >= len(factors) {
		return 0
	}
	return b.Money * factors[hits] / 100
}

// Payout es lo que gana el jugador con este ticket: la suma de sus apuestas.
func (t Ticket) Payout(balls Bitmask) int64 {
	var total int64
	for _, bet := range t.Metadata.Bets {
		total += bet.Payout(balls)
	}
	return total
}

// Resolve liquida el ticket del core contra las balotas sorteadas: calcula el
// win, deriva el nuevo estado (WINNING si win > 0, LOSS si no) y copia las
// balotas sorteadas a metadata.balls para dejar el ticket auto-contenido. Solo
// debe invocarse sobre tickets PENDING (ver ports.TicketRepository).
func (t *Ticket) Resolve(balls GameBalls) {
	t.Win = t.Payout(balls.Mask)
	if t.Win > 0 {
		t.State = WINNING
	} else {
		t.State = LOSS
	}
	t.Metadata.Balls = balls.Nums
}
