//go:build linux

package shell

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// desktopFileName es la entrada .desktop que deja el autoarranque XDG.
const desktopFileName = "opencode-tray.desktop"

// NewSystemRunKey devuelve un runKey respaldado por ~/.config/autostart
// (XDG Autostart), equivalente a HKCU\...\Run en Windows.
func NewSystemRunKey() (runKey, error) {
	dir, err := autostartDir()
	if err != nil {
		return nil, err
	}
	return newXDGRunKey(dir), nil
}

// autostartDir devuelve el directorio XDG de autoarranque del usuario.
func autostartDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "autostart"), nil
}

// xdgRunKey implementa runKey escribiendo una entrada .desktop. El valor que
// recibe Set ya viene entre comillas desde AutoStart (por si la ruta lleva
// espacios), y el estándar Desktop Entry acepta argumentos entrecomillados.
type xdgRunKey struct{ dir string }

func newXDGRunKey(dir string) xdgRunKey { return xdgRunKey{dir: dir} }

func (x xdgRunKey) path() string { return filepath.Join(x.dir, desktopFileName) }

func (x xdgRunKey) Set(name, value string) error {
	cmd := strings.TrimSpace(value)
	if cmd == "" {
		return errors.New("comando de autoarranque vacío")
	}
	// Una directiva por línea: se neutralizan los saltos de línea del valor.
	cmd = strings.NewReplacer("\n", " ", "\r", " ").Replace(cmd)
	if err := os.MkdirAll(x.dir, 0o755); err != nil {
		return err
	}
	entry := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=OpenCode Tray
Comment=Control del servicio de OpenCode
Exec=%s
Terminal=false
X-GNOME-Autostart-enabled=true
`, cmd)
	return os.WriteFile(x.path(), []byte(entry), 0o644)
}

func (x xdgRunKey) Delete(name string) error {
	err := os.Remove(x.path())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (x xdgRunKey) Get(name string) (string, bool, error) {
	data, err := os.ReadFile(x.path())
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "Exec="); ok {
			return strings.TrimSpace(v), true, nil
		}
	}
	return "", false, nil
}
