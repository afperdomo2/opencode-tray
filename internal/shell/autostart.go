package shell

import "fmt"

// runKey abstracta el almacén de autoarranque del SO (HKCU\...\Run en Windows,
// ~/.config/autostart en Linux), para poder testear la lógica sin tocar el
// sistema real.
type runKey interface {
	Set(name, value string) error
	Delete(name string) error
	Get(name string) (string, bool, error)
}

// AutoStart gestiona el arranque automático con el sistema operativo.
type AutoStart struct {
	key     runKey
	exePath func() (string, error)
	appName string
}

// NewAutoStart crea un gestor sobre la clave dada.
func NewAutoStart(key runKey, exePath func() (string, error)) *AutoStart {
	return &AutoStart{key: key, exePath: exePath, appName: "OpenCodeTray"}
}

// SetEnabled habilita o deshabilita el arranque automático.
func (a *AutoStart) SetEnabled(enabled bool) error {
	if enabled {
		p, err := a.exePath()
		if err != nil {
			return fmt.Errorf("obtener ruta del ejecutable: %w", err)
		}
		return a.key.Set(a.appName, fmt.Sprintf(`"%s"`, p))
	}
	return a.key.Delete(a.appName)
}

// Enabled reporta si el arranque automático está activo.
func (a *AutoStart) Enabled() (bool, error) {
	_, ok, err := a.key.Get(a.appName)
	return ok, err
}
