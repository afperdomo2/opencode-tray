package usage

import (
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Window describe una ventana de consumo (percent 0-100 + reset).
type Window struct {
	Percent  float64   `json:"percent"`
	ResetsAt time.Time `json:"resetsAt"`
	HasReset bool      `json:"hasReset"`
}

// Usage agrupa las tres ventanas de OpenCode Go.
type Usage struct {
	Rolling Window `json:"rolling"`
	Weekly  Window `json:"weekly"`
	Monthly Window `json:"monthly"`
}

// Level es la severidad del consumo (para el color del icono del tray).
type Level int

const (
	LevelOK Level = iota
	LevelWarn
	LevelCritical
)

// TokenStatus describe de dónde saldría el token sin ejecutar red.
type TokenStatus struct {
	Configured bool   `json:"configured"`
	Source     string `json:"source"`             // opencode-export | opencode-json | manual | none
	Provider   string `json:"provider,omitempty"` // opencode-go | opencode
}

// Service consulta el consumo de OpenCode Go.
type Service struct {
	client    *http.Client
	baseURL   string
	authFiles []string // rutas candidatas a auth.json legado (solo lectura)
	runner    Runner   // ejecuta `opencode auth export`

	manualMu    sync.RWMutex
	manualToken string // token manual (respaldo), solo en memoria

	exportMu       sync.RWMutex
	exportToken    string
	exportProvider string
}

// New crea un Service apuntando al endpoint oficial y a las credenciales del usuario.
func New() *Service {
	home, _ := os.UserHomeDir()
	files := []string{filepath.Join(home, ".local", "share", "opencode", "auth.json")}
	if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
		files = append(files, filepath.Join(dir, "opencode", "auth.json"))
	}
	if dir := os.Getenv("APPDATA"); dir != "" {
		files = append(files, filepath.Join(dir, "opencode", "auth.json"))
	}
	files = uniqueStrings(files)

	return &Service{
		client:    &http.Client{Timeout: 10 * time.Second},
		baseURL:   "https://opencode.ai",
		authFiles: files,
		runner:    execRunner{bin: "opencode"},
	}
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}
