package usage

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func writeAuthFile(t *testing.T, key string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "auth.json")
	content := `{"opencode-go":{"type":"api","key":"` + key + `"}}`
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func newTestService(t *testing.T, handler http.HandlerFunc) (*Service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return &Service{client: srv.Client(), baseURL: srv.URL}, srv
}

func TestFetchParsesWindows(t *testing.T) {
	s, _ := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer k1" {
			t.Errorf("Authorization = %q", got)
		}
		w.Write([]byte(`{"usage":{
			"rolling":{"status":"ok","percent":43.0,"resetsAt":"2026-10-14T06:00:00Z"},
			"weekly":{"status":"ok","percent":21.5,"resetsAt":null},
			"monthly":{"status":"ok","percent":8.2,"resetsAt":"2026-11-01T00:00:00Z"}
		}}`))
	})
	s.SetManualToken("k1")

	u, err := s.Fetch()
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if u.Rolling.Percent != 43.0 {
		t.Errorf("Rolling.Percent = %v", u.Rolling.Percent)
	}
	if !u.Rolling.HasReset {
		t.Error("Rolling debería tener reset")
	}
	if u.Weekly.HasReset {
		t.Error("Weekly no debería tener reset (null)")
	}
	if u.Monthly.Percent != 8.2 {
		t.Errorf("Monthly.Percent = %v", u.Monthly.Percent)
	}
}

func TestFetchHandlesNullPercent(t *testing.T) {
	s, _ := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"usage":{"rolling":{"percent":null},"weekly":{"percent":50},"monthly":{}}}`))
	})
	s.SetManualToken("k1")

	u, err := s.Fetch()
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if u.Rolling.Percent != 0 || u.Rolling.HasReset {
		t.Errorf("rolling null debería quedar vacío, got %+v", u.Rolling)
	}
	if u.Weekly.Percent != 50 {
		t.Errorf("Weekly.Percent = %v", u.Weekly.Percent)
	}
}

func TestFetchUnauthorized(t *testing.T) {
	s, _ := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	s.SetManualToken("bad")
	if _, err := s.Fetch(); err == nil {
		t.Error("401 debería dar error")
	}
}

func TestFetchNotSubscribed(t *testing.T) {
	s, _ := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":{"type":"EntitlementError"}}`))
	})
	s.SetManualToken("k1")
	_, err := s.Fetch()
	if err == nil || err.Error() != "sin suscripción OpenCode Go activa" {
		t.Errorf("error = %v", err)
	}
}

func TestResolveTokenManualWins(t *testing.T) {
	s := &Service{authFile: writeAuthFile(t, "from-auth")}
	s.SetManualToken("manual")
	tok, err := s.ResolveToken()
	if err != nil {
		t.Fatal(err)
	}
	if tok != "manual" {
		t.Errorf("token = %q, want manual", tok)
	}
	s.ClearManualToken()
	tok, err = s.ResolveToken()
	if err != nil {
		t.Fatal(err)
	}
	if tok != "from-auth" {
		t.Errorf("token = %q, want from-auth", tok)
	}
}

func TestResolveTokenFromAuthFile(t *testing.T) {
	s := &Service{authFile: writeAuthFile(t, "from-auth")}
	tok, err := s.ResolveToken()
	if err != nil {
		t.Fatal(err)
	}
	if tok != "from-auth" {
		t.Errorf("token = %q", tok)
	}
}

func TestResolveTokenMissing(t *testing.T) {
	s := &Service{authFile: filepath.Join(t.TempDir(), "nope.json")}
	if _, err := s.ResolveToken(); err == nil {
		t.Error("sin auth.json debería dar error")
	}
}

func TestSeverity(t *testing.T) {
	s := &Service{}
	cases := []struct {
		name string
		u    Usage
		want Level
	}{
		{"verde", Usage{Monthly: Window{Percent: 10}}, LevelOK},
		{"amarillo >=90", Usage{Weekly: Window{Percent: 90}}, LevelWarn},
		{"rojo >=98", Usage{Rolling: Window{Percent: 98}}, LevelCritical},
		{"usa la mayor ventana", Usage{Rolling: Window{Percent: 5}, Monthly: Window{Percent: 99}}, LevelCritical},
	}
	for _, c := range cases {
		if got := s.Severity(c.u); got != c.want {
			t.Errorf("%s: Severity() = %v, want %v", c.name, got, c.want)
		}
	}
}
