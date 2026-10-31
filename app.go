package main

import (
	"context"
	"errors"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

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

	// Caché del consumo compartida entre el panel y el tooltip del tray.
	usageMu    sync.Mutex
	usageCache usage.Usage
	usageAt    time.Time
	usageOK    bool

	// Último intento (éxito o error) para no martillar la API en cada tick.
	usageAttemptAt time.Time

	// Watcher que mantiene el tooltip al día aunque el panel nunca se abra.
	watcherStop chan struct{}

	// Caché de corta duración del estado del servicio. Resumen pide
	// ServiceStatus y WebAccessState en paralelo: ambos comparten una sola
	// lectura del CLI (singleflight + TTL).
	statusMu     sync.Mutex
	statusCache  servicecontrol.Status
	statusAt     time.Time
	statusOK     bool
	statusGen    uint64
	statusFlight *statusFlight
	statusTTL    time.Duration

	// Reloj inyectable para tests del tooltip y la caché.
	now func() time.Time
}

// statusFlight es una lectura de estado en curso que otros llamadores esperan.
type statusFlight struct {
	done  chan struct{}
	value servicecontrol.Status
	err   error
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
		statusTTL: time.Second,
		now:       time.Now,
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

// initTray registra el icono y el menú en el hilo principal. Debe llamarse
// antes de wails.Run para que el bucle de mensajes de Wails procese los clics.
func (a *App) initTray(icon []byte) {
	a.tray = shell.NewTrayHandler(shell.TrayOptions{
		Icon:    icon,
		Tooltip: "OpenCode Tray",
		OnClick: func() {
			if a.win != nil {
				a.win.ToggleAt(shell.CursorPosition())
			}
		},
		OnShow: func() {
			if a.win != nil {
				a.win.ShowAt(shell.CursorPosition())
			}
		},
		OnQuit: func() {
			if a.ctx != nil {
				wailsruntime.Quit(a.ctx)
				return
			}
			os.Exit(0)
		},
	})
	a.tray.Start()
}

// startup se ejecuta al arrancar: configura la ventana y el watcher.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Controlador de ventana (mostrar/ocultar el panel).
	a.win = shell.NewWindowController(wailsWindow{ctx})

	// Sacar la ventana de la barra de tareas y de Alt+Tab.
	shell.PrepareFlyoutWindow()

	// En Linux, Wails ignora StartHidden: window.go llama a Hide y acto seguido
	// a gtk_widget_show_all, así que el panel aparece al arrancar. Se oculta
	// aquí para que la app nazca solo con el icono de bandeja, como en Windows.
	if runtime.GOOS != "windows" {
		wailsruntime.WindowHide(ctx)
	}

	// Envío de notificaciones DESHABILITADO (Fase 4): inyectar aquí
	// wailsruntime.SendNotification (go-toast en Windows) bloqueaba el bucle de
	// mensajes de energye/systray y dejaba la bandeja muda (ni clic izquierdo ni
	// menú contextual). El notifier conserva su sender no-op; la lógica de
	// decisión queda probada y lista para reactivarse por una vía segura.

	// Aunque el panel no se abra, el tooltip debe mostrar el consumo y la
	// cuenta atrás del reinicio. Arranca tras el tray para que SetTooltip
	// encuentre el icono listo.
	a.startUsageWatcher()
}

// shutdown libera el icono del tray al salir.
func (a *App) shutdown(ctx context.Context) {
	a.stopUsageWatcher()
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
	positionFlyout(w.ctx, anchor, width, height)
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
	if err := a.settings.Save(st); err != nil {
		return err
	}
	// El tooltip es nativo y también debe cambiar de idioma al instante.
	a.usageMu.Lock()
	a.applyTooltipLocked()
	a.usageMu.Unlock()
	return nil
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

// cachedStatus lee el estado del servicio con TTL corto y singleflight: las
// llamadas concurrentes comparten una sola ejecución de `opencode service
// status`. La caché se invalida tras cada mutación del servicio.
func (a *App) cachedStatus() (servicecontrol.Status, error) {
	a.statusMu.Lock()
	if a.statusOK && a.clock().Sub(a.statusAt) < a.statusTTL {
		value := a.statusCache
		a.statusMu.Unlock()
		return value, nil
	}
	if flight := a.statusFlight; flight != nil {
		a.statusMu.Unlock()
		<-flight.done
		return flight.value, flight.err
	}
	gen := a.statusGen
	flight := &statusFlight{done: make(chan struct{})}
	a.statusFlight = flight
	a.statusMu.Unlock()

	value, err := a.service.Status()

	a.statusMu.Lock()
	flight.value, flight.err = value, err
	if gen == a.statusGen && err == nil {
		a.statusCache = value
		a.statusAt = a.clock()
		a.statusOK = true
	}
	if a.statusFlight == flight {
		a.statusFlight = nil
	}
	close(flight.done)
	a.statusMu.Unlock()
	return value, err
}

// invalidateStatusCache descarta el estado cacheado (tras start/stop/config).
func (a *App) invalidateStatusCache() {
	a.statusMu.Lock()
	a.statusGen++
	a.statusOK = false
	a.statusFlight = nil
	a.statusMu.Unlock()
}

// ServiceStatus devuelve el estado del servicio de fondo.
func (a *App) ServiceStatus() (ServiceStatusResult, error) {
	st, err := a.cachedStatus()
	if err != nil {
		return ServiceStatusResult{}, err
	}
	return ServiceStatusResult{Running: st.Running, URL: st.URL}, nil
}

// ServiceStart arranca el servicio.
func (a *App) ServiceStart() error { return a.runServiceAction(a.service.Start) }

// ServiceStop intenta detener el servicio normalmente.
func (a *App) ServiceStop() error { return a.runServiceAction(a.service.Stop) }

// ServiceForceStop detiene el servicio de forma fiable (workaround del bug).
func (a *App) ServiceForceStop() error { return a.runServiceAction(a.service.ForceStop) }

// runServiceAction invalida la caché antes y después de una mutación: el
// estado anterior ya no es válido y el nuevo debe leerse del CLI.
func (a *App) runServiceAction(fn func() error) error {
	a.invalidateStatusCache()
	err := fn()
	a.invalidateStatusCache()
	return err
}

// --- Consumo (usage) ---

const usageCacheTTL = 60 * time.Second

// UsageFetch consulta el consumo de OpenCode Go (rolling/weekly/monthly).
// Comparte caché con el watcher del tray para no duplicar peticiones.
func (a *App) UsageFetch() (usage.Usage, error) {
	a.usageMu.Lock()
	defer a.usageMu.Unlock()

	if a.usageOK && a.clock().Sub(a.usageAt) < usageCacheTTL {
		a.applyTooltipLocked()
		return a.usageCache, nil
	}
	return a.fetchUsageLocked(false)
}

// startUsageWatcher arranca el refresco periódico del tooltip.
func (a *App) startUsageWatcher() {
	if a.watcherStop != nil {
		return
	}
	stop := make(chan struct{})
	a.watcherStop = stop
	go a.usageLoop(stop)
}

// stopUsageWatcher detiene el refresco periódico (idempotente).
func (a *App) stopUsageWatcher() {
	if a.watcherStop == nil {
		return
	}
	close(a.watcherStop)
	a.watcherStop = nil
}

// usageLoop hace una lectura inicial tras el arranque y luego revisa cada
// minuto: recalcula la cuenta atrás y hace fetch cuando toca por RefreshMin.
func (a *App) usageLoop(stop <-chan struct{}) {
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case <-stop:
		return
	case <-timer.C:
	}

	a.backgroundRefresh(true)

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			a.backgroundRefresh(false)
		}
	}
}

// backgroundRefresh actualiza el tooltip con los datos cacheados y hace fetch
// solo cuando ha pasado el intervalo configurado.
func (a *App) backgroundRefresh(retryToken bool) {
	a.usageMu.Lock()
	defer a.usageMu.Unlock()

	st, _ := a.settings.Load()
	refresh := time.Duration(st.RefreshMin) * time.Minute
	if refresh < time.Minute {
		refresh = time.Minute
	}
	if a.usageOK && a.clock().Sub(a.usageAt) < refresh {
		a.applyTooltipLocked()
		return
	}
	if !a.usageAttemptAt.IsZero() && a.clock().Sub(a.usageAttemptAt) < refresh {
		a.applyTooltipLocked()
		return
	}
	_, _ = a.fetchUsageLocked(retryToken)
}

// fetchUsageLocked consulta el consumo y actualiza caché y tooltip. Si el
// llamador lo permite, reintenta una vez releyendo el token de OpenCode.
func (a *App) fetchUsageLocked(retryToken bool) (usage.Usage, error) {
	defer func() { a.usageAttemptAt = a.clock() }()

	u, err := a.usage.Fetch()
	if err != nil && retryToken && (errors.Is(err, usage.ErrNoToken) || errors.Is(err, usage.ErrInvalidToken)) {
		if _, reloadErr := a.usage.ReloadToken(a.context()); reloadErr == nil {
			u, err = a.usage.Fetch()
		}
	}
	if err != nil {
		// Sin datos nuevos se conserva el tooltip anterior.
		return usage.Usage{}, err
	}

	a.usageCache = u
	a.usageAt = a.clock()
	a.usageOK = true
	a.applyTooltipLocked()
	return u, nil
}

// applyTooltipLocked pinta el tooltip con el último consumo cacheado.
func (a *App) applyTooltipLocked() {
	if a.tray == nil || !a.usageOK {
		return
	}
	st, _ := a.settings.Load()
	a.tray.SetTooltip(usageTooltip(a.usageCache, st.Language, a.clock()))
}

// usageTooltip construye el texto del tooltip (barra + reinicio humanizado).
func usageTooltip(u usage.Usage, lang string, now time.Time) string {
	return usage.Tooltip(u, usage.TooltipOptions{
		Language: lang,
		Icons:    usage.EmojiIcons,
		Now:      now,
	})
}

// clock devuelve el reloj de la app (inyectable en tests).
func (a *App) clock() time.Time {
	if a.now == nil {
		return time.Now()
	}
	return a.now()
}

// context devuelve el contexto de Wails o uno de fondo si aún no arrancó.
func (a *App) context() context.Context {
	if a.ctx == nil {
		return context.Background()
	}
	return a.ctx
}

// UsageTokenStatus indica de dónde saldría el token (sin ejecutar red).
func (a *App) UsageTokenStatus() usage.TokenStatus { return a.usage.TokenStatus() }

// UsageReloadToken invalida el caché y relee el token vigente desde OpenCode.
func (a *App) UsageReloadToken() (usage.TokenStatus, error) {
	a.usageMu.Lock()
	defer a.usageMu.Unlock()
	status, err := a.usage.ReloadToken(a.context())
	if err == nil {
		a.invalidateUsageLocked()
	}
	return status, err
}

// UsageSetManualToken valida, fija y persiste cifrado el token manual.
func (a *App) UsageSetManualToken(tok string) error {
	tok = strings.TrimSpace(tok)
	if tok == "" {
		return errors.New("la API key no puede estar vacía")
	}
	a.usageMu.Lock()
	defer a.usageMu.Unlock()

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
	a.invalidateUsageLocked()
	return nil
}

// UsageClearManualToken elimina el token manual (en disco y en memoria).
func (a *App) UsageClearManualToken() error {
	a.usageMu.Lock()
	defer a.usageMu.Unlock()

	st, _ := a.settings.Load()
	st.UsageToken = ""
	if err := a.settings.Save(st); err != nil {
		return err
	}
	a.usage.ClearManualToken()
	a.invalidateUsageLocked()
	return nil
}

// invalidateUsageLocked fuerza el próximo fetch tras cambiar el token.
func (a *App) invalidateUsageLocked() {
	a.usageAt = time.Time{}
	a.usageOK = false
	a.usageAttemptAt = time.Time{}
}

// --- Acceso web (web-access) ---

// WebAccessApplyMode aplica un modo: "local", "network" o "remote".
func (a *App) WebAccessApplyMode(mode string) error {
	a.invalidateStatusCache()
	err := a.webaccess.ApplyMode(webaccess.Mode(mode))
	a.invalidateStatusCache()
	return err
}

// WebAccessState devuelve el modo activo, el estado del servicio y las URLs.
// Reutiliza el estado cacheado para no repetir `service status`.
func (a *App) WebAccessState() (webaccess.State, error) {
	st, _ := a.cachedStatus()
	return a.webaccess.StateWithStatus(st)
}

// WebAccessOpenURL abre una URL de acceso validada en el navegador del sistema.
func (a *App) WebAccessOpenURL(raw string) error {
	ok, err := a.webaccess.IsAccessURL(raw)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("URL de acceso no válida")
	}
	if a.ctx == nil {
		return errors.New("la app aún no está inicializada")
	}
	wailsruntime.BrowserOpenURL(a.ctx, raw)
	return nil
}

// WebAccessUsername devuelve el usuario del servidor (opencode).
func (a *App) WebAccessUsername() string { return a.webaccess.Username() }

// WebAccessCurrentPassword lee la contraseña actual del servicio.
func (a *App) WebAccessCurrentPassword() (string, error) { return a.webaccess.CurrentPassword() }

// WebAccessChangePassword cambia la contraseña y reinicia el servicio.
func (a *App) WebAccessChangePassword(newPw string) error {
	a.invalidateStatusCache()
	err := a.webaccess.ChangePassword(newPw)
	a.invalidateStatusCache()
	return err
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
