package main

import (
	"testing"
	"time"

	"opencode-tray/internal/usage"
)

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
