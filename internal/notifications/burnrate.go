package notifications

import "time"

// burnRateAlert devuelve el milestone (20/40/60/80) si el consumo está adelantado
// respecto al ritmo objetivo por más de `margin`. Solo aplica a períodos fijos.
func burnRateAlert(consumedPct, expectedPct, margin float64, milestones []float64) (float64, bool) {
	for _, m := range milestones {
		if consumedPct >= m && consumedPct-expectedPct >= margin {
			return m, true
		}
	}
	return 0, false
}

// expectedPercent calcula el % de tiempo transcurrido de una ventana fija.
func expectedPercent(now, resetsAt time.Time, total time.Duration) float64 {
	start := resetsAt.Add(-total)
	elapsed := now.Sub(start)
	if elapsed < 0 {
		return 0
	}
	if elapsed > total {
		return 100
	}
	return float64(elapsed) / float64(total) * 100
}
