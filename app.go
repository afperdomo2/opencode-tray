package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"opencode-tray/internal/configviewer"
	"opencode-tray/internal/notifications"
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
	notifier  *notifications.Service
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

	app := &App{
		settings:  shell.NewSettingsStore(path),
		autostart: shell.NewAutoStart(key, os.Executable),
		service:   svc,
		usage:     usage.New(),
		webaccess: webaccess.New(svc),
		cfgviewer: configviewer.New(),
		notifier:  notifications.New(),
	}
	app.loadManualUsageToken()
	return app
}

// loadManualUsageToken descifra el token manual guardado, si existe. Si DPAPI
// falla (otro usuario/máquina o blob corrupto), se ignora sin romper el arranque.
func (a *App) loadManualUsageToken() {
	st, err := a.settings.Load()
	if err != nil || st.UsageToken == "" {
		return
	}
	tok, err := shell.UnprotectSecret(st.UsageToken)
	if err != nil || tok == "" {
		return
	}
	a.usage.SetManualToken(tok)
}

// startup se ejecuta al arrancar: configura la ventana y el tray.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Controlador de ventana (mostrar/ocultar el panel).
	a.win = shell.NewWindowController(wailsWindow{ctx})

	// Sacar la ventana de la barra de tareas y de Alt+Tab.
	shell.PrepareFlyoutWindow()

	// Envío de notificaciones DESHABILITADO (Fase 4): inyectar aquí
	// wailsruntime.SendNotification (go-toast en Windows) bloqueaba el bucle de
	// mensajes de energye/systray y dejaba la bandeja muda (ni clic izquierdo ni
	// menú contextual). El notifier conserva su sender no-op; la lógica de
	// decisión queda probada y lista para reactivarse por una vía segura.

	// Arrancar el tray: clic izquierdo alterna el panel anclado al icono;
	// "Mostrar panel" siempre lo abre; "Salir" cierra la app.
	a.tray = shell.NewTrayHandler(shell.TrayOptions{
		Icon:    trayIcon,
		Tooltip: "OpenCode Tray",
		OnClick: func() { a.win.ToggleAt(shell.CursorPosition()) },
		OnShow:  func() { a.win.ShowAt(shell.CursorPosition()) },
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

func (w wailsWindow) Resize(width, height int) {
	wailsruntime.WindowSetSize(w.ctx, width, height)
}

func (w wailsWindow) Place(anchor shell.Point, width, height int) {
	shell.PositionFlyout(anchor, width, height)
}

func (w wailsWindow) Focus() { shell.FocusFlyout() }

func (w wailsWindow) IsForeground() bool { return shell.FlyoutForeground() }

// --- Ajustes y ventana ---

// GetSettings devuelve los ajustes actuales sin secretos.
func (a *App) GetSettings() shell.Settings {
	st, _ := a.settings.Load()
	return st.Public()
}

// SaveSettings persiste los ajustes dados preservando el token cifrado, que
// nunca viaja al frontend.
func (a *App) SaveSettings(st shell.Settings) error {
	current, _ := a.settings.Load()
	current.Language = st.Language
	current.AutoStart = st.AutoStart
	current.RefreshMin = st.RefreshMin
	return a.settings.Save(current)
}

// SetRefreshMin cambia y persiste el intervalo de refresco del consumo.
func (a *App) SetRefreshMin(min int) error {
	st, _ := a.settings.Load()
	st.RefreshMin = min
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
	a.win.ToggleAt(shell.CursorPosition())
}

// SetFlyoutHeight ajusta el alto del panel según su contenido.
func (a *App) SetFlyoutHeight(height int) {
	if a.win != nil {
		a.win.SetHeight(height)
	}
}

// HideWindow oculta el panel (por ejemplo, con Escape).
func (a *App) HideWindow() {
	if a.win != nil {
		a.win.Hide()
	}
}

// WindowBlurred oculta el panel al perder el foco, con guarda anti-carrera.
func (a *App) WindowBlurred() {
	if a.win != nil {
		a.win.HideFromBlur()
	}
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
func (a *App) UsageFetch() (usage.Usage, error) {
	u, err := a.usage.Fetch()
	if err == nil && a.tray != nil {
		a.tray.SetTooltip(usageTooltip(u))
	}
	return u, err
}

// usageTooltip resume los tres porcentajes para el tooltip nativo del tray.
func usageTooltip(u usage.Usage) string {
	return fmt.Sprintf(
		"OpenCode · 5h %d%% · 7d %d%% · 30d %d%%",
		int(math.Round(u.Rolling.Percent)),
		int(math.Round(u.Weekly.Percent)),
		int(math.Round(u.Monthly.Percent)),
	)
}

// UsageTokenStatus indica de dónde saldría el token (sin ejecutar red).
func (a *App) UsageTokenStatus() usage.TokenStatus { return a.usage.TokenStatus() }

// UsageReloadToken invalida el caché y relee el token vigente desde OpenCode.
func (a *App) UsageReloadToken() (usage.TokenStatus, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.usage.ReloadToken(ctx)
}

// UsageSetManualToken valida, fija y persiste cifrado el token manual.
func (a *App) UsageSetManualToken(tok string) error {
	tok = strings.TrimSpace(tok)
	if tok == "" {
		return errors.New("la API key no puede estar vacía")
	}
	err := a.usage.ValidateToken(tok)
	if err != nil && (errors.Is(err, usage.ErrInvalidToken) || errors.Is(err, usage.ErrNoEntitlement)) {
		return err
	}
	// Si falla la red no bloqueamos el guardado: el próximo fetch reintenta.
	blob, perr := shell.ProtectSecret(tok)
	if perr != nil {
		return perr
	}
	st, _ := a.settings.Load()
	st.UsageToken = blob
	if err := a.settings.Save(st); err != nil {
		return err
	}
	a.usage.SetManualToken(tok)
	return nil
}

// UsageClearManualToken elimina el token manual (en disco y en memoria).
func (a *App) UsageClearManualToken() error {
	st, _ := a.settings.Load()
	st.UsageToken = ""
	if err := a.settings.Save(st); err != nil {
		return err
	}
	a.usage.ClearManualToken()
	return nil
}

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

// --- Notificaciones (notifications) ---

// evaluateNotifier hace una evaluación inmediata del consumo y el servicio.
// Nota: el envío de toasts está deshabilitado (ver startup); esto solo evalúa
// la lógica con el sender no-op, por lo que no dispara nada hacia la bandeja.
func (a *App) evaluateNotifier() {
	u, err := a.usage.Fetch()
	if err != nil {
		return
	}
	svcStatus, svcErr := a.service.Status()
	running := svcErr == nil && svcStatus.Running
	a.notifier.Evaluate(u, running)
}

// NotifyEvaluate fuerza una evaluación inmediata (manual).
func (a *App) NotifyEvaluate() { a.evaluateNotifier() }

// NotifySetEnabled habilita/deshabilita un tipo de alerta:
// "burnrate", "exhaustion" o "service".
func (a *App) NotifySetEnabled(kind string, on bool) { a.notifier.SetEnabled(kind, on) }
