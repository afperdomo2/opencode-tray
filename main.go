package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

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

	err := wails.Run(&options.App{
		Title:            "OpenCode Tray",
		Width:            380,
		Height:           520,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
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
