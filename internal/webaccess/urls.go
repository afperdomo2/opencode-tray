package webaccess

import (
	"net"
	"net/url"
	"runtime"
	"sort"
	"strings"

	"opencode-tray/internal/servicecontrol"
)

// DefaultPort devuelve el puerto que usa OpenCode cuando no hay uno fijado.
// En Windows el rango efímero empieza en 49152, por eso el puerto es 45278.
func DefaultPort() string {
	if runtime.GOOS == "windows" {
		return "45278"
	}
	return "49374"
}

// State devuelve el modo activo, el estado del servicio y las URLs de acceso.
// Usa una sola lectura de `service status` y una sola de configuración.
func (s *Service) State() (State, error) {
	st, _ := s.sc.Status()
	return s.StateWithStatus(st)
}

// StateWithStatus construye el estado reutilizando un `service status` ya
// leído (App lo cachea para no repetir el comando en llamadas paralelas).
func (s *Service) StateWithStatus(st servicecontrol.Status) (State, error) {
	cfg, _ := s.config()
	items, err := s.computeURLs(st, cfg)
	if err != nil {
		return State{}, err
	}
	return State{
		Mode:    modeFromConfig(cfg),
		Running: st.Running,
		URLs:    items,
	}, nil
}

// config lee la configuración del servicio en una sola invocación. Si el
// snapshot falla (formato distinto del CLI), cae a lectura clave por clave.
func (s *Service) config() (map[string]string, error) {
	if cfg, err := s.sc.Snapshot(); err == nil {
		return cfg, nil
	}

	out := make(map[string]string, 3)
	for _, key := range []string{"hostname", "remote", "port"} {
		if value, err := s.sc.Get(key); err == nil {
			out[key] = value
		}
	}
	return out, nil
}

// modeFromConfig deduce el modo activo. "remote" manda sobre el bind: si el
// túnel está activo, es el modo visible. Las claves ausentes son defaults.
func modeFromConfig(cfg map[string]string) string {
	if strings.EqualFold(strings.TrimSpace(cfg["remote"]), "true") {
		return string(ModeRemote)
	}
	switch strings.TrimSpace(cfg["hostname"]) {
	case "0.0.0.0", "::":
		return string(ModeNetwork)
	default:
		return string(ModeLocal)
	}
}

// ComputeURLs devuelve las URLs de acceso con una sola lectura de estado.
func (s *Service) ComputeURLs() ([]URLItem, error) {
	st, _ := s.sc.Status()
	cfg, _ := s.config()
	return s.computeURLs(st, cfg)
}

// computeURLs calcula las URLs reutilizando estado y configuración ya leídos.
// /api/info solo se consulta en modo red/remoto: es la única fuente de las
// IPs LAN reales y del enlace del túnel. En local basta la URL de status.
func (s *Service) computeURLs(st servicecontrol.Status, cfg map[string]string) ([]URLItem, error) {
	mode := modeFromConfig(cfg)
	if st.Running {
		// Local con URL real: no hace falta /api/info.
		if mode == string(ModeLocal) && st.URL != "" {
			return normalizeItems([]string{st.URL}), nil
		}
		if info, err := s.sc.Info(); err == nil && len(info.URLs) > 0 {
			return normalizeItems(info.URLs), nil
		}
		if st.URL != "" {
			return normalizeItems([]string{st.URL}), nil
		}
	}

	host := strings.TrimSpace(cfg["hostname"])
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	port := strings.TrimSpace(cfg["port"])
	if port == "" {
		port = DefaultPort()
	}
	return []URLItem{{
		URL:  "http://" + net.JoinHostPort(host, port),
		Kind: KindLocal,
	}}, nil
}

// IsAccessURL comprueba que la URL pertenezca al conjunto calculado por la app.
// Evita abrir URLs arbitrarias desde el frontend.
func (s *Service) IsAccessURL(raw string) (bool, error) {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return false, nil
	}
	items, err := s.ComputeURLs()
	if err != nil {
		return false, err
	}
	for _, item := range items {
		if item.URL == raw {
			return true, nil
		}
	}
	return false, nil
}

// normalizeItems deduplica, clasifica y ordena local → LAN → remoto.
func normalizeItems(raw []string) []URLItem {
	seen := make(map[string]bool, len(raw))
	items := make([]URLItem, 0, len(raw))
	for _, u := range raw {
		u = strings.TrimSpace(u)
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		items = append(items, URLItem{URL: u, Kind: classify(u)})
	}
	sort.SliceStable(items, func(i, j int) bool {
		return kindRank(items[i].Kind) < kindRank(items[j].Kind)
	})
	return items
}

func classify(raw string) Kind {
	u, err := url.Parse(raw)
	if err != nil {
		return KindLAN
	}
	host := strings.ToLower(u.Hostname())
	switch {
	case host == "localhost", host == "127.0.0.1", host == "::1":
		return KindLocal
	case strings.HasSuffix(host, ".opentunnel.xyz"), u.Scheme == "https":
		return KindRemote
	default:
		return KindLAN
	}
}

func kindRank(k Kind) int {
	switch k {
	case KindLocal:
		return 0
	case KindLAN:
		return 1
	case KindRemote:
		return 2
	default:
		return 3
	}
}
