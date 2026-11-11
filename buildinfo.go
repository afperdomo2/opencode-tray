package main

import (
	_ "embed"
	"encoding/json"
	"runtime"
	"strings"
)

// wailsConfig es la configuración del proyecto embebida. De aquí sale la
// versión para que coincida siempre con los metadatos del .exe y del instalador
// (Info.productVersion), incluso en un `wails build` directo.
//
//go:embed wails.json
var wailsConfig []byte

// Inyectados por los scripts de build con -ldflags. Wails compila con
// -buildvcs=false, así que Go no puede leer el commit por su cuenta.
var (
	buildCommit string
	buildDate   string
	buildDirty  string
)

// fallbackVersion se muestra si wails.json no se puede interpretar.
const fallbackVersion = "dev"

// shortCommitLen es el largo del hash corto de git.
const shortCommitLen = 7

// VersionInfo describe la versión y la compilación que se está ejecutando.
type VersionInfo struct {
	Version  string `json:"version"`
	Commit   string `json:"commit"`
	Date     string `json:"date"`
	Dirty    bool   `json:"dirty"`
	Platform string `json:"platform"`
}

// GetAppVersion devuelve la versión y la compilación al frontend.
func (a *App) GetAppVersion() VersionInfo {
	return versionInfo(wailsConfig, buildCommit, buildDate, buildDirty)
}

// projectVersion extrae Info.productVersion de wails.json.
func projectVersion(config []byte) string {
	var cfg struct {
		Info struct {
			ProductVersion string `json:"productVersion"`
		} `json:"Info"`
	}
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fallbackVersion
	}
	version := strings.TrimSpace(cfg.Info.ProductVersion)
	if version == "" {
		return fallbackVersion
	}
	return version
}

// versionInfo compone el contrato que consume la UI. Es una función pura para
// poder probar los valores inyectados sin compilar con ldflags.
func versionInfo(config []byte, commit, date, dirty string) VersionInfo {
	return VersionInfo{
		Version:  projectVersion(config),
		Commit:   shortCommit(commit),
		Date:     strings.TrimSpace(date),
		Dirty:    strings.EqualFold(strings.TrimSpace(dirty), "true"),
		Platform: runtime.GOOS + "/" + runtime.GOARCH,
	}
}

// shortCommit recorta el hash de git a siete caracteres (o lo deja tal cual).
func shortCommit(commit string) string {
	commit = strings.TrimSpace(commit)
	if len(commit) > shortCommitLen {
		return commit[:shortCommitLen]
	}
	return commit
}
