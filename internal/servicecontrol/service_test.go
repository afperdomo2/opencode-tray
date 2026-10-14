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
