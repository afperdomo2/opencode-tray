package webaccess

import (
	"errors"
	"testing"

	"opencode-tray/internal/servicecontrol"
)

type mockBackend struct {
	sets    map[string]string
	runs    [][]string
	started int
	port    string
	status  servicecontrol.Status
	info    servicecontrol.ServerInfo
	infoErr error

	snapshotErr error
	statusCalls int
	configCalls int
	getCalls    int
	infoCalls   int
}

func newMockBackend() *mockBackend {
	return &mockBackend{sets: map[string]string{}, port: "5000"}
}

func (m *mockBackend) Set(key, value string) error { m.sets[key] = value; return nil }
func (m *mockBackend) Get(key string) (string, error) {
	m.getCalls++
	if key == "port" {
		return m.port, nil
	}
	return m.sets[key], nil
}
func (m *mockBackend) Snapshot() (map[string]string, error) {
	m.configCalls++
	if m.snapshotErr != nil {
		return nil, m.snapshotErr
	}
	out := make(map[string]string, len(m.sets)+1)
	for key, value := range m.sets {
		out[key] = value
	}
	if m.port != "" {
		out["port"] = m.port
	}
	return out, nil
}
func (m *mockBackend) RunRaw(args ...string) (string, error) {
	m.runs = append(m.runs, args)
	return "tunnel-link", nil
}
func (m *mockBackend) Start() error { m.started++; return nil }
func (m *mockBackend) Status() (servicecontrol.Status, error) {
	m.statusCalls++
	return m.status, nil
}
func (m *mockBackend) Info() (servicecontrol.ServerInfo, error) {
	m.infoCalls++
	return m.info, m.infoErr
}

func TestApplyModeLocalDisablesRemote(t *testing.T) {
	b := newMockBackend()
	b.sets["remote"] = "true"
	s := New(b)
	if err := s.ApplyMode(ModeLocal); err != nil {
		t.Fatal(err)
	}
	if b.sets["hostname"] != "127.0.0.1" {
		t.Errorf("hostname = %q, want 127.0.0.1", b.sets["hostname"])
	}
	if b.sets["remote"] != "false" {
		t.Errorf("remote = %q, want false", b.sets["remote"])
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
	if b.started != 1 {
		t.Errorf("started = %d, want 1", b.started)
	}
}

func TestApplyModeRemote(t *testing.T) {
	b := newMockBackend()
	s := New(b)
	if err := s.ApplyMode(ModeRemote); err != nil {
		t.Fatal(err)
	}
	if b.sets["hostname"] != "127.0.0.1" {
		t.Errorf("hostname = %q, want 127.0.0.1", b.sets["hostname"])
	}
	if len(b.runs) != 1 {
		t.Fatalf("runs = %v", b.runs)
	}
	got := b.runs[0]
	if len(got) != 3 || got[0] != "pair" || got[1] != "--remote" || got[2] != "--yes" {
		t.Errorf("RunRaw args = %v", got)
	}
	if b.started != 1 {
		t.Errorf("started = %d, want 1", b.started)
	}
}

func TestApplyModeInvalid(t *testing.T) {
	s := New(newMockBackend())
	if err := s.ApplyMode(Mode("otro")); err == nil {
		t.Error("modo inválido debería dar error")
	}
}

func TestComputeURLsRunningUsesInfo(t *testing.T) {
	b := newMockBackend()
	b.status = servicecontrol.Status{Running: true, URL: "http://127.0.0.1:45278"}
	// En red/remoto /api/info es la fuente de las IPs LAN y del túnel.
	b.sets["hostname"] = "0.0.0.0"
	b.info = servicecontrol.ServerInfo{URLs: []string{
		"http://192.168.1.5:45278",
		"http://127.0.0.1:45278",
		"https://abc123.def456.opentunnel.xyz",
		"http://127.0.0.1:45278",
	}}
	s := New(b)

	items, err := s.ComputeURLs()
	if err != nil {
		t.Fatal(err)
	}
	want := []URLItem{
		{URL: "http://127.0.0.1:45278", Kind: KindLocal},
		{URL: "http://192.168.1.5:45278", Kind: KindLAN},
		{URL: "https://abc123.def456.opentunnel.xyz", Kind: KindRemote},
	}
	if len(items) != len(want) {
		t.Fatalf("items = %v", items)
	}
	for i := range want {
		if items[i] != want[i] {
			t.Errorf("items[%d] = %v, want %v", i, items[i], want[i])
		}
	}
}

func TestComputeURLsStoppedFallback(t *testing.T) {
	b := newMockBackend()
	b.port = ""
	b.sets["hostname"] = "127.0.0.1"
	s := New(b)

	items, err := s.ComputeURLs()
	if err != nil {
		t.Fatal(err)
	}
	want := "http://127.0.0.1:" + DefaultPort()
	if len(items) != 1 || items[0].URL != want || items[0].Kind != KindLocal {
		t.Errorf("items = %v, want %q", items, want)
	}
}

func TestStateMode(t *testing.T) {
	b := newMockBackend()
	b.status = servicecontrol.Status{Running: true}
	b.info = servicecontrol.ServerInfo{URLs: []string{"http://127.0.0.1:5000"}}
	s := New(b)

	if st, _ := s.State(); st.Mode != "local" || !st.Running {
		t.Errorf("State local = %+v", st)
	}

	b.sets["hostname"] = "0.0.0.0"
	if st, _ := s.State(); st.Mode != "network" {
		t.Errorf("State network = %+v", st)
	}

	b.sets["remote"] = "true"
	if st, _ := s.State(); st.Mode != "remote" {
		t.Errorf("State remote = %+v", st)
	}
}

func TestIsAccessURL(t *testing.T) {
	b := newMockBackend()
	b.status = servicecontrol.Status{Running: true}
	b.info = servicecontrol.ServerInfo{URLs: []string{"http://127.0.0.1:5000"}}
	s := New(b)

	ok, err := s.IsAccessURL("http://127.0.0.1:5000")
	if err != nil || !ok {
		t.Errorf("IsAccessURL válida = %v, %v", ok, err)
	}
	ok, err = s.IsAccessURL("http://evil.example.com")
	if err != nil || ok {
		t.Errorf("IsAccessURL externa = %v, %v", ok, err)
	}
	ok, err = s.IsAccessURL("file:///c:/secreto")
	if err != nil || ok {
		t.Errorf("IsAccessURL file = %v, %v", ok, err)
	}
}

func TestStateReusesOneStatusAndOneConfig(t *testing.T) {
	b := newMockBackend()
	b.status = servicecontrol.Status{Running: true, URL: "http://127.0.0.1:5000"}
	b.sets["hostname"] = "127.0.0.1"
	s := New(b)

	if _, err := s.State(); err != nil {
		t.Fatal(err)
	}
	if b.statusCalls != 1 {
		t.Errorf("statusCalls = %d, want 1", b.statusCalls)
	}
	if b.configCalls != 1 {
		t.Errorf("configCalls = %d, want 1", b.configCalls)
	}
	if b.infoCalls != 0 {
		t.Errorf("infoCalls = %d, want 0 en modo local", b.infoCalls)
	}
}

func TestStateNetworkUsesInfo(t *testing.T) {
	b := newMockBackend()
	b.status = servicecontrol.Status{Running: true, URL: "http://0.0.0.0:5000"}
	b.sets["hostname"] = "0.0.0.0"
	b.info = servicecontrol.ServerInfo{URLs: []string{
		"http://192.168.1.5:5000",
		"http://127.0.0.1:5000",
	}}
	s := New(b)

	st, err := s.State()
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode != string(ModeNetwork) {
		t.Errorf("Mode = %q", st.Mode)
	}
	if b.infoCalls != 1 {
		t.Errorf("infoCalls = %d, want 1", b.infoCalls)
	}
	if len(st.URLs) != 2 {
		t.Errorf("URLs = %v", st.URLs)
	}
}

func TestStateLocalWithoutStatusURLUsesInfo(t *testing.T) {
	b := newMockBackend()
	b.status = servicecontrol.Status{Running: true} // corriendo pero sin URL
	b.sets["hostname"] = "127.0.0.1"
	b.info = servicecontrol.ServerInfo{URLs: []string{"http://127.0.0.1:5000"}}
	s := New(b)

	st, err := s.State()
	if err != nil {
		t.Fatal(err)
	}
	if b.infoCalls != 1 {
		t.Errorf("infoCalls = %d, want 1 (sin URL de status)", b.infoCalls)
	}
	if len(st.URLs) != 1 || st.URLs[0].URL != "http://127.0.0.1:5000" {
		t.Errorf("URLs = %v", st.URLs)
	}
}

func TestStateFallsBackWhenSnapshotFails(t *testing.T) {
	b := newMockBackend()
	b.snapshotErr = errors.New("formato distinto")
	b.sets["hostname"] = "0.0.0.0"
	s := New(b)

	st, err := s.State()
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode != string(ModeNetwork) {
		t.Errorf("Mode = %q, want network", st.Mode)
	}
	if b.configCalls != 1 {
		t.Errorf("configCalls = %d, want 1", b.configCalls)
	}
	if b.getCalls != 3 {
		t.Errorf("getCalls = %d, want 3 (hostname, remote, port)", b.getCalls)
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
