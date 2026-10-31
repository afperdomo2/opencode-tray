//go:build windows

package main

import (
	"context"

	"opencode-tray/internal/shell"
)

// positionFlyout ancla el panel al icono del tray usando la clase Win32 propia
// (ver internal/shell/flyout_windows.go).
func positionFlyout(ctx context.Context, anchor shell.Point, width, height int) {
	shell.PositionFlyout(anchor, width, height)
}
