package webaccess

import (
	"net"
	"net/url"
	"runtime"
	"sort"
	"strings"
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
func (s *Service) State() (State, error) {
	st, _ := s.sc.Status()
	items, err := s.ComputeURLs()
	if err != nil {
		return State{}, err
	}
	return State{
		Mode:    s.currentMode(),
		Running: st.Running,
		URLs:    items,
	}, nil
}

// currentMode deduce el modo activo a partir de la configuración del servicio.
// "remote" manda sobre el bind: si el túnel está activo, es el modo visible.
func (s *Service) currentMode() string {
	if remote, err := s.sc.Get("remote"); err == nil && strings.EqualFold(strings.TrimSpace(remote), "true") {
		return string(ModeRemote)
	}
	host, _ := s.sc.Get("hostname")
	switch strings.TrimSpace(host) {
	case "0.0.0.0", "::":
		return string(ModeNetwork)
	default:
		return string(ModeLocal)
	}
}

// ComputeURLs devuelve las URLs de acceso. Cuando el servicio está corriendo
// usa /api/info: es la fuente oficial e incluye la URL del túnel remoto y las
// IPs LAN reales. Si está detenido, devuelve solo la URL local configurada.
func (s *Service) ComputeURLs() ([]URLItem, error) {
	st, _ := s.sc.Status()
	if st.Running {
		if info, err := s.sc.Info(); err == nil && len(info.URLs) > 0 {
			return normalizeItems(info.URLs), nil
		}
		if st.URL != "" {
			return normalizeItems([]string{st.URL}), nil
		}
	}

	host, _ := s.sc.Get("hostname")
	host = strings.TrimSpace(host)
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	port, _ := s.sc.Get("port")
	port = strings.TrimSpace(port)
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
