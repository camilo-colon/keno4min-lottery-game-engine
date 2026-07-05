package config

// Zone representa el nivel de intervención del sistema sobre un admin
type Zone int

const (
	ZoneDead Zone = iota // Sin intervención, draw puro aleatorio
	ZoneSoft             // Corrección suave
	ZoneHard             // Corrección agresiva
)

// RTPConfig contiene la configuración de RTP (Return To Player) por bandas
type RTPConfig struct {
	TargetRTP          float64
	NumDrawCandidates  int
	TopNCandidates     int
	MaxCorrectionRatio float64

	// Bandas: dead zone = sin intervención
	DeadZoneLow  float64 // Límite inferior de dead zone (ej: 0.88)
	DeadZoneHigh float64 // Límite superior de dead zone (ej: 0.92)

	// Bandas: hard zone = corrección agresiva
	HardZoneLow  float64 // Por debajo de esto → hard correction (ej: 0.85)
	HardZoneHigh float64 // Por encima de esto → hard correction (ej: 0.95)

	// Penalización por zona (aplicada al overpay)
	SoftPenalty float64 // Penalización en soft zone (ej: 1.2)
	HardPenalty float64 // Penalización en hard zone (ej: 2.0)
}

// Zone clasifica un RTP en su zona de intervención
func (c *RTPConfig) Zone(rtp float64) Zone {
	if rtp >= c.DeadZoneLow && rtp <= c.DeadZoneHigh {
		return ZoneDead
	}
	if rtp < c.HardZoneLow || rtp >= c.HardZoneHigh {
		return ZoneHard
	}
	return ZoneSoft
}

// Penalty retorna el multiplicador de penalización para una zona.
// Dead zone retorna 0 — estos admins no influyen en la selección.
func (c *RTPConfig) Penalty(z Zone) float64 {
	switch z {
	case ZoneSoft:
		return c.SoftPenalty
	case ZoneHard:
		return c.HardPenalty
	default:
		return 0
	}
}

// loadRTPConfig carga la configuración de RTP desde variables de entorno con defaults
func loadRTPConfig() RTPConfig {
	return RTPConfig{
		TargetRTP:          getEnvFloat("RTP_TARGET", 0.90),
		NumDrawCandidates:  getEnvInt("RTP_NUM_DRAW_CANDIDATES", 1000),
		TopNCandidates:     getEnvInt("RTP_TOP_N_CANDIDATES", 50),
		MaxCorrectionRatio: getEnvFloat("RTP_MAX_CORRECTION_RATIO", 2.0),

		DeadZoneLow:  getEnvFloat("RTP_DEAD_ZONE_LOW", 0.88),
		DeadZoneHigh: getEnvFloat("RTP_DEAD_ZONE_HIGH", 0.92),
		HardZoneLow:  getEnvFloat("RTP_HARD_ZONE_LOW", 0.85),
		HardZoneHigh: getEnvFloat("RTP_HARD_ZONE_HIGH", 0.95),

		SoftPenalty: getEnvFloat("RTP_SOFT_PENALTY", 1.2),
		HardPenalty: getEnvFloat("RTP_HARD_PENALTY", 2.0),
	}
}
