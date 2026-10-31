//go:build !windows && !bindings

package shell

import (
	"testing"
	"time"
)

func TestAcquireSingleInstanceLockExcluyeSegunda(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	const name = "opencode-tray-test-lock"

	ok, release := AcquireSingleInstanceLock(name)
	if !ok {
		t.Fatal("la primera instancia debería adquirir el lock")
	}

	second, releaseSecond := AcquireSingleInstanceLock(name)
	if second {
		t.Error("la segunda instancia no debería adquirir el lock")
	}
	releaseSecond()
	release()

	// Al liberar, otra instancia puede volver a entrar.
	third, releaseThird := AcquireSingleInstanceLock(name)
	if !third {
		t.Error("tras liberar, debería poder adquirirse de nuevo")
	}
	releaseThird()
}

func TestServeYRequestShow(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	const name = "opencode-tray-test-show"

	called := make(chan struct{}, 1)
	ServeShowRequests(name, func() { called <- struct{}{} })

	if !RequestShow(name) {
		t.Fatal("RequestShow debería entregar la petición")
	}

	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatal("la petición no llegó a onShow")
	}
}

func TestRequestShowSinServidor(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	if RequestShow("opencode-tray-test-nadie") {
		t.Error("no debería haber servidor escuchando")
	}
}
