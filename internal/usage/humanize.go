package usage

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf16"
)

// Icons agrupa los glifos que acompañan al consumo en el tooltip nativo.
// Windows solo admite texto en el tooltip, así que los "iconos" son Unicode.
type Icons struct {
	Clock    string
	Calendar string
	Reset    string
	Warning  string
}

// EmojiIcons es la variante preferida (Windows 10/11 con Segoe UI Emoji).
var EmojiIcons = Icons{Clock: "🕒", Calendar: "🗓", Reset: "↻", Warning: "⚠"}

// SafeIcons es el respaldo si el shell no renderiza emoji (glifos geométricos).
var SafeIcons = Icons{Clock: "◷", Calendar: "▦", Reset: "↻", Warning: "!"}

// TooltipOptions controla el texto del tooltip.
type TooltipOptions struct {
	Language string
	Icons    Icons
	Now      time.Time
}

var monthES = [...]string{"ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"}
var monthEN = [...]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

// Tooltip construye el texto del tooltip del tray: barra de consumo mensual y
// reinicio humanizado. Se limita a 127 unidades UTF-16 para no desbordar szTip.
func Tooltip(u Usage, opts TooltipOptions) string {
	lang := normalizeLang(opts.Language)
	icons := opts.Icons
	if icons == (Icons{}) {
		icons = EmojiIcons
	}
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}

	pct := clampPercent(u.Monthly.Percent)
	label := "Mensual"
	if lang == "en" {
		label = "Monthly"
	}

	parts := []string{
		"OpenCode",
		label + " " + fmt.Sprintf("%s %d%%", BatteryBar(pct), int(math.Round(pct))),
		iconsForReset(u.Monthly, now, icons) + " " + HumanReset(u.Monthly, now, lang),
	}
	return TruncateUTF16(strings.Join(parts, " · "), 127)
}

// HumanReset describe cuándo se restablece una ventana en lenguaje natural.
// Reglas: mismo día → "en 4 h"; mañana → "mañana 12:30"; más lejos → fecha.
func HumanReset(w Window, now time.Time, lang string) string {
	lang = normalizeLang(lang)
	if !w.HasReset || w.ResetsAt.IsZero() {
		if lang == "en" {
			return "no reset"
		}
		return "sin reinicio"
	}

	reset := w.ResetsAt.In(now.Location())
	d := reset.Sub(now)
	if d <= 0 {
		if lang == "en" {
			return "now"
		}
		return "ahora"
	}

	sameDay := sameDate(reset, now)
	if sameDay {
		if d < time.Minute {
			if lang == "en" {
				return "in <1 min"
			}
			return "en <1 min"
		}
		if d < time.Hour {
			mins := int(math.Ceil(d.Minutes()))
			if lang == "en" {
				return fmt.Sprintf("in %d min", mins)
			}
			return fmt.Sprintf("en %d min", mins)
		}
		hours := int(d.Hours())
		if hours < 1 {
			hours = 1
		}
		if lang == "en" {
			return fmt.Sprintf("in %d h", hours)
		}
		return fmt.Sprintf("en %d h", hours)
	}

	tomorrow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
	dayAfter := tomorrow.AddDate(0, 0, 1)
	if reset.Before(dayAfter) {
		if lang == "en" {
			return "tomorrow at " + formatClock(reset, lang)
		}
		return "mañana a las " + formatClock(reset, lang)
	}

	return formatDate(reset, lang)
}

// BatteryBar dibuja el consumo en 5 celdas (0–100 %). La barra crece con lo
// consumido, no con lo restante, para no confundirla con una batería.
func BatteryBar(percent float64) string {
	pct := clampPercent(percent)
	filled := int(math.Floor(pct / 20))
	if filled > 5 {
		filled = 5
	}
	if filled < 0 {
		filled = 0
	}
	return strings.Repeat("▰", filled) + strings.Repeat("▱", 5-filled)
}

// TruncateUTF16 corta sin partir pares sustitutos (emoji) y sin superar max.
func TruncateUTF16(s string, max int) string {
	if max <= 0 {
		return ""
	}
	units := utf16.Encode([]rune(s))
	if len(units) <= max {
		return s
	}
	if units[max-1] >= 0xD800 && units[max-1] <= 0xDBFF {
		max--
	}
	return string(utf16.Decode(units[:max]))
}

func iconsForReset(w Window, now time.Time, icons Icons) string {
	if !w.HasReset || w.ResetsAt.IsZero() || !w.ResetsAt.After(now) {
		return icons.Reset
	}
	reset := w.ResetsAt.In(now.Location())
	if sameDate(reset, now) {
		return icons.Clock
	}
	return icons.Calendar
}

func sameDate(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func formatClock(t time.Time, lang string) string {
	if lang == "en" {
		return t.Format("3:04 PM")
	}
	return t.Format("15:04")
}

func formatDate(t time.Time, lang string) string {
	if lang == "en" {
		return fmt.Sprintf("%s %d, %d", monthEN[t.Month()-1], t.Day(), t.Year())
	}
	return fmt.Sprintf("el %d %s %02d", t.Day(), monthES[t.Month()-1], t.Year()%100)
}

func normalizeLang(lang string) string {
	if strings.EqualFold(strings.TrimSpace(lang), "en") {
		return "en"
	}
	return "es"
}

func clampPercent(percent float64) float64 {
	if math.IsNaN(percent) || percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}
