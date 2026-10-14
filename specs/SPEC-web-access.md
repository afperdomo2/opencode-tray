# Spec: web-access (OpenCode Tray)

> Traza al module id `web-access` del `CAPABILITY-MAP.md`. Depende de `shell` y `service-control`.

## Objective

`web-access` permite controlar **cómo se accede al servidor web de OpenCode** desde el panel, en tres modos, y gestionar las **credenciales** (usuario fijo `opencode` + contraseña).

**Modos:**
1. **Local** — el servidor escucha solo en `127.0.0.1` (solo esta máquina). Config por defecto.
2. **Red (por IP)** — el servidor escucha en `0.0.0.0` (todas las interfaces). El panel muestra la(s) URL(s) de acceso: IP(s) LAN de la máquina + puerto.
3. **Remoto (túnel)** — acceso desde internet vía OpenTunnel (`opencode pair --remote`). El panel muestra el enlace del túnel.

**Credenciales:** muestra el usuario (`opencode`, fijo) y la contraseña actual, con opción de **mostrar/ocultar** y de **cambiarla**.

Este módulo **no** ejecuta el CLI directamente; usa el contrato de `service-control` (`Set/Get/Unset` + `Start/Stop`).

## Tech Stack

- Go 1.24+ (usa `servicecontrol.Service` y `net` para detectar IPs locales).
- Frontend Vue 3 (sección "Acceso web" del panel).
- Sin dependencias nuevas salvo las heredadas.

## Commands (subyacentes, vía `service-control`)

```bash
opencode service set hostname 127.0.0.1   # local
opencode service set hostname 0.0.0.0     # red
opencode service set port <puerto>
opencode service set password "<secreto>"
opencode pair --remote --yes              # remoto (OpenTunnel)
opencode service set remote false         # apagar remoto
opencode service start                    # aplicar config (un set detiene el servicio)
```

## Project Structure

```
internal/
└── webaccess/
    ├── webaccess.go      # struct Service + New()
    ├── modes.go          # SetMode(local|network|remote)
    ├── urls.go           # ComputeURLs() → IPs LAN + puerto
    ├── credentials.go    # CurrentPassword() / ChangePassword() / Username()
    └── *_test.go

frontend/src/components/
└── web-access/           # sección "Acceso web" (selector de modo, credenciales)
```

## Code Style

- La lógica de "qué argumentos genera cada modo" vive en Go, aislada y testeable:

```go
// modes.go
type Mode string

const (
    ModeLocal   Mode = "local"
    ModeNetwork Mode = "network"
    ModeRemote  Mode = "remote"
)

func (s *Service) ApplyMode(m Mode) (Mode, error) {
    switch m {
    case ModeLocal:
        return m, s.sc.Set("hostname", "127.0.0.1")
    case ModeNetwork:
        if err := s.sc.Set("hostname", "0.0.0.0"); err != nil {
            return "", err
        }
        return m, nil
    case ModeRemote:
        // opencode pair --remote --yes vía service-control
        return m, s.enableRemote()
    }
    return "", fmt.Errorf("modo inválido: %s", m)
}
```

- La contraseña nunca se muestra en logs; el frontend la pinta enmascarada por defecto (campo `type=password` con toggle).

## Testing Strategy

- **Go `testing`**: mockeando `servicecontrol.Service` (interface). Casos:
  - `ApplyMode` forma el hostname correcto por modo.
  - `ComputeURLs` devuelve IP(s) LAN + puerto (inyectando IPs fake, sin tocar red real).
  - `ChangePassword` llama a `Set("password", ...)` y reinicia el servicio.
  - `Username()` devuelve siempre `"opencode"`.
- **Vue (Vitest)**: sección Acceso web — render de modo actual, toggle mostrar/ocultar contraseña, y llamada a `ChangePassword` con confirmación.

## Boundaries

- **Always do:**
  - Usar el contrato de `service-control` (nunca `exec` directo aquí).
  - Contraseña enmascarada por defecto; nunca loguear secretos.
  - Tras `set` de config, arrancar el servicio (un `set` lo detiene).
- **Ask first:**
  - Cambiar la semántica de "remoto" (p. ej. usar otro proveedor de túnel).
  - Exponer opciones avanzadas de túnel en la UI.
- **Never:**
  - Habilitar `0.0.0.0` o remoto sin mostrarlo claramente en la UI (riesgo de seguridad).
  - Mostrar la contraseña sin acción explícita del usuario.

## Success Criteria

- [ ] Alternar local ↔ red ↔ remoto funciona y el estado queda reflejado en el panel.
- [ ] En modo red se muestran las URL(s) `http://<IP-LAN>:<puerto>` correctas.
- [ ] En modo remoto se muestra un enlace de túnel que funciona.
- [ ] El usuario `opencode` y la contraseña se ven (con mostrar/ocultar).
- [ ] Cambiar la contraseña persiste y el servicio arranca con la nueva.
- [ ] Tests Go y Vitest en verde.

## Open Questions

1. **URL del túnel persistente**: `opencode pair --remote` imprime un enlace de un solo uso/5 min. ¿Hay forma de obtener la URL del túnel estable (p. ej. `opencode service get remote`)? **Se verifica en implementación** (hecho, no decisión); mientras, el panel regenera el enlace al activar remoto.
2. **Detección de IPs LAN**: usar `net.Interfaces()` + filtrar IPv4 privadas no-loopback. Detalle de implementación; se decide la heurística exacta en el plan.
3. **Puerto por defecto**: OpenCode usa `49374`; confirmar si el usuario quiere poder fijar otro puerto en Ajustes (probable mejora, no bloqueante).
