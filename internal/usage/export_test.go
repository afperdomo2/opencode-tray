package usage

import (
	"context"
	"errors"
	"testing"
)

const exportJSON = `[
  {"id":"1","integrationID":"github-copilot","label":"OAuth","active":true,"value":{"type":"oauth","refresh":"r","access":"a","expires":1}},
  {"id":"2","integrationID":"opencode-go","label":"OpenCode Go","active":true,"value":{"type":"key","key":"go-key"}}
]`

func TestParseAuthExportArray(t *testing.T) {
	key, err := parseAuthExport([]byte(exportJSON))
	if err != nil {
		t.Fatal(err)
	}
	if key.Provider != providerGo || key.Token != "go-key" || key.Source != sourceExport {
		t.Errorf("key = %+v", key)
	}
}

func TestParseAuthExportObject(t *testing.T) {
	data := `{"opencode":{"type":"api","key":"legacy-key"}}`
	key, err := parseAuthExport([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if key.Provider != providerLegacy || key.Token != "legacy-key" {
		t.Errorf("key = %+v", key)
	}
}

func TestParseAuthExportPrefersGo(t *testing.T) {
	data := `[
	  {"integrationID":"opencode","active":true,"value":{"type":"key","key":"legacy"}},
	  {"integrationID":"opencode-go","active":true,"value":{"type":"key","key":"go"}}
	]`
	key, err := parseAuthExport([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if key.Token != "go" {
		t.Errorf("token = %q, want go", key.Token)
	}
}

func TestParseAuthExportSkipsInactive(t *testing.T) {
	data := `[{"integrationID":"opencode-go","active":false,"value":{"type":"key","key":"old"}}]`
	if _, err := parseAuthExport([]byte(data)); !errors.Is(err, errNoExportedToken) {
		t.Errorf("inactivo debería ignorarse, got %v", err)
	}
}

func TestParseAuthExportIgnoresOtherProviders(t *testing.T) {
	data := `[{"integrationID":"github-copilot","active":true,"value":{"type":"oauth","access":"a"}}]`
	if _, err := parseAuthExport([]byte(data)); !errors.Is(err, errNoExportedToken) {
		t.Errorf("otro proveedor debería ignorarse, got %v", err)
	}
}

func TestParseAuthExportInvalid(t *testing.T) {
	if _, err := parseAuthExport([]byte("{nope")); err == nil || errors.Is(err, errNoExportedToken) {
		t.Errorf("JSON inválido debería dar error de parseo, got %v", err)
	}
}

type scriptedRunner struct {
	calls [][]string
	out   [][]byte
	errs  []error
}

func (r *scriptedRunner) Run(_ context.Context, args []string) ([]byte, error) {
	r.calls = append(r.calls, append([]string(nil), args...))
	i := len(r.calls) - 1
	var out []byte
	var err error
	if i < len(r.out) {
		out = r.out[i]
	}
	if i < len(r.errs) {
		err = r.errs[i]
	}
	return out, err
}

func TestReloadTokenUsesStandaloneFallback(t *testing.T) {
	r := &scriptedRunner{
		out:  [][]byte{nil, []byte(exportJSON)},
		errs: []error{errors.New("sin servicio"), nil},
	}
	s := &Service{runner: r}

	status, err := s.ReloadToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !status.Configured || status.Source != sourceExport || status.Provider != providerGo {
		t.Errorf("status = %+v", status)
	}
	if len(r.calls) != 2 {
		t.Fatalf("llamadas = %d, want 2", len(r.calls))
	}
	if got := r.calls[1]; len(got) != 3 || got[2] != "--standalone" {
		t.Errorf("segunda llamada = %v", got)
	}
}

func TestReloadTokenFallsBackWhenFastExportLacksToken(t *testing.T) {
	r := &scriptedRunner{
		out: [][]byte{
			[]byte(`[{"integrationID":"opencode-go","active":true}]`),
			[]byte(exportJSON),
		},
	}
	s := &Service{runner: r}

	status, err := s.ReloadToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.Provider != providerGo {
		t.Errorf("status = %+v", status)
	}
	if len(r.calls) != 2 {
		t.Errorf("llamadas = %d, want 2", len(r.calls))
	}
}

func TestReloadTokenNoCredentials(t *testing.T) {
	r := &scriptedRunner{out: [][]byte{nil, []byte(`[]`)}}
	s := &Service{runner: r}
	status, err := s.ReloadToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.Configured {
		t.Errorf("status = %+v", status)
	}
}

func TestReloadTokenClearsPreviousCache(t *testing.T) {
	r := &scriptedRunner{out: [][]byte{[]byte(exportJSON)}}
	s := &Service{runner: r}
	s.setCachedExport(AuthKey{Provider: providerLegacy, Token: "stale", Source: sourceExport})

	status, err := s.ReloadToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.Provider != providerGo {
		t.Errorf("provider = %q, want opencode-go", status.Provider)
	}
	key, ok := s.cachedExport()
	if !ok || key.Token != "go-key" {
		t.Errorf("cache = %+v, ok=%v", key, ok)
	}
}
