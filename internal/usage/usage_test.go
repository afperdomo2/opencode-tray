package usage

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func writeAuthFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func authJSON(provider, key string) string {
	return `{"` + provider + `":{"type":"api","key":"` + key + `"}}`
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
	if _, err := s.Fetch(); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("401 debería dar ErrInvalidToken, got %v", err)
	}
}

func TestFetchNotSubscribed(t *testing.T) {
	s, _ := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":{"type":"EntitlementError"}}`))
	})
	s.SetManualToken("k1")
	if _, err := s.Fetch(); !errors.Is(err, ErrNoEntitlement) {
		t.Errorf("403 entitlement debería dar ErrNoEntitlement, got %v", err)
	}
}

func TestFetchFallsBackToNextCandidate(t *testing.T) {
	var seen []string
	s, _ := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		tok := r.Header.Get("Authorization")
		seen = append(seen, tok)
		if tok == "Bearer legacy" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Write([]byte(`{"usage":{"rolling":{"percent":10},"weekly":{"percent":20},"monthly":{"percent":30}}}`))
	})
	s.authFiles = []string{writeAuthFile(t, authJSON(providerLegacy, "legacy"))}
	s.SetManualToken("manual")

	u, err := s.Fetch()
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if u.Monthly.Percent != 30 {
		t.Errorf("Monthly.Percent = %v", u.Monthly.Percent)
	}
	if len(seen) != 2 || seen[0] != "Bearer legacy" || seen[1] != "Bearer manual" {
		t.Errorf("orden de candidatos = %v", seen)
	}
}

func TestResolveTokenPrefersAuthOverManual(t *testing.T) {
	s := &Service{authFiles: []string{writeAuthFile(t, authJSON(providerGo, "from-auth"))}}
	s.SetManualToken("manual")
	tok, err := s.ResolveToken()
	if err != nil {
		t.Fatal(err)
	}
	if tok != "from-auth" {
		t.Errorf("token = %q, want from-auth", tok)
	}
	s.authFiles = nil
	tok, err = s.ResolveToken()
	if err != nil {
		t.Fatal(err)
	}
	if tok != "manual" {
		t.Errorf("token = %q, want manual", tok)
	}
}

func TestResolveTokenAcceptsLegacyProvider(t *testing.T) {
	for _, provider := range []string{providerGo, providerLegacy} {
		t.Run(provider, func(t *testing.T) {
			s := &Service{authFiles: []string{writeAuthFile(t, authJSON(provider, "tok"))}}
			tok, err := s.ResolveToken()
			if err != nil {
				t.Fatal(err)
			}
			if tok != "tok" {
				t.Errorf("token = %q", tok)
			}
			status := s.TokenStatus()
			if !status.Configured || status.Source != sourceJSON || status.Provider != provider {
				t.Errorf("status = %+v", status)
			}
		})
	}
}

func TestResolveTokenMissing(t *testing.T) {
	s := &Service{authFiles: []string{filepath.Join(t.TempDir(), "nope.json")}}
	if _, err := s.ResolveToken(); !errors.Is(err, ErrNoToken) {
		t.Errorf("sin token debería dar ErrNoToken, got %v", err)
	}
	status := s.TokenStatus()
	if status.Configured || status.Source != sourceNone {
		t.Errorf("status = %+v", status)
	}
}

func TestTokenStatusManualAndExport(t *testing.T) {
	s := &Service{}
	s.SetManualToken("manual")
	status := s.TokenStatus()
	if !status.Configured || status.Source != sourceManual {
		t.Errorf("status manual = %+v", status)
	}

	s.setCachedExport(AuthKey{Provider: providerGo, Token: "exported", Source: sourceExport})
	status = s.TokenStatus()
	if !status.Configured || status.Source != sourceExport || status.Provider != providerGo {
		t.Errorf("status export = %+v", status)
	}
}

func TestValidateToken(t *testing.T) {
	s, _ := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "Bearer good" {
			w.Write([]byte(`{"usage":{"rolling":{"percent":1},"weekly":{"percent":2},"monthly":{"percent":3}}}`))
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	})
	if err := s.ValidateToken("good"); err != nil {
		t.Errorf("token válido: %v", err)
	}
	if err := s.ValidateToken("bad"); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("token inválido = %v", err)
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
