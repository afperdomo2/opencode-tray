//go:build windows

package shell

import (
	"encoding/base64"
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ProtectSecret cifra un secreto con DPAPI (ámbito del usuario actual) y lo
// devuelve como blob base64 para poder guardarlo en settings.json.
func ProtectSecret(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	data := []byte(plain)
	in := windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
	var out windows.DataBlob
	if err := windows.CryptProtectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return "", err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))

	blob := unsafe.Slice(out.Data, int(out.Size))
	return base64.StdEncoding.EncodeToString(blob), nil
}

// UnprotectSecret descifra un blob DPAPI generado por ProtectSecret.
func UnprotectSecret(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	if len(raw) == 0 {
		return "", errors.New("blob DPAPI vacío")
	}
	in := windows.DataBlob{Size: uint32(len(raw)), Data: &raw[0]}
	var out windows.DataBlob
	if err := windows.CryptUnprotectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return "", err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))

	return string(unsafe.Slice(out.Data, int(out.Size))), nil
}
