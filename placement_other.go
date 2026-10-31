//go:build !windows

package main

import (
	"context"
	"sync/atomic"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"opencode-tray/internal/shell"
)

// centered evita volver a centrar el panel: en Linux la ventana se puede
// arrastrar y esa posición debe conservarse al ocultar y volver a mostrar.
var centered atomic.Bool

// positionFlyout centra el panel en el monitor principal la primera vez. En
// Linux no hay API para leer la geometría del icono de la bandeja, y Wayland
// puede ignorar la posición (el compositor ya la centra).
func positionFlyout(ctx context.Context, anchor shell.Point, width, height int) {
	if !centered.CompareAndSwap(false, true) {
		return
	}
	screens, err := wailsruntime.ScreenGetAll(ctx)
	if err != nil || len(screens) == 0 {
		return
	}
	primary := screens[0]
	for _, s := range screens {
		if s.IsPrimary {
			primary = s
			break
		}
	}
	work := shell.Rect{
		Left:   0,
		Top:    0,
		Right:  primary.Size.Width,
		Bottom: primary.Size.Height,
	}
	p := shell.CenterPosition(work, width, height)
	wailsruntime.WindowSetPosition(ctx, p.X, p.Y)
}
