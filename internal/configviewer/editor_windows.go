//go:build windows

package configviewer

// defaultOpenArgs abre una ruta con el programa por defecto de Windows.
func defaultOpenArgs(path string) []string {
	return []string{"cmd", "/c", "start", "", path}
}
