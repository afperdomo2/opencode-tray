package main

import (
	"context"
	"os"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"opencode-tray/internal/configviewer"
	"opencode-tray/internal/servicecontrol"
	"opencode-tray/internal/shell"
	"opencode-tray/internal/usage"
	"opencode-tray/internal/webaccess"
)

// App es la estructura principal que cablea el backend con Wails.
type App struct {
	ctx       context.Context
	settings  *shell.SettingsStore
	autostart *shell.AutoStart
	win       *shell.WindowController
	tray      *shell.TrayHandler
	service   *servicecontrol.Service
	usage     *usage.Service
	webaccess *webaccess.Service
	cfgviewer *configviewer.Service
}

// NewApp crea la App con sus dependencias del backend.
func NewApp() *App {
	// Ruta de ajustes por defecto (%APPDATA%\OpenCodeTray\settings.json).
	path, err := shell.DefaultPath()
	if err != nil {
		path = "settings.json"
	}
	// Clave de registro del auto-inicio (Windows).
	key, _ := shell.NewSystemRunKey()
	svc := servicecontrol.New()

	return &App{
		settings:  shell.NewSettingsStore(path),
		autostart: shell.NewAutoStart(key, os.Executable),
		service:   svc,
		usage:     usage.New(),
		webaccess: webaccess.New(svc),
		cfgviewer: configviewer.New(),
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

// --- Ajustes y ventana ---

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

// --- Control del servicio (service-control) ---

// ServiceStatusResult expone el estado del servicio al frontend.
type ServiceStatusResult struct {
	Running bool   `json:"running"`
	URL     string `json:"url"`
}

// ServiceStatus devuelve el estado del servicio de fondo.
func (a *App) ServiceStatus() (ServiceStatusResult, error) {
	st, err := a.service.Status()
	if err != nil {
		return ServiceStatusResult{}, err
	}
	return ServiceStatusResult{Running: st.Running, URL: st.URL}, nil
}

// ServiceStart arranca el servicio.
func (a *App) ServiceStart() error { return a.service.Start() }

// ServiceStop intenta detener el servicio normalmente.
func (a *App) ServiceStop() error { return a.service.Stop() }

// ServiceForceStop detiene el servicio de forma fiable (workaround del bug).
func (a *App) ServiceForceStop() error { return a.service.ForceStop() }

// --- Consumo (usage) ---

// UsageFetch consulta el consumo de OpenCode Go (rolling/weekly/monthly).
func (a *App) UsageFetch() (usage.Usage, error) { return a.usage.Fetch() }

// UsageSetManualToken fija el token manual de Go (desde Ajustes).
func (a *App) UsageSetManualToken(tok string) { a.usage.SetManualToken(tok) }

// UsageClearManualToken elimina el token manual (vuelve a auth.json).
func (a *App) UsageClearManualToken() { a.usage.ClearManualToken() }

// --- Acceso web (web-access) ---

// WebAccessApplyMode aplica un modo: "local", "network" o "remote".
func (a *App) WebAccessApplyMode(mode string) error {
	return a.webaccess.ApplyMode(webaccess.Mode(mode))
}

// WebAccessURLs devuelve las URLs de acceso (IPs LAN + puerto).
func (a *App) WebAccessURLs() ([]string, error) { return a.webaccess.ComputeURLs() }

// WebAccessUsername devuelve el usuario del servidor (opencode).
func (a *App) WebAccessUsername() string { return a.webaccess.Username() }

// WebAccessCurrentPassword lee la contraseña actual del servicio.
func (a *App) WebAccessCurrentPassword() (string, error) { return a.webaccess.CurrentPassword() }

// WebAccessChangePassword cambia la contraseña y reinicia el servicio.
func (a *App) WebAccessChangePassword(newPw string) error {
	return a.webaccess.ChangePassword(newPw)
}

// --- Config (config-viewer) ---

// ConfigListMCP lista los servidores MCP con su estado de conexión.
func (a *App) ConfigListMCP() ([]configviewer.MCPServer, error) { return a.cfgviewer.ListMCP() }

// ConfigJSON lee el config global de OpenCode con secretos enmascarados.
func (a *App) ConfigJSON() (string, error) { return a.cfgviewer.ConfigJSON() }

// ConfigOpenInEditor abre el config global en el editor.
func (a *App) ConfigOpenInEditor() error {
	p, err := a.cfgviewer.ConfigPath()
	if err != nil {
		return err
	}
	return a.cfgviewer.OpenInEditor(p)
}

// ConfigOpenDir abre la carpeta de configuración de OpenCode.
func (a *App) ConfigOpenDir() error { return a.cfgviewer.OpenConfigDir() }
