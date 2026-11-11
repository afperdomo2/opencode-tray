package main

import (
	"runtime"
	"strings"
	"testing"
)

func TestProjectVersion(t *testing.T) {
	tests := []struct {
		name   string
		config string
		want   string
	}{
		{"válido", `{"Info":{"productVersion":"1.4.0"}}`, "1.4.0"},
		{"con espacios", `{"Info":{"productVersion":" 1.4.0 "}}`, "1.4.0"},
		{"vacío", `{"Info":{"productVersion":""}}`, fallbackVersion},
		{"sin Info", `{}`, fallbackVersion},
		{"json inválido", `no-json`, fallbackVersion},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := projectVersion([]byte(tt.config)); got != tt.want {
				t.Errorf("projectVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEmbeddedConfigHasVersion(t *testing.T) {
	got := projectVersion(wailsConfig)
	if got == fallbackVersion {
		t.Fatalf("wails.json embebido no tiene Info.productVersion válida")
	}
	if !strings.HasPrefix(got, "1.") {
		t.Errorf("versión inesperada: %q", got)
	}
}

func TestVersionInfoBuildFields(t *testing.T) {
	got := versionInfo([]byte(`{"Info":{"productVersion":"1.4.0"}}`), "a1b2c3d4e5f6", "2026-11-11T15:04:05Z", "true")

	if got.Version != "1.4.0" {
		t.Errorf("Version = %q", got.Version)
	}
	if got.Commit != "a1b2c3d" {
		t.Errorf("Commit = %q", got.Commit)
	}
	if got.Date != "2026-11-11T15:04:05Z" {
		t.Errorf("Date = %q", got.Date)
	}
	if !got.Dirty {
		t.Error("Dirty = false, want true")
	}
	wantPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if got.Platform != wantPlatform {
		t.Errorf("Platform = %q, want %q", got.Platform, wantPlatform)
	}
}

func TestVersionInfoWithoutBuildFields(t *testing.T) {
	got := versionInfo([]byte(`{"Info":{"productVersion":"1.4.0"}}`), "", "", "")

	if got.Commit != "" || got.Date != "" {
		t.Errorf("commit/date = %q/%q, want vacíos", got.Commit, got.Date)
	}
	if got.Dirty {
		t.Error("Dirty = true, want false")
	}
}

func TestShortCommit(t *testing.T) {
	tests := map[string]string{
		"":                "",
		"abc":             "abc",
		"a1b2c3d":         "a1b2c3d",
		"a1b2c3d4e5f6":    "a1b2c3d",
		"  a1b2c3d4e5f6 ": "a1b2c3d",
	}
	for in, want := range tests {
		if got := shortCommit(in); got != want {
			t.Errorf("shortCommit(%q) = %q, want %q", in, got, want)
		}
	}
}
