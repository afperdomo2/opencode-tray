package webaccess

import (
	"fmt"
	"net"
)

// DefaultPort es el puerto por defecto de OpenCode.
const DefaultPort = "49374"

// ComputeURLs devuelve las URLs de acceso (IPs LAN + puerto del servicio).
func (s *Service) ComputeURLs() ([]string, error) {
	port, err := s.sc.Get("port")
	if err != nil || port == "" {
		port = DefaultPort
	}
	ips, err := s.lanIPs()
	if err != nil {
		return nil, err
	}
	var urls []string
	for _, ip := range ips {
		urls = append(urls, fmt.Sprintf("http://%s:%s", ip, port))
	}
	return urls, nil
}

// localLANIPs devuelve las IPs IPv4 privadas (no loopback) de la máquina.
func localLANIPs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	var ips []string
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ips = append(ips, ipnet.IP.String())
		}
	}
	return ips, nil
}
