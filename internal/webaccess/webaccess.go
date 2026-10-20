package webaccess

import "opencode-tray/internal/servicecontrol"

// Mode es un modo de acceso al servidor web de OpenCode.
type Mode string

const (
	ModeLocal   Mode = "local"
	ModeNetwork Mode = "network"
	ModeRemote  Mode = "remote"
)

// Kind clasifica una URL de acceso para que la UI la etiquete.
type Kind string

const (
	KindLocal  Kind = "local"
	KindLAN    Kind = "lan"
	KindRemote Kind = "remote"
)

// URLItem es una URL de acceso con su tipo.
type URLItem struct {
	URL  string `json:"url"`
	Kind Kind   `json:"kind"`
}

// State describe el modo activo y las URLs disponibles.
type State struct {
	Mode    string    `json:"mode"` // local | network | remote
	Running bool      `json:"running"`
	URLs    []URLItem `json:"urls"`
}

// serviceBackend es el contrato que web-access necesita de service-control.
type serviceBackend interface {
	Set(key, value string) error
	Get(key string) (string, error)
	RunRaw(args ...string) (string, error)
	Start() error
	Status() (servicecontrol.Status, error)
	Info() (servicecontrol.ServerInfo, error)
}

// Service gestiona modos de acceso y credenciales del servidor de OpenCode.
type Service struct {
	sc serviceBackend
}

// New crea un Service sobre un service-control (inyectable para tests).
func New(sc serviceBackend) *Service {
	return &Service{sc: sc}
}
