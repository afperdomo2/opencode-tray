//go:build windows

package shell

import "testing"

func TestSecretRoundTrip(t *testing.T) {
	blob, err := ProtectSecret("sk-token-de-prueba")
	if err != nil {
		t.Fatalf("ProtectSecret() error: %v", err)
	}
	if blob == "" || blob == "sk-token-de-prueba" {
		t.Fatal("el blob debería estar cifrado y no vacío")
	}
	plain, err := UnprotectSecret(blob)
	if err != nil {
		t.Fatalf("UnprotectSecret() error: %v", err)
	}
	if plain != "sk-token-de-prueba" {
		t.Errorf("plain = %q", plain)
	}
}

func TestSecretEmpty(t *testing.T) {
	if blob, err := ProtectSecret(""); err != nil || blob != "" {
		t.Errorf("ProtectSecret vacío = %q, %v", blob, err)
	}
	if plain, err := UnprotectSecret(""); err != nil || plain != "" {
		t.Errorf("UnprotectSecret vacío = %q, %v", plain, err)
	}
}

func TestUnprotectInvalid(t *testing.T) {
	if _, err := UnprotectSecret("no-es-base64"); err == nil {
		t.Error("blob inválido debería fallar")
	}
}
