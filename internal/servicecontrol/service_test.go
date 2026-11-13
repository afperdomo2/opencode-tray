package servicecontrol

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// mockRunner simula el CLI de OpenCode: mapea "comando con args" a su respuesta.
type mockRunner struct {
	responses map[string]mockResult
	calls     [][]string
}

type mockResult struct {
	stdout string
	stderr string
	err    error
}

func (m *mockRunner) Run(args []string) (string, string, error) {
	m.calls = append(m.calls, args)
	key := strings.Join(args, " ")
	r, ok := m.responses[key]
	if !ok {
		return "", "", fmt.Errorf("comando no simulado: %s", key)
	}
	return r.stdout, r.stderr, r.err
}

func TestStatusRunning(t *testing.T) {
	r := &mockRunner{responses: map[string]mockResult{
		"service status": {stdout: "http://127.0.0.1:49374\n"},
	}}
	s := NewWithRunner(r, nil)

	st, err := s.Status()
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}
	if !st.Running {
		t.Error("Running = false, want true")
	}
	if st.URL != "http://127.0.0.1:49374" {
		t.Errorf("URL = %q", st.URL)
	}
}

func TestStatusStopped(t *testing.T) {
	r := &mockRunner{responses: map[string]mockResult{
		"service status": {stderr: "not running", err: errors.New("exit 1")},
	}}
	s := NewWithRunner(r, nil)

	st, err := s.Status()
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}
	if st.Running {
		t.Error("Running = true, want false")
	}
}

func TestStatusStoppedWithExitZero(t *testing.T) {
	r := &mockRunner{responses: map[string]mockResult{
		"service status": {stdout: "stopped\n"},
	}}
	s := NewWithRunner(r, nil)

	st, err := s.Status()
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}
	if st.Running {
		t.Error("Running = true, want false")
	}
	if st.URL != "" {
		t.Errorf("URL = %q, want empty", st.URL)
	}
}

func TestStatusRunningWithoutURL(t *testing.T) {
	r := &mockRunner{responses: map[string]mockResult{
		"service status": {stdout: "running\n"},
	}}
	s := NewWithRunner(r, nil)

	st, err := s.Status()
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}
	if !st.Running {
		t.Error("Running = false, want true")
	}
	if st.URL != "" {
		t.Errorf("URL = %q, want empty", st.URL)
	}
}

func TestInfoParsesPIDVersionURLs(t *testing.T) {
	r := &mockRunner{responses: map[string]mockResult{
		"api get /api/info": {stdout: `{"version":"2.0.29","pid":3692,"urls":["http://127.0.0.1:49374"]}`},
	}}
	s := NewWithRunner(r, nil)

	info, err := s.Info()
	if err != nil {
		t.Fatalf("Info() error: %v", err)
	}
	if info.PID != 3692 {
		t.Errorf("PID = %d, want 3692", info.PID)
	}
	if info.Version != "2.0.29" {
		t.Errorf("Version = %q", info.Version)
	}
	if len(info.URLs) != 1 {
		t.Errorf("URLs = %v", info.URLs)
	}
}

func TestForceStopKillsWhenStillRunning(t *testing.T) {
	r := &mockRunner{responses: map[string]mockResult{
		"service stop":      {stderr: "el bug: no detiene", err: errors.New("exit 1")},
		"service status":    {stdout: "http://127.0.0.1:49374\n"}, // sigue corriendo
		"api get /api/info": {stdout: `{"pid":3692}`},
	}}
	var killed []int
	s := NewWithRunner(r, func(pid int) error {
		killed = append(killed, pid)
		return nil
	})

	if err := s.ForceStop(); err != nil {
		t.Fatalf("ForceStop() error: %v", err)
	}
	if len(killed) != 1 || killed[0] != 3692 {
		t.Errorf("killed = %v, want [3692]", killed)
	}
}

func TestForceStopDoesNotKillWhenAlreadyStopped(t *testing.T) {
	r := &mockRunner{responses: map[string]mockResult{
		"service stop":   {}, // stop normal funciona
		"service status": {stderr: "not running", err: errors.New("exit 1")},
	}}
	var killed []int
	s := NewWithRunner(r, func(pid int) error {
		killed = append(killed, pid)
		return nil
	})

	if err := s.ForceStop(); err != nil {
		t.Fatalf("ForceStop() error: %v", err)
	}
	if len(killed) != 0 {
		t.Errorf("no debería matar, killed = %v", killed)
	}
}

func TestConfigGetSetUnset(t *testing.T) {
	r := &mockRunner{responses: map[string]mockResult{
		"service get hostname":           {stdout: "0.0.0.0\n"},
		"service set hostname 127.0.0.1": {},
		"service unset hostname":         {},
	}}
	s := NewWithRunner(r, nil)

	v, err := s.Get("hostname")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if v != "0.0.0.0" {
		t.Errorf("Get(hostname) = %q, want 0.0.0.0", v)
	}
	if err := s.Set("hostname", "127.0.0.1"); err != nil {
		t.Errorf("Set() error: %v", err)
	}
	if err := s.Unset("hostname"); err != nil {
		t.Errorf("Unset() error: %v", err)
	}
}

func TestSnapshotParsesConfigInOneCall(t *testing.T) {
	r := &mockRunner{responses: map[string]mockResult{
		"service get": {stdout: `{"remoteAccepted":true,"hostname":"0.0.0.0","port":45278,"remote":true}`},
	}}
	s := NewWithRunner(r, nil)

	cfg, err := s.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() error: %v", err)
	}
	if cfg["hostname"] != "0.0.0.0" {
		t.Errorf("hostname = %q", cfg["hostname"])
	}
	if cfg["port"] != "45278" {
		t.Errorf("port = %q", cfg["port"])
	}
	if cfg["remote"] != "true" {
		t.Errorf("remote = %q", cfg["remote"])
	}
	if len(r.calls) != 1 || strings.Join(r.calls[0], " ") != "service get" {
		t.Errorf("calls = %v, want [[service get]]", r.calls)
	}
}

func TestSnapshotInvalidJSON(t *testing.T) {
	r := &mockRunner{responses: map[string]mockResult{
		"service get": {stdout: "no es json"},
	}}
	s := NewWithRunner(r, nil)

	if _, err := s.Snapshot(); err == nil {
		t.Error("Snapshot() debería fallar con JSON inválido")
	}
}

func TestSnapshotParsesRemoteObject(t *testing.T) {
	// OpenCode 2.x devuelve `remote` como objeto y redacta la ruta.
	r := &mockRunner{responses: map[string]mockResult{
		"service get": {stdout: `{"remote":{},"remoteAccepted":true,"hostname":"127.0.0.1"}`},
	}}
	s := NewWithRunner(r, nil)

	cfg, err := s.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() error: %v", err)
	}
	if cfg["remote"] != "true" {
		t.Errorf("remote = %q, want true", cfg["remote"])
	}
	if cfg["hostname"] != "127.0.0.1" {
		t.Errorf("hostname = %q", cfg["hostname"])
	}
}

func TestSnapshotRemoteObjectKeepsRouteOut(t *testing.T) {
	// La ruta es secreta: no debe acabar en el mapa del snapshot.
	r := &mockRunner{responses: map[string]mockResult{
		"service get": {stdout: `{"remote":{"route":"abc123"},"hostname":"127.0.0.1"}`},
	}}
	s := NewWithRunner(r, nil)

	cfg, err := s.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() error: %v", err)
	}
	if cfg["remote"] != "true" {
		t.Errorf("remote = %q, want true", cfg["remote"])
	}
	for key, value := range cfg {
		if strings.Contains(value, "abc123") {
			t.Errorf("%s contiene la ruta secreta: %q", key, value)
		}
	}
}

func TestSetRemotePassesArgs(t *testing.T) {
	r := &mockRunner{responses: map[string]mockResult{
		"service set remote true":  {},
		"service set remote false": {},
	}}
	s := NewWithRunner(r, nil)

	if err := s.SetRemote("true"); err != nil {
		t.Fatalf("SetRemote(true) error: %v", err)
	}
	if err := s.SetRemote("false"); err != nil {
		t.Fatalf("SetRemote(false) error: %v", err)
	}
	if len(r.calls) != 2 {
		t.Fatalf("calls = %v", r.calls)
	}
	if strings.Join(r.calls[0], " ") != "service set remote true" {
		t.Errorf("calls[0] = %v", r.calls[0])
	}
	if strings.Join(r.calls[1], " ") != "service set remote false" {
		t.Errorf("calls[1] = %v", r.calls[1])
	}
}

func TestSetRemoteRejectsInvalidValue(t *testing.T) {
	s := NewWithRunner(&mockRunner{}, nil)
	if err := s.SetRemote("yes"); err == nil {
		t.Error("SetRemote(yes) debería fallar")
	}
}

func TestStartPassesArgs(t *testing.T) {
	r := &mockRunner{responses: map[string]mockResult{
		"service start": {},
	}}
	s := NewWithRunner(r, nil)
	if err := s.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	if len(r.calls) != 1 || strings.Join(r.calls[0], " ") != "service start" {
		t.Errorf("calls = %v, want [[service start]]", r.calls)
	}
}
