package shell

import "github.com/energye/systray"

// TrayOptions define el icono y los callbacks del tray.
type TrayOptions struct {
	Icon    []byte
	Tooltip string
	OnShow  func()
	OnQuit  func()
}

// TrayHandler gestiona el icono y el menú de la bandeja del sistema.
type TrayHandler struct {
	opts  TrayOptions
	start func()
	end   func()
}

// NewTrayHandler registra el tray con un bucle externo (integrable con Wails).
// Llamar Start() cuando la app haya arrancado y End() al cerrar.
func NewTrayHandler(opts TrayOptions) *TrayHandler {
	h := &TrayHandler{opts: opts}
	start, end := systray.RunWithExternalLoop(h.onReady, func() {})
	h.start, h.end = start, end
	return h
}

func (h *TrayHandler) onReady() {
	systray.SetIcon(h.opts.Icon)
	systray.SetTooltip(h.opts.Tooltip)

	// Clic izquierdo en el icono = mostrar el panel.
	systray.SetOnClick(func(systray.IMenu) {
		if h.opts.OnShow != nil {
			h.opts.OnShow()
		}
	})

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
