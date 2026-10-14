package webaccess

import "testing"

type mockBackend struct {
	sets    map[string]string
	runs    [][]string
	started int
	port    string
}

func newMockBackend() *mockBackend {
	return &mockBackend{sets: map[string]string{}, port: "5000"}
}

func (m *mockBackend) Set(key, value string) error { m.sets[key] = value; return nil }
func (m *mockBackend) Get(key string) (string, error) {
	if key == "port" {
		return m.port, nil
	}
	return m.sets[key], nil
}
func (m *mockBackend) RunRaw(args ...string) (string, error) {
	m.runs = append(m.runs, args)
	return "tunnel-link", nil
}
func (m *mockBackend) Start() error { m.started++; return nil }

func TestApplyModeLocal(t *testing.T) {
	b := newMockBackend()
	s := New(b)
	if err := s.ApplyMode(ModeLocal); err != nil {
		t.Fatal(err)
	}
	if b.sets["hostname"] != "127.0.0.1" {
		t.Errorf("hostname = %q, want 127.0.0.1", b.sets["hostname"])
	}
	if b.started != 1 {
		t.Errorf("started = %d, want 1", b.started)
	}
}

func TestApplyModeNetwork(t *testing.T) {
	b := newMockBackend()
	s := New(b)
	if err := s.ApplyMode(ModeNetwork); err != nil {
		t.Fatal(err)
	}
	if b.sets["hostname"] != "0.0.0.0" {
		t.Errorf("hostname = %q, want 0.0.0.0", b.sets["hostname"])
	}
}

func TestApplyModeRemote(t *testing.T) {
	b := newMockBackend()
	s := New(b)
	if err := s.ApplyMode(ModeRemote); err != nil {
		t.Fatal(err)
	}
	if len(b.runs) != 1 {
		t.Fatalf("runs = %v", b.runs)
	}
	got := b.runs[0]
	if len(got) != 3 || got[0] != "pair" || got[1] != "--remote" || got[2] != "--yes" {
		t.Errorf("RunRaw args = %v", got)
	}
}

func TestApplyModeInvalid(t *testing.T) {
	s := New(newMockBackend())
	if err := s.ApplyMode(Mode("otro")); err == nil {
		t.Error("modo inválido debería dar error")
	}
}

func TestComputeURLs(t *testing.T) {
	b := newMockBackend()
	s := New(b)
	s.lanIPs = func() ([]string, error) { return []string{"192.168.1.5", "10.0.0.2"}, nil }

	urls, err := s.ComputeURLs()
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 2 || urls[0] != "http://192.168.1.5:5000" || urls[1] != "http://10.0.0.2:5000" {
		t.Errorf("urls = %v", urls)
	}
}

func TestCredentials(t *testing.T) {
	b := newMockBackend()
	b.sets["password"] = "secret"
	s := New(b)

	if s.Username() != "opencode" {
		t.Errorf("Username = %q", s.Username())
	}
	if pw, err := s.CurrentPassword(); err != nil || pw != "secret" {
		t.Errorf("CurrentPassword = %q, %v", pw, err)
	}
	if err := s.ChangePassword("nueva"); err != nil {
		t.Fatal(err)
	}
	if b.sets["password"] != "nueva" {
		t.Errorf("password = %q", b.sets["password"])
	}
	if b.started != 1 {
		t.Errorf("ChangePassword debería reiniciar, started = %d", b.started)
	}
}
