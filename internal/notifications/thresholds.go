package notifications

// exhaustionAlert devuelve el umbral de agotamiento más alto alcanzado
// (p. ej. 90, 98, 100) si el consumo lo cruza.
func exhaustionAlert(consumedPct float64, thresholds []float64) (float64, bool) {
	var highest float64
	var found bool
	for _, th := range thresholds {
		if consumedPct >= th && th >= highest {
			highest = th
			found = true
		}
	}
	return highest, found
}
