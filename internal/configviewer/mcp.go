package configviewer

import (
	"fmt"
	"strings"
)

// ListMCP lista los servidores MCP con su estado, vía `opencode mcp list`.
// Salida típica: "✓ context7  connected" / "✗ foo  needs authentication".
func (s *Service) ListMCP() ([]MCPServer, error) {
	stdout, stderr, err := s.runner.Run([]string{"mcp", "list"})
	if err != nil {
		return nil, fmt.Errorf("opencode mcp list: %w (%s)", err, strings.TrimSpace(stderr))
	}

	var servers []MCPServer
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Quitar el símbolo inicial (✓, ✗, ●, ...).
		line = strings.TrimLeft(line, "✓✗●◦▪▸✓✔ ")
		if line == "" {
			continue
		}
		// El primer token es el nombre; el resto es el estado.
		fields := strings.Fields(line)
		name := fields[0]
		status := strings.Join(fields[1:], " ")
		servers = append(servers, MCPServer{Name: name, Status: status})
	}
	return servers, nil
}
