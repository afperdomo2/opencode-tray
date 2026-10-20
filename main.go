package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"opencode-tray/internal/shell"
)

//go:embed all:frontend/dist
var assets embed.FS

// Icono de la bandeja (se usa como imagen del tray).
//
//go:embed build/windows/icon.ico
var trayIcon []byte

func main() {
	// Una sola instancia: si ya hay otra corriendo, salir en silencio.
	ok, release := shell.AcquireSingleInstanceLock("opencode-tray-instance")
	if !ok {
		return
	}
	defer release()

	app := NewApp()

	// El tray se registra en el hilo principal, antes de que Wails arranque su
	// bucle de mensajes. Si se registrara dentro de OnStartup (que Wails ejecuta
	// en otra goroutine), la ventana del tray quedaría en un hilo sin bucle de
	// mensajes y los clics/menú no funcionarían.
	app.initTray(trayIcon)

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
