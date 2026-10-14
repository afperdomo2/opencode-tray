package main

import (
	"context"
	"os"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"opencode-tray/internal/shell"
)

// App es la estructura principal que cablea el backend (shell) con Wails.
type App struct {
	ctx       context.Context
	settings  *shell.SettingsStore
	autostart *shell.AutoStart
	win       *shell.WindowController
	tray      *shell.TrayHandler
}

// NewApp crea la App con el store de ajustes y el gestor de auto-inicio.
func NewApp() *App {
	// Ruta de ajustes por defecto (%APPDATA%\OpenCodeTray\settings.json).
	path, err := shell.DefaultPath()
	if err != nil {
		path = "settings.json"
	}
	// Clave de registro del auto-inicio (Windows).
	key, _ := shell.NewSystemRunKey()

	return &App{
		settings:  shell.NewSettingsStore(path),
		autostart: shell.NewAutoStart(key, os.Executable),
	}
}

// startup se ejecuta al arrancar: configura la ventana y el tray.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Controlador de ventana (mostrar/ocultar el panel).
	a.win = shell.NewWindowController(wailsWindow{ctx})

	// Arrancar el tray: clic izquierdo muestra el panel; "Salir" cierra la app.
	a.tray = shell.NewTrayHandler(shell.TrayOptions{
		Icon:    trayIcon,
		Tooltip: "OpenCode Tray",
		OnShow:  func() { a.win.Show() },
		OnQuit:  func() { wailsruntime.Quit(ctx) },
	})
	a.tray.Start()
}

// shutdown libera el icono del tray al salir.
func (a *App) shutdown(ctx context.Context) {
	if a.tray != nil {
		a.tray.End()
	}
}

// wailsWindow adapta el runtime de Wails a shell.WindowBackend.
type wailsWindow struct{ ctx context.Context }

func (w wailsWindow) Show() { wailsruntime.WindowShow(w.ctx) }
func (w wailsWindow) Hide() { wailsruntime.WindowHide(w.ctx) }

// --- Métodos expuestos al frontend ---

// GetSettings devuelve los ajustes actuales.
func (a *App) GetSettings() shell.Settings {
	st, _ := a.settings.Load()
	return st
}

// SaveSettings persiste los ajustes dados.
func (a *App) SaveSettings(st shell.Settings) error {
	return a.settings.Save(st)
}

// SetLanguage cambia y persiste el idioma de la UI.
func (a *App) SetLanguage(lang string) error {
	st, _ := a.settings.Load()
	st.Language = lang
	return a.settings.Save(st)
}

// SetAutoStart activa o desactiva el arranque con Windows (registro + ajustes).
func (a *App) SetAutoStart(enabled bool) error {
	if err := a.autostart.SetEnabled(enabled); err != nil {
		return err
	}
	st, _ := a.settings.Load()
	st.AutoStart = enabled
	return a.settings.Save(st)
}

// ToggleWindow alterna mostrar/ocultar el panel desde el tray.
func (a *App) ToggleWindow() {
	a.win.Toggle()
}
