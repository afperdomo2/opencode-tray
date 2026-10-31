package shell

import (
	"runtime"
	"sync"

	"github.com/energye/systray"
)

// TrayOptions define el icono y los callbacks del tray.
type TrayOptions struct {
	Icon    []byte
	Tooltip string
	// OnClick es el clic izquierdo: alterna el panel.
	OnClick func()
	// OnShow es el ítem de menú "Mostrar panel": siempre lo abre.
	OnShow func()
	OnQuit func()
}

// TrayHandler gestiona el icono y el menú de la bandeja del sistema.
type TrayHandler struct {
	opts       TrayOptions
	start      func()
	end        func()
	setTooltip func(string)

	// status replica el tooltip como primera fila del menú en los escritorios
	// que no muestran el tooltip nativo (GNOME/AppIndicator).
	mu     sync.Mutex
	status *systray.MenuItem
}

// NewTrayHandler registra el tray con un bucle externo (integrable con Wails).
// Llamar Start() cuando la app haya arrancado y End() al cerrar.
//
// Importante: debe llamarse desde el hilo principal antes de wails.Run. Wails
// ejecuta OnStartup en otra goroutine; si el tray se registra allí, la ventana
// de la bandeja queda en un hilo sin bucle de mensajes y los clics y el menú
// contextual dejan de funcionar.
func NewTrayHandler(opts TrayOptions) *TrayHandler {
	h := &TrayHandler{opts: opts, setTooltip: systray.SetTooltip}
	start, end := systray.RunWithExternalLoop(h.onReady, func() {})
	h.start, h.end = start, end
	return h
}

func (h *TrayHandler) onReady() {
	systray.SetIcon(h.opts.Icon)
	systray.SetTooltip(h.opts.Tooltip)

	// Clic izquierdo en el icono = alternar el panel.
	systray.SetOnClick(func(systray.IMenu) {
		if h.opts.OnClick != nil {
			h.opts.OnClick()
		}
	})

	// GNOME/AppIndicator no dibuja el tooltip del icono: el consumo se replica
	// como primera fila del menú, deshabilitada y actualizada por SetTooltip.
	if runtime.GOOS != "windows" {
		status := systray.AddMenuItem("OpenCode Tray", "Consumo de OpenCode Go")
		status.Disable()
		h.mu.Lock()
		h.status = status
		h.mu.Unlock()
		systray.AddSeparator()
	}

	mShow := systray.AddMenuItem("Mostrar panel", "Abre el panel de OpenCode Tray")
	mShow.Click(func() {
		if h.opts.OnShow != nil {
			h.opts.OnShow()
		}
	})

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Salir", "Cierra OpenCode Tray")
	mQuit.Click(func() {
		if h.opts.OnQuit != nil {
			h.opts.OnQuit()
		}
		systray.Quit()
	})
}

// Start arranca el tray sin bloquear.
func (h *TrayHandler) Start() {
	if h.start != nil {
		h.start()
	}
}

// End detiene el tray al cerrar la app.
func (h *TrayHandler) End() {
	if h.end != nil {
		h.end()
	}
}

// SetTooltip actualiza el texto del tooltip nativo y, fuera de Windows, la
// primera fila del menú (los escritorios GNOME no muestran el tooltip).
func (h *TrayHandler) SetTooltip(text string) {
	if text == "" {
		return
	}
	if h.setTooltip != nil {
		h.setTooltip(text)
	}
	h.mu.Lock()
	status := h.status
	h.mu.Unlock()
	if status != nil {
		status.SetTitle(text)
	}
}
