//go:build bindings

package shell

// AcquireSingleInstanceLock se desactiva durante la generación de bindings de
// Wails: el helper de bindings debe poder ejecutarse aunque la app esté abierta.
// Sin esto, el mutex de instancia única hace que `wails build` no regenere
// `frontend/wailsjs` y falle en silencio.
func AcquireSingleInstanceLock(name string) (bool, func()) {
	return true, func() {}
}

// RequestShow no hace nada durante la generación de bindings.
func RequestShow(name string) bool { return false }

// ServeShowRequests no hace nada durante la generación de bindings.
func ServeShowRequests(name string, onShow func()) {}
