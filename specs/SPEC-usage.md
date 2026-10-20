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

## Resolución del token

OpenCode v2 guarda las credenciales nuevas en SQLite (`opencode.db`, tabla `credential`); `auth.json` es legado y puede quedar desactualizado. La app **nunca escribe** en ninguno de los dos.

Orden de candidatos (auto primero):

1. **Token exportado de OpenCode v2** (`opencode auth export`), cacheado **solo en memoria** durante la sesión.
2. **`auth.json` legado**, releyéndolo en cada consulta. Se aceptan los proveedores `opencode-go` y `opencode` (estructura `{"<provider>":{"type":"api","key":"..."}}`).
3. **Token manual** guardado en Ajustes, cifrado con Windows DPAPI. Es **respaldo**: solo se usa si OpenCode no tiene un token válido.

`Fetch` prueba los candidatos en orden y usa el primero que responda `200`:

- `401` → prueba el siguiente candidato.
- `403` → prueba el siguiente candidato (otra cuenta puede tener suscripción).
- Error de red → se devuelve sin seguir probando.
- Sin candidatos → `USAGE_NO_TOKEN`.

Errores estables con prefijo para que el frontend los traduzca:

```
USAGE_NO_TOKEN: ...
USAGE_INVALID_TOKEN: ...
USAGE_NO_ENTITLEMENT: ...
USAGE_FORBIDDEN: ...
USAGE_TRANSPORT: ...
USAGE_HTTP_<code>: ...
USAGE_PARSE: ...
```

**Reintento de token:**

- `ReloadToken()` invalida el caché de export y ejecuta `opencode auth export`.
  - Primero sin `--standalone` con timeout corto; si falla, con `--standalone` y timeout amplio.
  - Parsea la salida (array u objeto), elige `opencode-go` y luego `opencode`, solo credenciales `active`.
  - Nunca loguea stdout/stderr ni incluye secretos en errores.
- El frontend reintenta automáticamente **una vez** ante `USAGE_NO_TOKEN`/`USAGE_INVALID_TOKEN`; si falla, muestra un botón manual.
- No se reintenta automáticamente ante `USAGE_NO_ENTITLEMENT`, red o parseo.

**Caché de datos en frontend:**

`usage` no cachea datos; el panel usa un composable nativo `useQuery` con:

- `staleTime` 60 s y deduplicación de peticiones en vuelo.
- Datos previos conservados durante refrescos y errores.
- Skeleton **solo** cuando no hay datos (`isInitialLoading`).
- Polling según `RefreshMin` (Ajustes; por defecto 15 min), pausado con la ventana oculta y refresco al recuperar foco si está stale.
- Sin persistencia en disco: tras reiniciar, el primer fetch puede mostrar skeleton.

## Contrato público

```go
type Window struct { Percent float64; ResetsAt time.Time; HasReset bool }
type Usage struct { Rolling, Weekly, Monthly Window }

type TokenStatus struct {
    Configured bool
    Source     string // opencode-export | opencode-json | manual | none
    Provider   string // opencode-go | opencode | ""
}

func (s *Service) Fetch() (Usage, error)                       // prueba candidatos
func (s *Service) TokenStatus() TokenStatus                    // sin red
func (s *Service) ReloadToken(ctx context.Context) (TokenStatus, error)
func (s *Service) SetManualToken(tok string)                   // solo memoria
func (s *Service) ClearManualToken()
func (s *Service) ValidateToken(tok string) error              // petición directa
func (s *Service) Severity(u Usage) Level
```

Wiring en `app.go`:

```go
func (a *App) UsageTokenStatus() usage.TokenStatus
func (a *App) UsageReloadToken() (usage.TokenStatus, error)
func (a *App) UsageSetManualToken(tok string) error
func (a *App) UsageClearManualToken() error
```

## Tech Stack

- Go 1.24+ (`net/http` con timeout 10 s, `encoding/json`).
- Frontend Vue 3 (widget de Consumo con barras de progreso).
- Sin dependencias nuevas: DPAPI con `golang.org/x/sys/windows` (ya presente) y CLI oficial de OpenCode.

## Project Structure

```
internal/
└── usage/
    ├── usage.go        # struct Service + New()
    ├── fetch.go        # GET /zen/go/v1/usage + candidatos
    ├── token.go        # auth.json, prioridad, TokenStatus
    ├── export.go       # opencode auth export + parser
    ├── exec_windows.go # consola oculta
    ├── severity.go     # mapeo consumo → nivel
    └── *_test.go

frontend/src/
├── lib/query.ts        # caché nativo tipo TanStack
├── lib/settings.ts     # ajustes compartidos
├── lib/usage.ts        # useUsageQuery + reintento de token
└── components/
    ├── Resumen.vue
    └── Ajustes.vue
```

## Code Style

- Parseo defensivo y tolerante a campos faltantes (`percent` puede ser `null`).
- Nunca loguear ni devolver el token completo; el input manual es write-only.
- Secretos del export se capturan en memoria y se descartan.

## Testing Strategy

- **Go `testing`** con `httptest.Server` (HTTP mockeado, sin red) y runner de export mockeado:
  - Parseo de respuesta válida (3 ventanas) y de `percent: null`.
  - `401` → error de key inválida; `403` + `EntitlementError` → "sin suscripción".
  - Candidatos: export → auth.json → manual; 401 en el primero y 200 en el segundo.
  - `auth.json` con proveedores `opencode` y `opencode-go`.
  - Parser de `opencode auth export` (array/objeto, activos/inactivos, key/oauth).
  - DPAPI round-trip y blob inválido.
  - `Severity` en cada banda.
- **Vue (Vitest)**:
  - Caché: remontar dentro de `staleTime` no muestra skeleton ni repite fetch.
  - Reintento automático una vez y botón manual.
  - Estado sin token/inválido con CTA; entrada manual y selector de intervalo.

## Boundaries

- **Always do:**
  - Timeout en el HTTP (10 s) y en `opencode auth export` (25–30 s).
  - El token **nunca** se loguea ni se muestra completo.
  - Leer `auth.json`/SQLite en modo solo-lectura (nunca escribirlos).
  - Preferir el token vigente de OpenCode; el manual es respaldo.
- **Ask first:**
  - Cambiar el endpoint o añadir métricas de coste en $ (hoy solo %).
  - Añadir una dependencia SQLite (hoy se usa el CLI oficial).
- **Never:**
  - Enviar el token a otro host que no sea `opencode.ai`.
  - Commitear el token, el blob DPAPI o el contenido de `auth.json`.
  - Cachear el token de `auth.json` en memoria/disco.

## Success Criteria

- [ ] `Fetch()` devuelve rolling/weekly/monthly con % y reset correctos.
- [ ] Lee tokens de `auth.json` con `opencode` y `opencode-go`.
- [ ] Lee el token v2 con `opencode auth export` bajo demanda y lo cachea en memoria.
- [ ] Reintenta una vez ante token ausente/inválido y ofrece botón manual.
- [ ] El token manual se guarda cifrado (DPAPI) y funciona como respaldo.
- [ ] El panel no muestra skeleton si ya hay datos cacheados.
- [ ] Tests Go y Vitest en verde.

## Open Questions

1. ~~Ruta exacta de `auth.json` en Windows~~: resuelta, `opencode debug paths data` → `~/.local/share/opencode`; se añaden fallbacks `%LOCALAPPDATA%`/`%APPDATA%`.
2. ~~Frecuencia de refresco~~: `RefreshMin`, por defecto 15 min, configurable en Ajustes.
3. **Cifrado del token manual**: DPAPI implementado.

---

# Iteración 8 — Humanizador del reinicio y caché compartida

## Objective

Exponer un texto de reinicio legible para humanos y compartir la caché de consumo entre el panel y el watcher del tray.

## Contrato añadido

```go
type Icons struct { Clock, Calendar, Reset, Warning string }
var EmojiIcons = Icons{...}
var SafeIcons  = Icons{...}

type TooltipOptions struct {
    Language string
    Icons    Icons
    Now      time.Time
}

func Tooltip(u Usage, opts TooltipOptions) string
func HumanReset(w Window, now time.Time, lang string) string
func BatteryBar(percent float64) string
func TruncateUTF16(s string, max int) string
```

Reglas de `HumanReset`:

| Caso | Texto |
|---|---|
| `!HasReset` | `sin reinicio` / `no reset` |
| `resetsAt <= now` | `ahora` / `now` |
| Mismo día, < 1 min | `en <1 min` / `in <1 min` |
| Mismo día, < 1 h | `en N min` / `in N min` |
| Mismo día | `en N h` / `in N h` |
| Mañana | `mañana a las HH:MM` / `tomorrow at h:MM PM` |
| 2+ días | `el D mmm AA` / `Mmm D, AAAA` |

- `now` y la zona horaria se inyectan; `resetsAt` se convierte a la zona local.
- `BatteryBar`: 5 celdas, `floor(percent/20)`, tope 0–100.
- `Tooltip`: `OpenCode · Mensual <barra> N% · <icono> <reset>` y truncado a 127 UTF-16.

## Caché

- `App` mantiene `usageCache`, `usageAt` y `usageOK` bajo mutex.
- `UsageFetch` devuelve caché si tiene menos de 60 s.
- `backgroundRefresh` hace fetch solo si pasó `RefreshMin` (mínimo 1 min).
- `ReloadToken`, `SetManualToken` y `ClearManualToken` invalidan la caché.
- Un error no borra el último consumo cacheado (el tooltip conserva el dato anterior).

## Testing Strategy

- Tabla de `HumanReset` con reloj y zona fijos, en español e inglés.
- `BatteryBar` en 0, 19, 20, 42, 99, 100.
- `TruncateUTF16` con emoji y límite 127.
- `Tooltip` completo en ambos idiomas y con iconos seguros.

## Success Criteria

- [ ] El tooltip usa el idioma configurado y la hora local.
- [ ] Nunca supera el límite de `szTip`.
- [ ] El panel y el watcher comparten la misma caché.
- [ ] Tests en verde.

