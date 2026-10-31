// Package cli localiza el ejecutable de OpenCode.
//
// La app se lanza desde el menú de aplicaciones o desde el auto-inicio, donde
// el PATH no incluye los directorios del shell del usuario (por ejemplo
// ~/.opencode/bin). Resolver el binario aquí evita que el usuario tenga que
// crear enlaces o tocar su entorno.
package cli

import (
	"os"
	"os/exec"
	"path/filepath"
)

// EnvVar permite forzar la ruta del CLI de OpenCode.
const EnvVar = "OPENCODE_BIN"

// binaryName es el nombre del ejecutable que se busca.
const binaryName = "opencode"

// Resolve devuelve la ruta del CLI de OpenCode (o "opencode" si no lo
// encuentra, para conservar los mensajes de error habituales).
func Resolve() string {
	home, _ := os.UserHomeDir()
	return resolve(os.Getenv(EnvVar), exec.LookPath, home, isExecutable)
}

// resolve implementa la búsqueda con dependencias inyectables (para tests).
func resolve(override string, lookPath func(string) (string, error), home string, isExec func(string) bool) string {
	if override != "" && isExec(override) {
		return override
	}
	if path, err := lookPath(binaryName); err == nil && path != "" {
		return path
	}

	// Rutas típicas de instalación del CLI cuando no está en el PATH.
	candidates := make([]string, 0, 5)
	if home != "" {
		candidates = append(candidates,
			filepath.Join(home, ".opencode", "bin", binaryName),
			filepath.Join(home, ".local", "bin", binaryName),
		)
	}
	candidates = append(candidates,
		filepath.Join("/usr", "local", "bin", binaryName),
		filepath.Join("/usr", "bin", binaryName),
		filepath.Join("/snap", "bin", binaryName),
	)
	for _, candidate := range candidates {
		if isExec(candidate) {
			return candidate
		}
	}
	return binaryName
}

// isExecutable comprueba que la ruta existe y tiene permiso de ejecución.
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	return info.Mode().Perm()&0o111 != 0
}
