package shell

import "sync/atomic"

// WindowBackend abstrae mostrar/ocultar la ventana para poder testear la
// lógica de estado sin depender de Wails.
type WindowBackend interface {
	Show()
	Hide()
}

// WindowController gestiona el estado de visibilidad del panel.
type WindowController struct {
	backend WindowBackend
	visible atomic.Bool
}

// NewWindowController crea un controlador sobre el backend dado.
func NewWindowController(b WindowBackend) *WindowController {
	return &WindowController{backend: b}
}

// Show muestra la ventana y marca el estado como visible.
func (w *WindowController) Show() {
	w.visible.Store(true)
	w.backend.Show()
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
	} else {
		w.Show()
	}
}

// IsVisible devuelve el estado actual de visibilidad.
func (w *WindowController) IsVisible() bool {
	return w.visible.Load()
}
