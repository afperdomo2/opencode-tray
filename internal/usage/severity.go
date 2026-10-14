package usage

// Severity calcula el nivel de severidad a partir del mayor consumo entre las
// tres ventanas. Regla (del spec de usage):
//   - >= 98%  → LevelCritical (rojo, "al límite / se acabó")
//   - >= 90%  → LevelWarn (amarillo/naranja, "casi al límite")
//   - resto   → LevelOK (verde)
func (s *Service) Severity(u Usage) Level {
	maxPct := u.Rolling.Percent
	if u.Weekly.Percent > maxPct {
		maxPct = u.Weekly.Percent
	}
	if u.Monthly.Percent > maxPct {
		maxPct = u.Monthly.Percent
	}

	switch {
	case maxPct >= 98:
		return LevelCritical
	case maxPct >= 90:
		return LevelWarn
	default:
		return LevelOK
	}
}
