package notifications

// serviceAlert detecta un "detenido inesperadamente": estaba corriendo y ya no.
// La propia transición evita el spam (solo dispara una vez por parada).
func serviceAlert(prevRunning, running bool) bool {
	return prevRunning && !running
}
