package main

import (
	"testing"

	"opencode-tray/internal/usage"
)

func TestUsageTooltip(t *testing.T) {
	u := usage.Usage{
		Rolling: usage.Window{Percent: 17.6},
		Weekly:  usage.Window{Percent: 42.4},
		Monthly: usage.Window{Percent: 99.5},
	}
	want := "OpenCode · 5h 18% · 7d 42% · 30d 100%"
	if got := usageTooltip(u); got != want {
		t.Errorf("tooltip = %q, want %q", got, want)
	}
}
