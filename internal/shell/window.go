package shell

import (
	"sync"
	"sync/atomic"
	"time"
)

// blurGuard evita reabrir el panel cuando el clic en el icono del tray provocó
// el blur que lo ocultó.
const blurGuard = 300 * time.Millisecond

// showBlurGuard ignora el blur transitorio que Windows genera al cambiar el
// foco desde el tray al flyout recién mostrado.
const showBlurGuard = 400 * time.Millisecond

// stopTimer abstrae time.Timer para inyectar temporizadores en los tests.
type stopTimer interface {
	Stop() bool
}

type timerFactory func(time.Duration, func()) stopTimer

// WindowBackend abstrae la ventana para poder testear la lógica sin Wails.
type WindowBackend interface {
	Show()
	Hide()
	Resize(width, height int)
	Place(anchor Point, width, height int)
	Focus()
	IsForeground() bool
}

// WindowController gestiona visibilidad, tamaño y anclaje del panel.
type WindowController struct {
	backend WindowBackend
	visible atomic.Bool
	now     func() time.Time

	mu        sync.Mutex
	width     int
	height    int
	anchor    Point
	lastBlur  time.Time
	lastShow  time.Time
	showSeq   uint64
	blurTimer stopTimer
	after     timerFactory
}

// NewWindowController crea un controlador sobre el backend dado.
func NewWindowController(b WindowBackend) *WindowController {
	return newWindowControllerWithClock(b, time.Now)
}

func newWindowControllerWithClock(b WindowBackend, now func() time.Time) *WindowController {
	return newWindowControllerWithDeps(b, now, func(d time.Duration, f func()) stopTimer {
		return time.AfterFunc(d, f)
	})
}

func newWindowControllerWithDeps(b WindowBackend, now func() time.Time, after timerFactory) *WindowController {
	return &WindowController{
		backend: b,
		now:     now,
		after:   after,
		width:   FlyoutWidth,
		height:  FlyoutMaxHeight,
	}
}

// stopBlurTimerLocked cancela la comprobación pendiente. Requiere w.mu.
func (w *WindowController) stopBlurTimerLocked() {
	if w.blurTimer != nil {
		w.blurTimer.Stop()
		w.blurTimer = nil
	}
}

// Show muestra la ventana con el tamaño actual.
func (w *WindowController) Show() {
	w.ShowAt(Point{})
}

// ShowAt muestra la ventana anclada al punto dado.
func (w *WindowController) ShowAt(anchor Point) {
	w.mu.Lock()
	w.stopBlurTimerLocked()
	w.anchor = anchor
	w.lastBlur = time.Time{}
	w.lastShow = w.now()
	w.showSeq++
	width, height := w.width, w.height
	w.mu.Unlock()

	// Mover primero (oculto) para que Wails aplique el DPI del monitor destino;
	// luego redimensionar y volver a anclar con el tamaño físico resultante.
	w.backend.Place(anchor, width, height)
	w.backend.Resize(width, height)
	w.backend.Place(anchor, width, height)
	w.visible.Store(true)
	w.backend.Show()
	w.backend.Focus()
}

// Hide oculta la ventana y marca el estado como no visible.
func (w *WindowController) Hide() {
	w.mu.Lock()
	w.stopBlurTimerLocked()
	w.showSeq++
	w.mu.Unlock()

	w.visible.Store(false)
	w.backend.Hide()
}

// HideFromBlur oculta la ventana cuando pierde el foco. Si el blur llega justo
// después de abrir, espera a que termine el cambio de foco y solo cierra si el
// flyout ya no está en primer plano.
func (w *WindowController) HideFromBlur() {
	if !w.visible.Load() {
		return
	}

	w.mu.Lock()
	now := w.now()
	if now.Sub(w.lastShow) < showBlurGuard {
		if w.blurTimer == nil {
			delay := showBlurGuard - now.Sub(w.lastShow)
			seq := w.showSeq
			w.blurTimer = w.after(delay, func() { w.hideIfUnfocused(seq) })
		}
		w.mu.Unlock()
		return
	}
	w.lastBlur = now
	w.mu.Unlock()
	w.Hide()
}

// hideIfUnfocused cierra el panel si la comprobación sigue siendo válida y la
// ventana perdió realmente el foco.
func (w *WindowController) hideIfUnfocused(seq uint64) {
	w.mu.Lock()
	if seq != w.showSeq {
		w.mu.Unlock()
		return
	}
	w.blurTimer = nil
	w.mu.Unlock()

	if !w.visible.Load() || w.backend.IsForeground() {
		return
	}

	w.mu.Lock()
	w.lastBlur = w.now()
	w.mu.Unlock()
	w.Hide()
}

// Toggle alterna la visibilidad.
func (w *WindowController) Toggle() {
	if w.visible.Load() {
		w.Hide()
		return
	}
	w.Show()
}

// ToggleAt alterna la visibilidad anclando el panel al punto dado.
func (w *WindowController) ToggleAt(anchor Point) {
	if w.visible.Load() {
		w.Hide()
		return
	}

	w.mu.Lock()
	recentBlur := !w.lastBlur.IsZero() && w.now().Sub(w.lastBlur) < blurGuard
	w.lastBlur = time.Time{}
	w.mu.Unlock()
	if recentBlur {
		return
	}
	w.ShowAt(anchor)
}

// SetHeight actualiza el alto del panel, lo acota y lo reancla si está visible.
func (w *WindowController) SetHeight(height int) {
	w.mu.Lock()
	height = ClampFlyoutHeight(height, FlyoutMaxHeight+2*flyoutMargin)
	if height == w.height {
		w.mu.Unlock()
		return
	}
	w.height = height
	width, anchor := w.width, w.anchor
	w.mu.Unlock()

	w.backend.Resize(width, height)
	if w.visible.Load() {
		w.backend.Place(anchor, width, height)
	}
}

// IsVisible devuelve el estado actual de visibilidad.
func (w *WindowController) IsVisible() bool {
	return w.visible.Load()
}
