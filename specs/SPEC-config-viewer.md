# Spec: config-viewer (OpenCode Tray)

> Traza al module id `config-viewer` del `CAPABILITY-MAP.md`. Depende de `shell`.

## Objective

`config-viewer` permite **ver** (solo lectura) la configuración de OpenCode desde el panel: **MCP servers** (con su estado de conexión), **skills**, **agents** y el **JSON de configuración global**. Incluye "abrir en editor" y "abrir carpeta" para editar fuera de la app.

**Fuentes de datos:**
- MCP (con estado): `opencode mcp list`.
- Skills / agents: `GET /api/skill` y `GET /api/agent` (o, como respaldo, parseando la config).
- Config global cruda: archivo `~/.config/opencode/opencode.json(c)`.

**Editor:** si `code` está en el PATH → `code <archivo>`; si no → el editor por defecto del SO (`start` en Windows, `xdg-open` en Linux). "Abrir carpeta" abre el directorio de config en el explorador.

**Contrato público:**

```go
type MCPServer struct { Name, Type, Status string }
type Skill struct { Name string }
type Agent struct { Name string }

func (s *Service) ListMCP() ([]MCPServer, error)
func (s *Service) ListSkills() ([]Skill, error)
func (s *Service) ListAgents() ([]Agent, error)
func (s *Service) ConfigJSON() (string, error)   // texto JSONC crudo (con secretos enmascarados)
func (s *Service) ConfigPath() (string, error)
func (s *Service) OpenInEditor(path string) error
func (s *Service) OpenConfigDir() error
```

## Tech Stack

- Go 1.24+ (ejecuta `opencode mcp list` vía `servicecontrol`/wrapper; lee archivos con `os`; resuelve `code` con `exec.LookPath`).
- Frontend Vue 3 (sección "Config": listas + visor JSON + botones).
- Sin dependencias nuevas salvo un resaltado ligero de JSON en el frontend (opcional).

## Project Structure

```
internal/
└── configviewer/
    ├── configviewer.go   # struct Service + New()
    ├── mcp.go            # ListMCP (parsea `opencode mcp list`)
    ├── catalog.go        # ListSkills / ListAgents
    ├── config.go         # ConfigJSON / ConfigPath (lee ~/.config/opencode/opencode.jsonc)
    ├── mask.go           # enmascarado de secretos
    ├── editor.go         # OpenInEditor / OpenConfigDir (code → fallback SO)
    └── *_test.go

frontend/src/components/
└── config/               # sección "Config"
```

## Code Style

- Enmascarado de secretos centralizado en `mask.go`, aplicado a **todo** texto de config antes de enviarlo al frontend:

```go
// mask.go — oculta valores sensibles (keys/tokens/secretos) para no exponerlos en la UI
var secretKeys = []string{"apiKey", "apikey", "key", "token", "secret", "authorization", "password"}

func MaskSecrets(raw string) string { /* reemplaza valores de esas claves por "••••" */ }
```

- `OpenInEditor` prueba `code` y cae al SO, sin asumir editor fijo:

```go
func (s *Service) OpenInEditor(path string) error {
    if code, err := exec.LookPath("code"); err == nil {
        return exec.Command(code, path).Start()
    }
    return openDefault(path) // start (win) / xdg-open (linux)
}
```

## Testing Strategy

- **Go `testing`**: mockeando el runner de comandos y con archivos temporales. Casos:
  - `ListMCP` parsea la salida de `opencode mcp list` (incluido "needs authentication").
  - `ConfigJSON` lee un `opencode.jsonc` de prueba y enmascara keys/tokens.
  - `ConfigPath` resuelve la ruta correcta (homedir; verificable con variable de entorno fake).
  - `OpenInEditor` elige `code` si existe, si no el fallback (PATH fakeado).
- **Vue (Vitest)**: sección Config — listas renderizadas, visor JSON, botones de abrir.

## Boundaries

- **Always do:**
  - Solo lectura: jamás escribir en `opencode.json(c)`.
  - Enmascarar secretos antes de mostrar.
  - Usar `exec.LookPath` para resolver editores (no rutas hardcodeadas).
- **Ask first:**
  - Añadir edición inline de config (cambia el modelo de solo-lectura).
  - Mostrar config de proyecto además de la global.
- **Never:**
  - Mostrar API keys/tokens/passwords en claro.
  - Editar `service.json` ni la DB de OpenCode.

## Success Criteria

- [ ] Se listan los MCP (globales) con su estado de conexión.
- [ ] Se listan skills y agents.
- [ ] El visor muestra el `opencode.json(c)` global con secretos enmascarados.
- [ ] "Abrir en editor" usa VS Code si existe; si no, el editor por defecto.
- [ ] "Abrir carpeta" abre el directorio de config.
- [ ] Tests Go y Vitest en verde.

## Open Questions

1. **Estado de conexión MCP**: `opencode mcp list` da el estado pero puede iniciar el servicio. ¿Preferimos siempre `opencode mcp list`, o leer la config (sin estado) cuando el servicio esté detenido? → **Se decide en plan** (hecho de comportamiento); mi default: usar `mcp list` y tolerar que arranque el servicio.
2. **Config de proyecto además de global**: solo global por ahora (el usuario pidió "nivel global"); proyecto como mejora.
3. **Resaltado de JSON** en el visor: opcional (librería ligera en frontend).
