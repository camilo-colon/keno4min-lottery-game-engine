package domain

// AdminStats representa las estadísticas acumuladas de un admin
type AdminStats struct {
	AdminID     string  `bson:"admin_id"`
	TotalIncome float64 `bson:"total_income"`
	TotalPaid   float64 `bson:"total_paid"`
}

// IsValid verifica si las estadísticas son válidas para cálculos
func (s *AdminStats) IsValid() bool {
	return s.TotalIncome > 0
}

// CurrentRTP calcula el RTP actual del admin
func (s *AdminStats) CurrentRTP() float64 {
	if s.TotalIncome <= 0 {
		return 0
	}
	return s.TotalPaid / s.TotalIncome
}

// ProjectedRTP calcula el RTP proyectado considerando un payout adicional
func (s *AdminStats) ProjectedRTP(additionalPayout float64) float64 {
	if s.TotalIncome <= 0 {
		return 0
	}
	return (s.TotalPaid + additionalPayout) / s.TotalIncome
}
