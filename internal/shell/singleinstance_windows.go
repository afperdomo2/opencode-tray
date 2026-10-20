//go:build windows && !bindings

package shell

import (
	"errors"

	"golang.org/x/sys/windows"
)

// AcquireSingleInstanceLock intenta adquirir un mutex con nombre.
// Devuelve true si esta es la única instancia, y una función para liberarlo.
func AcquireSingleInstanceLock(name string) (bool, func()) {
	n, err := windows.UTF16PtrFromString(name)
	if err != nil {
		// Nombre inválido: no bloqueamos el arranque.
		return true, func() {}
	}
	h, err := windows.CreateMutex(nil, false, n)
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		if h != 0 {
			windows.CloseHandle(h)
		}
		return false, func() {}
	}
	if err != nil {
		return true, func() {}
	}
	return true, func() { windows.CloseHandle(h) }
}
