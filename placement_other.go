//go:build !windows

package main

import (
	"context"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"opencode-tray/internal/shell"
)

// positionFlyout coloca el panel en la esquina inferior derecha del monitor
// principal. En Linux/macOS no hay una API para leer la geometría del icono de
// la bandeja, así que se aproxima al comportamiento de un flyout del sistema.
// Bajo Wayland el compositor puede ignorar la posición: el panel aparecerá
// donde decida el gestor de ventanas.
func positionFlyout(ctx context.Context, anchor shell.Point, width, height int) {
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
	p := shell.CornerPosition(work, width, height)
	wailsruntime.WindowSetPosition(ctx, p.X, p.Y)
}
