//go:build windows

package shell

import (
	"errors"

	"golang.org/x/sys/windows/registry"
)

// NewSystemRunKey abre la clave HKCU\...\Run (real) como runKey.
func NewSystemRunKey() (runKey, error) {
	k, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Run`,
		registry.QUERY_VALUE|registry.SET_VALUE,
	)
	if err != nil {
		return nil, err
	}
	return registryRunKey{k}, nil
}

type registryRunKey struct{ key registry.Key }

func (r registryRunKey) Set(name, value string) error {
	return r.key.SetStringValue(name, value)
}

func (r registryRunKey) Delete(name string) error {
	err := r.key.DeleteValue(name)
	if errors.Is(err, registry.ErrNotExist) {
		return nil
	}
	return err
}

func (r registryRunKey) Get(name string) (string, bool, error) {
	v, _, err := r.key.GetStringValue(name)
	if errors.Is(err, registry.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return v, true, nil
}
