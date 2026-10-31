//go:build !windows && !linux

package shell

import "errors"

// NewSystemRunKey devuelve un runKey sin operación en sistemas sin autoarranque
// soportado (hoy solo Windows usa el registro y Linux XDG Autostart).
func NewSystemRunKey() (runKey, error) {
	return noopRunKey{}, nil
}

type noopRunKey struct{}

func (noopRunKey) Set(name, value string) error { return nil }
func (noopRunKey) Delete(name string) error     { return nil }
func (noopRunKey) Get(name string) (string, bool, error) {
	return "", false, errors.New("auto-inicio no soportado en este SO")
}
