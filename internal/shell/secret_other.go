//go:build !windows

package shell

import (
	"encoding/base64"
	"errors"
)

// ProtectSecret no está soportado fuera de Windows. Se usa base64 para no
// romper la compilación; la app solo distribuye binario Windows.
func ProtectSecret(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	return base64.StdEncoding.EncodeToString([]byte(plain)), nil
}

// UnprotectSecret invierte ProtectSecret en plataformas no Windows.
func UnprotectSecret(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	if len(raw) == 0 {
		return "", errors.New("blob vacío")
	}
	return string(raw), nil
}
