//go:build bindings

package shell

// AcquireSingleInstanceLock se desactiva durante la generación de bindings de
// Wails: el helper de bindings debe poder ejecutarse aunque la app esté abierta.
// Sin esto, el mutex de instancia única hace que `wails build` no regenere
// `frontend/wailsjs` y falle en silencio.
func AcquireSingleInstanceLock(name string) (bool, func()) {
	return true, func() {}
}
