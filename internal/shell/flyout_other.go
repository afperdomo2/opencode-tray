//go:build !windows

package shell

// En plataformas no Windows el flyout usa la ventana normal de Wails.

func CursorPosition() Point { return Point{} }

func PrepareFlyoutWindow() {}

func PositionFlyout(anchor Point, fallbackWidth, fallbackHeight int) {}

func FocusFlyout() {}

// FlyoutForeground no aplica fuera de Windows.
func FlyoutForeground() bool { return true }
