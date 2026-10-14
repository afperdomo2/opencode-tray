package usage

import (
	"net/http"
	"os"
	"path/filepath"
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

// Service consulta el consumo de OpenCode Go.
type Service struct {
	client      *http.Client
	baseURL     string
	authFile    string // ruta a auth.json (para leer el token de opencode-go)
	manualToken string // token manual que sobreescribe el de auth.json
}

// New crea un Service apuntando al endpoint oficial y al auth.json del usuario.
func New() *Service {
	home, _ := os.UserHomeDir()
	return &Service{
		client:   &http.Client{Timeout: 10 * time.Second},
		baseURL:  "https://opencode.ai",
		authFile: filepath.Join(home, ".local", "share", "opencode", "auth.json"),
	}
}
