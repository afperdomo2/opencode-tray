package webaccess

// Mode es un modo de acceso al servidor web de OpenCode.
type Mode string

const (
	ModeLocal   Mode = "local"
	ModeNetwork Mode = "network"
	ModeRemote  Mode = "remote"
)

// serviceBackend es el contrato que web-access necesita de service-control.
type serviceBackend interface {
	Set(key, value string) error
	Get(key string) (string, error)
	RunRaw(args ...string) (string, error)
	Start() error
}

// Service gestiona modos de acceso y credenciales del servidor de OpenCode.
type Service struct {
	sc     serviceBackend
	lanIPs func() ([]string, error)
}

// New crea un Service sobre un service-control (inyectable para tests).
func New(sc serviceBackend) *Service {
	return &Service{sc: sc, lanIPs: localLANIPs}
}
