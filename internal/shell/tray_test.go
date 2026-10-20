package shell

import "testing"

func TestTrayHandlerSetTooltip(t *testing.T) {
	var got string
	h := &TrayHandler{setTooltip: func(text string) { got = text }}

	h.SetTooltip("OpenCode · 5h 10% · 7d 20% · 30d 30%")
	if got != "OpenCode · 5h 10% · 7d 20% · 30d 30%" {
		t.Errorf("tooltip = %q", got)
	}

	h.SetTooltip("")
	if got != "OpenCode · 5h 10% · 7d 20% · 30d 30%" {
		t.Error("un texto vacío no debe pisar el tooltip")
	}
}
