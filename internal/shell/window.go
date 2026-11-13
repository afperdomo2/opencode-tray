package shell

import (
	"sync"
	"sync/atomic"
)

// WindowBackend abstrae la ventana para poder testear la lógica sin Wails.
type WindowBackend interface {
	Show()
	Hide()
	Resize(width, height int)
	Place(anchor Point, width, height int)
	Focus()
}

// WindowController gestiona visibilidad, tamaño y anclaje del panel.
type WindowController struct {
	backend WindowBackend
	visible atomic.Bool

	mu     sync.Mutex
	width  int
	height int
	anchor Point
}

// NewWindowController crea un controlador sobre el backend dado.
func NewWindowController(b WindowBackend) *WindowController {
	return &WindowController{
		backend: b,
		width:   FlyoutWidth,
		height:  FlyoutMaxHeight,
	}
}

// Show muestra la ventana con el tamaño actual.
func (w *WindowController) Show() {
	w.ShowAt(Point{})
}

// ShowAt muestra la ventana anclada al punto dado.
func (w *WindowController) ShowAt(anchor Point) {
	w.mu.Lock()
	w.anchor = anchor
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
	w.visible.Store(false)
	w.backend.Hide()
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
