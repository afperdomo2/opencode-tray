package main

import (
	"embed"
	"os"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"opencode-tray/internal/shell"
)

//go:embed all:frontend/dist
var assets embed.FS

// Iconos de la bandeja: ICO en Windows y PNG en el resto (Linux/macOS), porque
// `energye/systray` decodifica la imagen con image.Decode, que no soporta ICO.
//
//go:embed build/windows/icon.ico
var trayIconICO []byte

//go:embed build/linux/icon.png
var trayIconPNG []byte

// trayIcon devuelve el icono en el formato que la bandeja sabe decodificar en
// el SO actual.
func trayIcon() []byte {
	if runtime.GOOS == "windows" {
		return trayIconICO
	}
	return trayIconPNG
}

// instanceName identifica el lock y el socket de instancia única.
const instanceName = "opencode-tray-instance"

// showPanelArg indica si esta ejecución debe abrir el panel al arrancar. Lo usa
// el .desktop del menú de aplicaciones para dar feedback inmediato.
const showPanelArg = "--show-panel"

// wantsPanel comprueba si se pidió abrir el panel en los argumentos.
func wantsPanel(args []string) bool {
	for _, arg := range args {
		if arg == showPanelArg {
			return true
		}
	}
	return false
}

func main() {
	// Una sola instancia: si ya hay otra corriendo, se le pide que muestre el
	// panel y esta ejecución sale.
	ok, release := shell.AcquireSingleInstanceLock(instanceName)
	if !ok {
		shell.RequestShow(instanceName)
		return
	}
	defer release()

	app := NewApp()
	app.showOnStart = wantsPanel(os.Args[1:])

	// El tray se registra en el hilo principal, antes de que Wails arranque su
	// bucle de mensajes. Si se registrara dentro de OnStartup (que Wails ejecuta
	// en otra goroutine), la ventana del tray quedaría en un hilo sin bucle de
	// mensajes y los clics/menú no funcionarían.
	app.initTray(trayIcon())

	// Atiende a una segunda ejecución (p. ej. clic en el icono del menú cuando
	// la app ya corre desde el auto-inicio).
	shell.ServeShowRequests(instanceName, app.showPanel)

	err := wails.Run(&options.App{
		Title:            "OpenCode Tray",
		Width:            shell.FlyoutWidth,
		Height:           shell.FlyoutMaxHeight,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 15, G: 23, B: 42, A: 1},
		// Flyout nativo: sin marco, sin resize, siempre encima y con clase Win32
		// propia para anclarlo al icono del tray.
		Frameless:     true,
		DisableResize: true,
		AlwaysOnTop:   true,
		Windows: &windows.Options{
			WindowClassName: shell.FlyoutWindowClass,
		},
		// Cerrar la ventana la minimiza al tray; no sale de la app.
		HideWindowOnClose: true,
		// Arranca oculto: solo se ve el icono de bandeja.
		StartHidden: true,
		OnStartup:   app.startup,
		OnShutdown:  app.shutdown,
		Bind:        []interface{}{app},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
