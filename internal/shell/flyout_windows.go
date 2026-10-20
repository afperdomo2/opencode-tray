//go:build windows

package shell

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	user32                     = syscall.NewLazyDLL("user32.dll")
	kernel32                   = syscall.NewLazyDLL("kernel32.dll")
	procFindWindowW            = user32.NewProc("FindWindowW")
	procGetCursorPos           = user32.NewProc("GetCursorPos")
	procMonitorFromPoint       = user32.NewProc("MonitorFromPoint")
	procGetMonitorInfoW        = user32.NewProc("GetMonitorInfoW")
	procSetWindowPos           = user32.NewProc("SetWindowPos")
	procGetWindowRect          = user32.NewProc("GetWindowRect")
	procGetWindowLongW         = user32.NewProc("GetWindowLongW")
	procSetWindowLongW         = user32.NewProc("SetWindowLongW")
	procSetForeground          = user32.NewProc("SetForegroundWindow")
	procGetForeground          = user32.NewProc("GetForegroundWindow")
	procGetAncestor            = user32.NewProc("GetAncestor")
	procBringWindowToTop       = user32.NewProc("BringWindowToTop")
	procShowWindow             = user32.NewProc("ShowWindow")
	procSetFocus               = user32.NewProc("SetFocus")
	procSetActiveWindow        = user32.NewProc("SetActiveWindow")
	procAttachThreadInput      = user32.NewProc("AttachThreadInput")
	procGetWindowThreadProcess = user32.NewProc("GetWindowThreadProcessId")
	procGetCurrentThreadID     = kernel32.NewProc("GetCurrentThreadId")
)

const (
	monitorDefaultToNearest = 2
	gaRoot                  = 2

	swpNoSize       = 0x0001
	swpNoMove       = 0x0002
	swpNoActivate   = 0x0010
	swpFrameChanged = 0x0020

	gwlExStyle     = 0xFFFFFFEC // -20 como DWORD
	wsExToolWindow = 0x00000080
	wsExAppWindow  = 0x00040000

	swShow = 5
)

type winPoint struct {
	X int32
	Y int32
}

type winRect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type monitorInfo struct {
	CbSize    uint32
	RcMonitor winRect
	RcWork    winRect
	DwFlags   uint32
}

func rectFromWin(r winRect) Rect {
	return Rect{Left: int(r.Left), Top: int(r.Top), Right: int(r.Right), Bottom: int(r.Bottom)}
}

// CursorPosition devuelve la posición física del cursor (el punto de anclaje).
func CursorPosition() Point {
	var p winPoint
	if ok, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(&p))); ok == 0 {
		return Point{}
	}
	return Point{X: int(p.X), Y: int(p.Y)}
}

func findFlyoutWindow() uintptr {
	class, err := syscall.UTF16PtrFromString(FlyoutWindowClass)
	if err != nil {
		return 0
	}
	// La ventana puede tardar unos milisegundos en crearse al arrancar.
	for i := 0; i < 10; i++ {
		hwnd, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(class)), 0)
		if hwnd != 0 {
			return hwnd
		}
		time.Sleep(20 * time.Millisecond)
	}
	return 0
}

// ensureToolWindow aplica WS_EX_TOOLWINDOW una sola vez, para que la ventana
// quede fuera de la barra de tareas y de Alt+Tab.
func ensureToolWindow(hwnd uintptr) {
	style, _, _ := procGetWindowLongW.Call(hwnd, uintptr(gwlExStyle))
	if style&wsExToolWindow != 0 && style&wsExAppWindow == 0 {
		return
	}
	style = (style | wsExToolWindow) &^ wsExAppWindow
	procSetWindowLongW.Call(hwnd, uintptr(gwlExStyle), style)
	procSetWindowPos.Call(hwnd, 0, 0, 0, 0, 0, uintptr(swpNoMove|swpNoSize|swpFrameChanged))
}

// PrepareFlyoutWindow saca la ventana de la barra de tareas y de Alt+Tab.
func PrepareFlyoutWindow() {
	if hwnd := findFlyoutWindow(); hwnd != 0 {
		ensureToolWindow(hwnd)
	}
}

// PositionFlyout mueve la ventana para anclarla al punto dado. Usa el tamaño
// físico real de la ventana (ya escalado por DPI) y coordenadas absolutas.
func PositionFlyout(anchor Point, fallbackWidth, fallbackHeight int) {
	hwnd := findFlyoutWindow()
	if hwnd == 0 {
		return
	}
	// Reintenta el estilo en cada apertura por si el arranque llegó antes que
	// la creación de la ventana.
	ensureToolWindow(hwnd)

	width, height := fallbackWidth, fallbackHeight
	var windowRect winRect
	if ok, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&windowRect))); ok != 0 {
		if w := int(windowRect.Right - windowRect.Left); w > 0 {
			width = w
		}
		if h := int(windowRect.Bottom - windowRect.Top); h > 0 {
			height = h
		}
	}

	monitor := monitorFromPoint(anchor)
	if monitor == 0 {
		return
	}
	var info monitorInfo
	info.CbSize = uint32(unsafe.Sizeof(info))
	if ok, _, _ := procGetMonitorInfoW.Call(monitor, uintptr(unsafe.Pointer(&info))); ok == 0 {
		return
	}

	pos := AnchorPosition(anchor, rectFromWin(info.RcWork), rectFromWin(info.RcMonitor), width, height)
	procSetWindowPos.Call(
		hwnd,
		^uintptr(0), // HWND_TOPMOST
		uintptr(pos.X),
		uintptr(pos.Y),
		0,
		0,
		uintptr(swpNoSize|swpNoActivate),
	)
}

// monitorFromPoint pasa POINT por valor. En Windows amd64/arm64 (los targets de
// Wails) un struct de 8 bytes viaja en un único registro, por eso se empaqueta.
func monitorFromPoint(p Point) uintptr {
	packed := uint64(uint32(int32(p.X))) | uint64(uint32(int32(p.Y)))<<32
	monitor, _, _ := procMonitorFromPoint.Call(uintptr(packed), monitorDefaultToNearest)
	return monitor
}

// FocusFlyout trae la ventana al frente para que reciba teclado y blur. Si
// Windows bloquea la activación por provenir de otro hilo, se engancha a los
// hilos implicados durante la operación (receta clásica de AttachThreadInput).
func FocusFlyout() {
	hwnd := findFlyoutWindow()
	if hwnd == 0 {
		return
	}

	foreground, _, _ := procGetForeground.Call()
	if foreground == hwnd {
		return
	}

	current, _, _ := procGetCurrentThreadID.Call()
	foregroundThread, _, _ := procGetWindowThreadProcess.Call(foreground, 0)
	targetThread, _, _ := procGetWindowThreadProcess.Call(hwnd, 0)

	attached := make([]uintptr, 0, 2)
	if foregroundThread != 0 && foregroundThread != current {
		if ok, _, _ := procAttachThreadInput.Call(current, foregroundThread, 1); ok != 0 {
			attached = append(attached, foregroundThread)
		}
	}
	if targetThread != 0 && targetThread != current && targetThread != foregroundThread {
		if ok, _, _ := procAttachThreadInput.Call(current, targetThread, 1); ok != 0 {
			attached = append(attached, targetThread)
		}
	}

	procBringWindowToTop.Call(hwnd)
	procShowWindow.Call(hwnd, swShow)
	procSetForeground.Call(hwnd)
	procSetActiveWindow.Call(hwnd)
	procSetFocus.Call(hwnd)

	for _, thread := range attached {
		procAttachThreadInput.Call(current, thread, 0)
	}
}

// FlyoutForeground indica si el flyout (o un hijo suyo) es la ventana activa.
func FlyoutForeground() bool {
	hwnd := findFlyoutWindow()
	if hwnd == 0 {
		return false
	}
	foreground, _, _ := procGetForeground.Call()
	if foreground == 0 {
		return false
	}
	root, _, _ := procGetAncestor.Call(foreground, gaRoot)
	if root == 0 {
		root = foreground
	}
	return root == hwnd
}
