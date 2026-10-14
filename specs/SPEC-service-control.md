# Spec: service-control (OpenCode Tray)

> Traza al module id `service-control` del `CAPABILITY-MAP.md`. Depende de `shell`.

## Objective

`service-control` encapsula el control del **servicio de fondo de OpenCode**: estado, arranque, parada, reinicio y el **force-stop** (workaround para el bug de `opencode service stop`). Es el único módulo que ejecuta comandos de ciclo de vida del servicio; el resto lo consume vía su API.

**Problema real que resuelve:** `opencode service stop` no siempre detiene el servicio. El workaround conocido es consultar `opencode api get /api/info`, obtener el `pid` y matar el proceso. Este módulo encapsula esa lógica en un `ForceStop()` fiable.

**Contrato público (lo que otros módulos usan):**

```go
type Service struct { /* ... */ }

func New() *Service

// Estado
func (s *Service) Status() (Status, error)          // running bool + info (version, pid)
func (s *Service) Info() (ServerInfo, error)        // JSON de /api/info

// Ciclo de vida
func (s *Service) Start() error
func (s *Service) Stop() error                       // stop "normal" (puede fallar)
func (s *Service) ForceStop() error                  // workaround: stop → verificar → matar pid
func (s *Service) Restart() error

// Configuración del servicio (para web-access)
func (s *Service) Get(key string) (string, error)     // hostname, port, password, remote, cors…
func (s *Service) Set(key, value string) error
func (s *Service) Unset(key string) error
```

`web-access` usará `Set("hostname", ...)`, `Set("port", ...)`, `SetPassword` y `Start/Stop`; el contrato vive aquí.

## Tech Stack

- **Go 1.24+**, usando `os/exec` para invocar el CLI `opencode` (asumimos que está en el PATH).
- Sin dependencias externas salvo las de Wails heredadas de `shell`.
- Abstracción de "matar proceso" por SO (Windows `taskkill`, Linux `kill`) mediante un archivo por build tag.

## Commands (que envuelve este módulo)

```bash
opencode service status
opencode service start
opencode service stop
opencode service restart
opencode service get <key>
opencode service set <key> <value>
opencode service unset <key>
opencode api get /api/info        # → { "pid": ..., ... }
```

## Project Structure

```
internal/
└── servicecontrol/
    ├── service.go       # struct Service + New()
    ├── status.go        # Status / Info
    ├── lifecycle.go     # Start / Stop / Restart / ForceStop
    ├── config.go        # Get / Set / Unset
    ├── exec.go          # wrapper de exec.Command con timeout + captura de salida
    ├── kill_windows.go  # taskkill /PID <pid> /F
    ├── kill_linux.go    # kill <pid>   (futuro)
    └── *_test.go        # tests table-driven
```

## Code Style

- Cada comando del CLI se ejecuta vía un único helper `run(...)` que aplica timeout (p. ej. 15 s) y devuelve stdout/stderr + error. Nunca se llama `exec.Command` suelto.
- `ForceStop` es explícito y con pasos claros, no lógica implícita:

```go
// lifecycle.go
func (s *Service) ForceStop() error {
    if err := s.Stop(); err == nil {
        // stop normal funcionó; verificar
        if st, e := s.Status(); e == nil && !st.Running {
            return nil
        }
    }
    info, err := s.Info() // /api/info
    if err != nil {
        return fmt.Errorf("no se pudo obtener pid: %w", err)
    }
    return killByPID(info.PID)
}
```

## Testing Strategy

- **Go `testing`**, table-driven. Se inyecta un `runner` (interface) en `Service` para mockear `exec`; los tests no invocan `opencode` real.
- Casos cubiertos:
  - `Status` parsea correctamente la salida (corriendo / detenido).
  - `ForceStop` usa `kill` cuando `Stop` falla o no detiene; no mata si ya está detenido.
  - `Info` parsea el `pid` del JSON de `/api/info`.
  - `Set/Get/Unset` forman los argumentos correctos del CLI.
  - `killByPID` selecciona `taskkill /F` en Windows / `kill` en Linux (por build tag, testeable en la plataforma actual).

## Boundaries

- **Always do:**
  - Usar el wrapper `run` con timeout (nunca `exec` crudo).
  - `ForceStop` como camino fiable; no dejar el servicio colgado.
  - Tests table-driven para cada comando.
- **Ask first:**
  - Cambiar la semántica de `ForceStop` o el timeout.
  - Agregar dependencias para manejo de procesos.
- **Never:**
  - Editar `~/.config/opencode/service.json` o `service.json` de registro a mano (siempre vía CLI).
  - Matar procesos que no sean el `pid` reportado por OpenCode.

## Success Criteria

- [ ] `Status()` devuelve correctamente si el servicio corre o no.
- [ ] `Start()` arranca el servicio y `Status()` lo refleja.
- [ ] `ForceStop()` deja el servicio **siempre** detenido (prueba manual: con el bug del `service stop`, queda muerto el proceso).
- [ ] `Info()` extrae el `pid` del JSON de `/api/info`.
- [ ] `Set/Get/Unset` leen y escriben hostname, port y password.
- [ ] `go test ./internal/servicecontrol/...` en verde (con runner mockeado, sin tocar `opencode` real).

## Open Questions

1. **Lectura de la contraseña**: `opencode service get password` (¿devuelve la contraseña real o está oculta?) vs leer `~/.config/opencode/service.json`. **Se verifica en implementación** (es un hecho, no una decisión de producto); el contrato `Get("password")` queda listo para cualquiera de las dos.
2. **Fiabilidad del bug**: confirmar si `opencode service stop` falla siempre o intermitente; `ForceStop` se diseña para cubrir ambos casos igual.
3. **Timeout por defecto** de `run`: 15 s razonable para comandos de servicio (arranque puede tardar más; se ajusta si hace falta).
