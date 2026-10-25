package main

import (
	"testing"
	"time"

	"opencode-tray/internal/servicecontrol"
	"opencode-tray/internal/usage"
)

// statusRunner cuenta cuántas veces se invoca el CLI en los tests de caché.
type statusRunner struct {
	calls  int
	stdout string
}

func (r *statusRunner) Run(args []string) (string, string, error) {
	r.calls++
	return r.stdout, "", nil
}

func TestServiceStatusCacheTTL(t *testing.T) {
	r := &statusRunner{stdout: "http://127.0.0.1:5000\n"}
	now := time.Date(2026, time.June, 12, 10, 0, 0, 0, time.UTC)
	a := &App{
		service:   servicecontrol.NewWithRunner(r, nil),
		statusTTL: 2 * time.Second,
		now:       func() time.Time { return now },
	}

	st, err := a.ServiceStatus()
	if err != nil {
		t.Fatal(err)
	}
	if !st.Running || st.URL != "http://127.0.0.1:5000" {
		t.Errorf("status = %+v", st)
	}

	// Dentro del TTL no se vuelve a invocar el CLI.
	now = now.Add(time.Second)
	if _, err := a.ServiceStatus(); err != nil {
		t.Fatal(err)
	}
	if r.calls != 1 {
		t.Errorf("calls = %d, want 1 dentro del TTL", r.calls)
	}

	// Tras una mutación, la caché se invalida y se relee.
	a.invalidateStatusCache()
	if _, err := a.ServiceStatus(); err != nil {
		t.Fatal(err)
	}
	if r.calls != 2 {
		t.Errorf("calls = %d, want 2 tras invalidar", r.calls)
	}
}

func TestUsageTooltip(t *testing.T) {
	loc := time.FixedZone("COT", -5*60*60)
	now := time.Date(2026, time.June, 12, 10, 0, 0, 0, loc)
	u := usage.Usage{
		Rolling: usage.Window{Percent: 17.6},
		Weekly:  usage.Window{Percent: 42.4},
		Monthly: usage.Window{Percent: 42, ResetsAt: now.Add(4 * time.Hour), HasReset: true},
	}
	want := "OpenCode · Mensual ▰▰▱▱▱ 42% · 🕒 en 4 h"
	if got := usageTooltip(u, "es", now); got != want {
		t.Errorf("tooltip = %q, want %q", got, want)
	}
}

func TestUsageTooltipEnglish(t *testing.T) {
	loc := time.FixedZone("COT", -5*60*60)
	now := time.Date(2026, time.June, 12, 10, 0, 0, 0, loc)
	u := usage.Usage{
		Monthly: usage.Window{Percent: 8, ResetsAt: now.AddDate(0, 0, 1), HasReset: true},
	}
	want := "OpenCode · Monthly ▱▱▱▱▱ 8% · 🗓 tomorrow at 10:00 AM"
	if got := usageTooltip(u, "en", now); got != want {
		t.Errorf("tooltip = %q, want %q", got, want)
	}
}
