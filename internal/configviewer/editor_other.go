//go:build !windows

package configviewer

// defaultOpenArgs abre una ruta con el programa por defecto (Unix).
func defaultOpenArgs(path string) []string {
	return []string{"xdg-open", path}
}
