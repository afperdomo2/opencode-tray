# Spec: usage (OpenCode Tray)

> Traza al module id `usage` del `CAPABILITY-MAP.md`. Depende de `shell`. Es la fuente de datos de `notifications`.

## Objective

`usage` consulta el **consumo de OpenCode Go** (suscripción) y lo expone para el panel y para `notifications`. Muestra tres ventanas: **rolling (5 h)**, **weekly** y **monthly**, con su % consumido y su fecha de reinicio.

**Endpoint (confirmado):**

```
GET https://opencode.ai/zen/go/v1/usage
Authorization: Bearer <API_KEY>
```

Respuesta:

```json
{
  "usage": {
    "rolling": { "status": "ok", "percent": 43.0, "resetsAt": "2026-10-14T06:00:00Z" },
    "weekly":  { "status": "ok", "percent": 21.5, "resetsAt": "..." },
    "monthly": { "status": "ok", "percent": 8.2,  "resetsAt": "..." }
  }
}
```

- `percent` = **% consumido** (0–100). `resetsAt` = reinicio (RFC3339, puede ser `null`).
- Errores: `401` = key inválida; `403` con `error.type == "EntitlementError"` = sin suscripción Go activa.

**Resolución del token (en orden):**
1. Si hay **token manual** configurado en Ajustes → se usa ese.
2. Si no → se lee automáticamente de `~/.local/share/opencode/auth.json`, campo `opencode-go.key` (estructura `{"opencode-go":{"key":"..."}}`).
3. Si no hay ninguno → estado "sin token configurado" (la UI lo indica y ofrece pegar la key).

**Contrato público:**

```go
type Window struct { Percent float64; ResetsAt time.Time; HasReset bool }
type Usage struct { Rolling, Weekly, Monthly Window }

func (s *Service) Fetch() (Usage, error)          // GET + parseo
func (s *Service) ResolveToken() (string, error)  // manual > auth.json
func (s *Service) SetManualToken(tok string) error
func (s *Service) ClearManualToken() error
func (s *Service) Severity(u Usage) Level         // green | yellow | red (para el tray de shell)
```

## Tech Stack

- Go 1.24+ (`net/http` con timeout 10 s, `encoding/json`).
- Frontend Vue 3 (sección "Consumo" con barras de progreso).
- Sin dependencias nuevas; el HTTP se inyecta como interface para tests.

## Project Structure

```
internal/
└── usage/
    ├── usage.go       # struct Service + New()
    ├── fetch.go       # GET /zen/go/v1/usage + parseo
    ├── token.go       # ResolveToken / SetManualToken / lectura de auth.json
    ├── severity.go    # mapeo consumo → nivel
    └── *_test.go

frontend/src/components/
└── usage/             # sección "Consumo" (3 barras + % + reset)
```

## Code Style

- Parseo defensivo y tolerante a campos faltantes (`percent` puede ser `null`):

```go
// fetch.go
type apiWindow struct {
    Percent  *float64 `json:"percent"`
    ResetsAt *string  `json:"resetsAt"`
}

func (w apiWindow) toWindow() (Window, bool) {
    if w.Percent == nil {
        return Window{}, false
    }
    out := Window{Percent: *w.Percent}
    if w.ResetsAt != nil {
        if t, err := time.Parse(time.RFC3339Nano, *w.ResetsAt); err == nil {
            out.ResetsAt, out.HasReset = t, true
        }
    }
    return out, true
}
```

- `Severity` sigue la regla de color del tray: verde = normal, amarillo = alguna ventana ≥ 90 % o ritmo adelantado, rojo = ≥ 98 % o error/token inválido.

## Testing Strategy

- **Go `testing`** con `httptest.Server` (HTTP mockeado, sin red). Casos:
  - Parseo de respuesta válida (3 ventanas) y de `percent: null`.
  - `401` → error de key inválida; `403` + `EntitlementError` → "sin suscripción".
  - `ResolveToken`: manual > auth.json (auth.json en archivo temporal); sin ninguno → error.
  - `Severity` en cada banda (verde/amarillo/rojo).
- **Vue (Vitest)**: sección Consumo — 3 barras con % y reset; estado "sin token" con CTA para pegar key.

## Boundaries

- **Always do:**
  - Timeout en el HTTP (10 s) y manejo de errores por código de estado.
  - El token **nunca** se loguea ni se muestra completo (solo enmascarado en Ajustes).
  - Leer `auth.json` en modo solo-lectura (nunca escribirlo).
- **Ask first:**
  - Cambiar el endpoint o añadir métricas de coste en $ (hoy solo %).
  - Añadir cifrado DPAPI al token manual (mejora de seguridad, Windows).
- **Never:**
  - Enviar el token a otro host que no sea `opencode.ai`.
  - Commitear el token o el contenido de `auth.json`.

## Success Criteria

- [ ] `Fetch()` devuelve rolling/weekly/monthly con % y reset correctos.
- [ ] Token auto-resuelto desde `auth.json`; el manual tiene prioridad.
- [ ] La UI muestra 3 barras con % consumido y fecha de reinicio.
- [ ] Errores (401 / sin suscripción / sin token) se muestran con mensaje claro.
- [ ] `Severity` alimenta el color del icono del tray.
- [ ] Tests Go y Vitest en verde.

## Open Questions

1. **Ruta exacta de `auth.json` en Windows**: `~/.local/share/opencode/auth.json` (homedir). Se verifica con `opencode debug paths data` en implementación para resolver el data dir correcto en cada SO.
2. **Frecuencia de refresco**: default **15 min** + refresco manual (configurable en Ajustes). ¿Te parece bien 15 min por defecto?
3. **Cifrado del token manual**: DPAPI en Windows como mejora de seguridad posterior (no bloqueante).
