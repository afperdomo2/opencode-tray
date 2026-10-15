package notifications

import (
	"fmt"
	"sync"
	"time"

	"opencode-tray/internal/usage"
)

// Level es la severidad de una notificación.
type Level int

const (
	LevelInfo Level = iota
	LevelWarn
	LevelCritical
)

// Service gestiona y dispara notificaciones nativas. La lógica de decisión es
// pura (ver burnrate.go / thresholds.go / servicewatch.go) y el envío va por un
// sender inyectable, para poder testear sin notificaciones reales.
type Service struct {
	mu          sync.Mutex
	fired       map[string]bool
	prevRunning bool

	notify     func(title, msg string, lvl Level) error
	now        func() time.Time
	milestones []float64
	exhaustion []float64
	margin     float64
	enabled    map[string]bool
}

// New crea un Service con la lógica por defecto (sender no-op; app.go inyecta el real).
func New() *Service {
	return &Service{
		fired:      map[string]bool{},
		notify:     noopNotify,
		now:        time.Now,
		milestones: []float64{20, 40, 60, 80},
		exhaustion: []float64{90, 98, 100},
		margin:     15,
		enabled:    map[string]bool{"burnrate": true, "exhaustion": true, "service": true},
	}
}

// SetEnabled habilita/deshabilita un tipo de alerta: "burnrate", "exhaustion" o "service".
func (s *Service) SetEnabled(kind string, on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled[kind] = on
}

// SetSender inyecta el envío real de notificaciones (p. ej. Wails SendNotification).
func (s *Service) SetSender(fn func(title, msg string, lvl Level) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notify = fn
}

// Evaluate evalúa el consumo y el estado del servicio, disparando solo las
// alertas nuevas (dedupe por ventana/período).
func (s *Service) Evaluate(u usage.Usage, svcRunning bool) {
	now := s.now()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Ritmo: solo ventanas de período fijo (weekly/monthly).
	if s.enabled["burnrate"] {
		s.evalBurnRate("weekly", u.Weekly, now, 7*24*time.Hour)
		s.evalBurnRate("monthly", u.Monthly, now, time.Duration(daysInMonth(now))*24*time.Hour)
	}
	// Agotamiento: las tres ventanas.
	if s.enabled["exhaustion"] {
		s.evalExhaustion("rolling", u.Rolling)
		s.evalExhaustion("weekly", u.Weekly)
		s.evalExhaustion("monthly", u.Monthly)
	}
	// Servicio detenido inesperadamente (transición corriendo → detenido).
	if s.enabled["service"] {
		if serviceAlert(s.prevRunning, svcRunning) {
			_ = s.notify("Servicio detenido", "El servicio de OpenCode se detuvo inesperadamente.", LevelCritical)
		}
		s.prevRunning = svcRunning
	}
}

func (s *Service) evalBurnRate(name string, w usage.Window, now time.Time, total time.Duration) {
	if !w.HasReset {
		return
	}
	expected := expectedPercent(now, w.ResetsAt, total)
	m, ok := burnRateAlert(w.Percent, expected, s.margin, s.milestones)
	if !ok {
		return
	}
	key := fmt.Sprintf("burnrate|%s|%v|%s", name, m, w.ResetsAt)
	if !s.shouldFire(key) {
		return
	}
	_ = s.notify("Ritmo de consumo",
		fmt.Sprintf("%s: %.0f%% usado (ritmo objetivo ~%.0f%%). Vas más rápido de lo previsto.", name, w.Percent, expected),
		LevelInfo)
}

func (s *Service) evalExhaustion(name string, w usage.Window) {
	th, ok := exhaustionAlert(w.Percent, s.exhaustion)
	if !ok {
		return
	}
	key := fmt.Sprintf("exhaustion|%s|%v|%s", name, th, resetsKey(w))
	if !s.shouldFire(key) {
		return
	}
	lvl := LevelWarn
	if th >= 98 {
		lvl = LevelCritical
	}
	_ = s.notify("Consumo alto",
		fmt.Sprintf("%s: %.0f%% usado (umbral %v%%).", name, w.Percent, th),
		lvl)
}

func (s *Service) shouldFire(key string) bool {
	if s.fired[key] {
		return false
	}
	s.fired[key] = true
	return true
}

func resetsKey(w usage.Window) string {
	if w.HasReset {
		return w.ResetsAt.String()
	}
	return ""
}

func daysInMonth(t time.Time) int {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
