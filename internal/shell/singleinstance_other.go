//go:build !windows && !bindings

package shell

// AcquireSingleInstanceLock en no-Windows no restringe instancias. El tag
// `bindings` se excluye porque durante la generación de bindings de Wails
// singleinstance_bindings.go ya define esta función.
func AcquireSingleInstanceLock(name string) (bool, func()) {
	return true, func() {}
}
