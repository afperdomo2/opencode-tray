//go:build !windows

package shell

// AcquireSingleInstanceLock en no-Windows no restringe instancias.
func AcquireSingleInstanceLock(name string) (bool, func()) {
	return true, func() {}
}
