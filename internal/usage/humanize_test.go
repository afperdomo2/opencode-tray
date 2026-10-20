package usage

import (
	"strings"
	"testing"
	"time"
	"unicode/utf16"
)

func fixedNow() time.Time {
	loc := time.FixedZone("COT", -5*60*60)
	return time.Date(2026, time.June, 12, 10, 0, 0, 0, loc)
}

func TestHumanResetSpanish(t *testing.T) {
	now := fixedNow()
	cases := []struct {
		name string
		win  Window
		want string
	}{
		{"sin reset", Window{}, "sin reinicio"},
		{"vencido", Window{ResetsAt: now.Add(-time.Minute), HasReset: true}, "ahora"},
		{"menos de un minuto", Window{ResetsAt: now.Add(30 * time.Second), HasReset: true}, "en <1 min"},
		{"minutos", Window{ResetsAt: now.Add(47 * time.Minute), HasReset: true}, "en 47 min"},
		{"horas", Window{ResetsAt: now.Add(4*time.Hour + 20*time.Minute), HasReset: true}, "en 4 h"},
		{"mañana", Window{ResetsAt: now.AddDate(0, 0, 1), HasReset: true}, "mañana a las 10:00"},
		{"fecha", Window{ResetsAt: time.Date(2026, time.June, 15, 12, 30, 0, 0, now.Location()), HasReset: true}, "el 15 jun 26"},
	}
	for _, tc := range cases {
		if got := HumanReset(tc.win, now, "es"); got != tc.want {
			t.Errorf("%s: HumanReset = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestHumanResetEnglish(t *testing.T) {
	now := fixedNow()
	cases := []struct {
		name string
		win  Window
		want string
	}{
		{"no reset", Window{}, "no reset"},
		{"now", Window{ResetsAt: now.Add(-time.Minute), HasReset: true}, "now"},
		{"minutes", Window{ResetsAt: now.Add(5 * time.Minute), HasReset: true}, "in 5 min"},
		{"hours", Window{ResetsAt: now.Add(4 * time.Hour), HasReset: true}, "in 4 h"},
		{"tomorrow", Window{ResetsAt: time.Date(2026, time.June, 13, 15, 30, 0, 0, now.Location()), HasReset: true}, "tomorrow at 3:30 PM"},
		{"date", Window{ResetsAt: time.Date(2026, time.June, 15, 12, 30, 0, 0, now.Location()), HasReset: true}, "Jun 15, 2026"},
	}
	for _, tc := range cases {
		if got := HumanReset(tc.win, now, "en"); got != tc.want {
			t.Errorf("%s: HumanReset = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestBatteryBar(t *testing.T) {
	cases := []struct {
		pct  float64
		want string
	}{
		{0, "▱▱▱▱▱"},
		{19, "▱▱▱▱▱"},
		{20, "▰▱▱▱▱"},
		{42, "▰▰▱▱▱"},
		{99, "▰▰▰▰▱"},
		{100, "▰▰▰▰▰"},
		{150, "▰▰▰▰▰"},
	}
	for _, tc := range cases {
		if got := BatteryBar(tc.pct); got != tc.want {
			t.Errorf("BatteryBar(%v) = %q, want %q", tc.pct, got, tc.want)
		}
	}
}

func TestTooltip(t *testing.T) {
	now := fixedNow()
	u := Usage{
		Monthly: Window{Percent: 42, ResetsAt: now.Add(4 * time.Hour), HasReset: true},
	}
	got := Tooltip(u, TooltipOptions{Language: "es", Now: now, Icons: EmojiIcons})
	want := "OpenCode · Mensual ▰▰▱▱▱ 42% · 🕒 en 4 h"
	if got != want {
		t.Errorf("Tooltip = %q, want %q", got, want)
	}

	got = Tooltip(u, TooltipOptions{Language: "en", Now: now, Icons: EmojiIcons})
	want = "OpenCode · Monthly ▰▰▱▱▱ 42% · 🕒 in 4 h"
	if got != want {
		t.Errorf("Tooltip EN = %q, want %q", got, want)
	}
}

func TestTooltipUsesCalendarForFutureDate(t *testing.T) {
	now := fixedNow()
	u := Usage{Monthly: Window{Percent: 8, ResetsAt: time.Date(2026, time.June, 15, 12, 30, 0, 0, now.Location()), HasReset: true}}
	got := Tooltip(u, TooltipOptions{Language: "es", Now: now, Icons: SafeIcons})
	if !strings.HasPrefix(got, "OpenCode · Mensual ▱▱▱▱▱ 8% · ▦ el 15 jun 26") {
		t.Errorf("Tooltip = %q", got)
	}
}

func TestTooltipNoReset(t *testing.T) {
	now := fixedNow()
	u := Usage{Monthly: Window{Percent: 0}}
	got := Tooltip(u, TooltipOptions{Language: "es", Now: now, Icons: EmojiIcons})
	if !strings.Contains(got, "↻ sin reinicio") {
		t.Errorf("Tooltip = %q", got)
	}
}

func TestTruncateUTF16(t *testing.T) {
	// Cada emoji ocupa 2 unidades UTF-16.
	s := strings.Repeat("🕒", 80)
	got := TruncateUTF16(s, 127)
	if units := len([]rune(got)); units != 63 {
		t.Errorf("runas = %d, want 63", units)
	}
	if !strings.HasPrefix(s, got) {
		t.Error("el truncado debe ser un prefijo")
	}
}

func TestTooltipNeverExceedsLimit(t *testing.T) {
	now := fixedNow()
	u := Usage{Monthly: Window{Percent: 100, ResetsAt: time.Date(2126, time.December, 31, 23, 59, 0, 0, now.Location()), HasReset: true}}
	got := Tooltip(u, TooltipOptions{Language: "es", Now: now, Icons: EmojiIcons})
	if n := len(utf16.Encode([]rune(got))); n > 127 {
		t.Errorf("tooltip de %d unidades UTF-16", n)
	}
}
