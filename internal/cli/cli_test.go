package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePrecedencia(t *testing.T) {
	home := "/home/dev"
	exists := map[string]bool{
		"/opt/custom/opencode":                        true,
		"/usr/local/bin/opencode":                     true,
		filepath.Join(home, ".opencode/bin/opencode"): true,
	}
	isExec := func(p string) bool { return exists[p] }
	lookPathFound := func(p string) (string, error) { return "/usr/bin/opencode", nil }
	lookPathMissing := func(string) (string, error) { return "", errors.New("no encontrado") }

	tests := []struct {
		name     string
		override string
		lookPath func(string) (string, error)
		want     string
	}{
		{name: "OPENCODE_BIN válido gana", override: "/opt/custom/opencode", lookPath: lookPathFound, want: "/opt/custom/opencode"},
		{name: "OPENCODE_BIN inexistente se ignora", override: "/no/existe", lookPath: lookPathFound, want: "/usr/bin/opencode"},
		{name: "PATH antes que rutas conocidas", lookPath: lookPathFound, want: "/usr/bin/opencode"},
		{name: "sin PATH usa ~/.opencode/bin", lookPath: lookPathMissing, want: filepath.Join(home, ".opencode/bin/opencode")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolve(tt.override, tt.lookPath, home, isExec); got != tt.want {
				t.Errorf("resolve() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveSinCandidatos(t *testing.T) {
	isExec := func(string) bool { return false }
	lookPath := func(string) (string, error) { return "", errors.New("no encontrado") }
	if got := resolve("", lookPath, "/home/dev", isExec); got != "opencode" {
		t.Errorf("resolve() = %q, want %q", got, "opencode")
	}
}

func TestResolveIgnoraDirectorios(t *testing.T) {
	dir := t.TempDir()
	// ~/.opencode/bin/opencode como directorio no debe aceptarse.
	if err := os.MkdirAll(filepath.Join(dir, ".opencode", "bin", "opencode"), 0o755); err != nil {
		t.Fatal(err)
	}
	lookPath := func(string) (string, error) { return "", errors.New("no encontrado") }
	if got := resolve("", lookPath, dir, isExecutable); got != "opencode" {
		t.Errorf("resolve() = %q, want %q", got, "opencode")
	}
}
