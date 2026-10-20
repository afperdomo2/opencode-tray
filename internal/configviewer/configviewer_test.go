package configviewer

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type mockRunner struct{ out string }

func (m *mockRunner) Run(args []string) (string, string, error) { return m.out, "", nil }

func TestListMCPParsesOutput(t *testing.T) {
	r := &mockRunner{out: "✓ context7  connected\n✗ foo  needs authentication\n\n"}
	s := NewWithRunner(r, t.TempDir())

	servers, err := s.ListMCP()
	if err != nil {
		t.Fatalf("ListMCP() error: %v", err)
	}
	if len(servers) != 2 {
		t.Fatalf("servers = %v, want 2", servers)
	}
	if servers[0].Name != "context7" || servers[0].Status != "connected" {
		t.Errorf("servers[0] = %+v", servers[0])
	}
	if servers[1].Name != "foo" || servers[1].Status != "needs authentication" {
		t.Errorf("servers[1] = %+v", servers[1])
	}
}

func TestListMCPIgnoresStatusMarkers(t *testing.T) {
	r := &mockRunner{out: "○ github  disabled\n● local  connected\n✗ foo  needs authentication\n"}
	s := NewWithRunner(r, t.TempDir())

	servers, err := s.ListMCP()
	if err != nil {
		t.Fatalf("ListMCP() error: %v", err)
	}
	want := []MCPServer{
		{Name: "github", Status: "disabled"},
		{Name: "local", Status: "connected"},
		{Name: "foo", Status: "needs authentication"},
	}
	if !reflect.DeepEqual(servers, want) {
		t.Errorf("servers = %+v, want %+v", servers, want)
	}
}

func TestConfigJSONMasksSecrets(t *testing.T) {
	dir := t.TempDir()
	cfg := `{
  "mcp": { "servers": { "a": { "type": "remote", "headers": { "Authorization": "Bearer SECRET123" } } } },
  "providers": { "p": { "settings": { "apiKey": "sk-REAL" } } }
}`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	s := NewWithRunner(&mockRunner{}, dir)

	out, err := s.ConfigJSON()
	if err != nil {
		t.Fatalf("ConfigJSON() error: %v", err)
	}
	if strings.Contains(out, "SECRET123") || strings.Contains(out, "sk-REAL") {
		t.Errorf("no debería contener secretos:\n%s", out)
	}
	if !strings.Contains(out, "••••") {
		t.Errorf("debería contener el marcador enmascarado:\n%s", out)
	}
	if !strings.Contains(out, `"a"`) {
		t.Errorf("debería conservar la estructura:\n%s", out)
	}
}

func TestMaskSecrets(t *testing.T) {
	in := `{"apiKey":"sk-1","token":"tk","normal":"ok","key":"k"}`
	out := MaskSecrets(in)
	if strings.Contains(out, "sk-1") || strings.Contains(out, `"tk"`) || strings.Contains(out, `"k"`) {
		t.Errorf("no enmascaró: %s", out)
	}
	if !strings.Contains(out, `"normal":"ok"`) {
		t.Errorf("debió conservar normal: %s", out)
	}
}

func TestOpenInEditorUsesCodeWhenAvailable(t *testing.T) {
	s := NewWithRunner(&mockRunner{}, t.TempDir())
	var ran []string
	s.open = func(args ...string) error { ran = args; return nil }
	s.lookPath = func(string) (string, error) { return "C:/bin/code", nil }

	if err := s.OpenInEditor("/x/file.json"); err != nil {
		t.Fatal(err)
	}
	if len(ran) != 2 || ran[0] != "C:/bin/code" || ran[1] != "/x/file.json" {
		t.Errorf("ran = %v", ran)
	}
}

func TestOpenInEditorFallsBackToDefault(t *testing.T) {
	s := NewWithRunner(&mockRunner{}, t.TempDir())
	var ran []string
	s.open = func(args ...string) error { ran = args; return nil }
	s.lookPath = func(string) (string, error) { return "", errors.New("no code") }

	if err := s.OpenInEditor("/x/file.json"); err != nil {
		t.Fatal(err)
	}
	want := defaultOpenArgs("/x/file.json")
	if strings.Join(ran, " ") != strings.Join(want, " ") {
		t.Errorf("ran = %v, want %v", ran, want)
	}
}

func TestOpenConfigDir(t *testing.T) {
	dir := t.TempDir()
	s := NewWithRunner(&mockRunner{}, dir)
	var ran []string
	s.open = func(args ...string) error { ran = args; return nil }

	if err := s.OpenConfigDir(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(ran, " "), dir) {
		t.Errorf("ran = %v, debería contener %s", ran, dir)
	}
}
