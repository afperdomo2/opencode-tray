package configviewer

import "regexp"

// secretKeyRe coincide el valor de claves sensibles en JSON/JSONC.
var secretKeyRe = regexp.MustCompile(`(?i)("(?:apiKey|apikey|key|token|secret|authorization|password)"\s*:\s*")([^"]*)(")`)

// MaskSecrets reemplaza el valor de claves sensibles por "••••" para no
// exponerlos en la UI. Es sobre texto (no parsea) para tolerar JSONC.
func MaskSecrets(raw string) string {
	return secretKeyRe.ReplaceAllString(raw, "${1}••••${3}")
}
