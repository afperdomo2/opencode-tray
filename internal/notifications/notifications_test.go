package notifications

import (
	"testing"
	"time"

	"opencode-tray/internal/usage"
)

type captured struct {
	title string
	msg   string
	lvl   Level
}

func newTestService(now time.Time, out *[]captured) *Service {
	return &Service{
		fired: map[string]bool{},
		notify: func(title, msg string, lvl Level) error {
			*out = append(*out, captured{title, msg, lvl})
			return nil
		},
		now:        func() time.Time { return now },
		milestones: []float64{20, 40, 60, 80},
		exhaustion: []float64{90, 98, 100},
		margin:     15,
		enabled:    map[string]bool{"burnrate": true, "exhaustion": true, "service": true},
	}
}

func TestBurnRateAlert(t *testing.T) {
	ms := []float64{20, 40, 60, 80}
	if m, ok := burnRateAlert(20, 5, 15, ms); !ok || m != 20 {
		t.Errorf("20 vs5 → got %v %v, want20 true", m, ok)
	}
	if _, ok := burnRateAlert(20, 18, 15, ms); ok {
		t.Error("al ritmo no debería alertar")
	}
	if _, ok := burnRateAlert(10, 5, 15, ms); ok {
		t.Error("10% no cruza milestone")
	}
}

func TestExpectedPercent(t *testing.T) {
	resetsAt := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	total := 7 * 24 * time.Hour
	if got := expectedPercent(resetsAt.Add(-total), resetsAt, total); got != 0 {
		t.Errorf("inicio = %v", got)
	}
	if got := expectedPercent(resetsAt.Add(-total/2), resetsAt, total); got != 50 {
		t.Errorf("mitad = %v", got)
	}
}

func TestExhaustionAlert(t *testing.T) {
	th := []float64{90, 98, 100}
	if _, ok := exhaustionAlert(85, th); ok {
		t.Error("85% no debería alertar")
	}
	if got, ok := exhaustionAlert(92, th); !ok || got != 90 {
		t.Errorf("92 → got %v %v, want90", got, ok)
	}
	if got, ok := exhaustionAlert(99, th); !ok || got != 98 {
		t.Errorf("99 → got %v %v, want98", got, ok)
	}
}

func TestServiceAlertTransition(t *testing.T) {
	if serviceAlert(true, true) {
		t.Error("corriendo→corriendo no alerta")
	}
	if !serviceAlert(true, false) {
		t.Error("corriendo→detenido alerta")
	}
	if serviceAlert(false, false) {
		t.Error("detenido→detenido no alerta")
	}
}

func TestEvaluateBurnRateOnceAndReset(t *testing.T) {
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	// Ventana semanal que termina en6 días → empezó hace1 día (esperado ~14%).
	// Consumido60% → adelantado → alerta en milestone60.
	resetsAt := now.Add(6 * 24 * time.Hour)
	u := usage.Usage{Weekly: usage.Window{Percent: 60, ResetsAt: resetsAt, HasReset: true}}

	var out []captured
	s := newTestService(now, &out)

	s.Evaluate(u, true)
	if len(out) != 1 {
		t.Fatalf("alerts = %d, want 1: %+v", len(out), out)
	}
	if out[0].title != "Ritmo de consumo" {
		t.Errorf("title = %q", out[0].title)
	}

	// Misma llamada → dedupe, no duplica.
	s.Evaluate(u, true)
	if len(out) != 1 {
		t.Errorf("dedupe falló, alerts = %d", len(out))
	}

	// Nuevo período (resetsAt distinto) → vuelve a disparar.
	u.Weekly.ResetsAt = resetsAt.Add(7 * 24 * time.Hour)
	s.Evaluate(u, true)
	if len(out) != 2 {
		t.Errorf("nuevo período debería disparar, alerts = %d", len(out))
	}
}

func TestEvaluateExhaustion(t *testing.T) {
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	u := usage.Usage{Rolling: usage.Window{Percent: 95}}

	var out []captured
	s := newTestService(now, &out)
	s.Evaluate(u, false) // svc no corriendo y prev false → sin alerta de servicio

	if len(out) != 1 {
		t.Fatalf("alerts = %d, want 1: %+v", len(out), out)
	}
	if out[0].title != "Consumo alto" || out[0].lvl != LevelWarn {
		t.Errorf("alert = %+v, want Consumo alto/Warn", out[0])
	}
}

func TestEvaluateServiceStop(t *testing.T) {
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	var out []captured
	s := newTestService(now, &out)

	s.Evaluate(usage.Usage{}, true) // primer estado: corriendo, sin alerta
	if len(out) != 0 {
		t.Errorf("no debería alertar aún, alerts = %d", len(out))
	}
	s.Evaluate(usage.Usage{}, false) // corriendo → detenido
	if len(out) != 1 || out[0].title != "Servicio detenido" {
		t.Errorf("alerts = %+v, want Servicio detenido", out)
	}
}

func TestEvaluateRespectsEnabled(t *testing.T) {
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	u := usage.Usage{Rolling: usage.Window{Percent: 99}}
	var out []captured
	s := newTestService(now, &out)
	s.SetEnabled("exhaustion", false)
	s.Evaluate(u, false)
	if len(out) != 0 {
		t.Errorf("exhaustion deshabilitado no debería alertar, alerts = %d", len(out))
	}
}
