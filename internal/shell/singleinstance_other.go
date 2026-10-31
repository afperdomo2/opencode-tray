//go:build !windows && !bindings

package shell

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// requestShow es el mensaje que envía una segunda ejecución al proceso vivo.
const requestShow = "show"

// runtimeFile devuelve una ruta dentro de $XDG_RUNTIME_DIR (o el temporal) para
// el archivo name+ext.
func runtimeFile(name, ext string) string {
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		dir = os.TempDir()
	}
	return filepath.Join(dir, name+ext)
}

// AcquireSingleInstanceLock bloquea un archivo con flock: una segunda instancia
// no consigue el lock y sale. Al terminar el proceso el lock se libera solo, así
// que un cierre abrupto no deja la app bloqueada.
func AcquireSingleInstanceLock(name string) (bool, func()) {
	f, err := os.OpenFile(runtimeFile(name, ".lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		// Sin lock disponible no se impide el arranque.
		return true, func() {}
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		return false, func() {}
	}
	return true, func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}
}

// RequestShow pide a la instancia que ya está corriendo que muestre el panel y
// devuelve true si alguien lo recibió. Reintenta durante ~0,5 s porque una
// segunda ejecución puede coincidir con el arranque de la primera.
func RequestShow(name string) bool {
	path := runtimeFile(name, ".sock")
	for i := 0; i < 20; i++ {
		conn, err := net.DialTimeout("unix", path, 200*time.Millisecond)
		if err == nil {
			_, _ = conn.Write([]byte(requestShow + "\n"))
			_ = conn.Close()
			return true
		}
		time.Sleep(25 * time.Millisecond)
	}
	return false
}

// ServeShowRequests escucha peticiones de "mostrar panel" de otras instancias.
// Si no se puede escuchar, la app sigue funcionando (solo se pierde ese atajo).
func ServeShowRequests(name string, onShow func()) {
	path := runtimeFile(name, ".sock")
	// Un cierre abrupto puede dejar el socket; sin borrarlo no se puede escuchar.
	_ = os.Remove(path)
	listener, err := net.Listen("unix", path)
	if err != nil {
		return
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			buf := make([]byte, 32)
			n, _ := conn.Read(buf)
			_ = conn.Close()
			if strings.TrimSpace(string(buf[:n])) == requestShow && onShow != nil {
				onShow()
			}
		}
	}()
}
