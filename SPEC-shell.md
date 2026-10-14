# Spec: shell (OpenCode Tray)

> Módulo base. Todo lo demás depende de él. Traza al module id `shell` del `CAPABILITY-MAP.md`.

## Objective

`shell` es la fundación de **OpenCode Tray**: una app de bandeja (tray) para Windows escrita con Wails que da soporte estructural al resto de módulos. No implementa lógica de OpenCode por sí misma; provee el andamiaje (ventana, tray, ajustes, idioma, arranque) al que se enchufan `service-control`, `web-access`, `usage`, `config-viewer` y `notifications`.

**Usuario**: un desarrollador que usa OpenCode a diario y quiere controlar su servicio, acceso web y consumo desde la bandeja, sin abrir la terminal.

**Qué hace `shell`:**
1. Icono en la bandeja del sistema (refleja estado general: servicio corriendo/detenido, alerta de consumo).
2. Al hacer clic izquierdo, abre un **panel popover compacto** anclado al tray (se cierra al hacer clic fuera o con Escape).
3. **Una sola instancia**: si se lanza dos veces, se enfoca la existente.
4. **Cerrar = minimizar al tray** (el proceso no sale; salir solo desde el menú del tray).
5. **Persistencia de ajustes** en un archivo JSON local (no en la config de OpenCode).
6. **Toggle "Iniciar con Windows"** (registro `HKCU\...\Run`).
7. **i18n español/inglés** con selector de idioma.
8. Define el **esqueleto de secciones** del panel (Resumen, Servicio, Acceso web, Consumo, Config, Ajustes) como contenedores vacíos que los demás módulos rellenan.

## Tech Stack

| Capa | Tecnología | Versión |
|---|---|---|
| Shell desktop | Wails | v2 (última estable) |
| Backend | Go | 1.24+ |
| Frontend | Vue 3 + Vite + TypeScript | Vue 3.x |
| Icono de bandeja | `energye/systray` | última |
| i18n | `vue-i18n` | 11.x |
| Tests frontend | Vitest + @vue/test-utils | última |

> La librería de tray se confirma en "Open Questions" (Wails v2 no trae tray nativo; se usa una lib externa cross-platform).

## Commands

```bash
# Scaffold inicial (una sola vez)
wails init -n opencode-tray -t vue-ts

# Desarrollo (hot reload del frontend)
wails dev

# Build de producción (genera el .exe portable)
wails build

# Frontend
cd frontend && npm install
cd frontend && npm run dev

# Tests backend (Go)
go test ./...

# Tests frontend
cd frontend && npm test          # vitest run
cd frontend && npm run test:watch
```

## Project Structure

```
opencode-tray/
├── CAPABILITY-MAP.md          # índice de módulos (ya existe)
├── SPEC-shell.md              # esta spec
├── SPEC-service-control.md    # (futuro)
├── SPEC-web-access.md         # (futuro)
├── SPEC-usage.md              # (futuro)
├── SPEC-config-viewer.md      # (futuro)
├── SPEC-notifications.md      # (futuro)
├── README.md                  # se mantiene en sincronía con lo de arriba
├── tasks/
│   ├── plan.md                # plan técnico
│   └── todo.md                # lista de tareas
├── design-system/
│   └── opencode-tray/
│       └── MASTER.md          # sistema de diseño (ui-ux-pro-max)
├── go.mod / go.sum
├── main.go                    # entry point Wails
├── app.go                     # setup de la ventana + tray + wiring de módulos
├── internal/
│   ├── shell/                 # ← este módulo (tray, window lifecycle, settings, autostart, i18n)
│   ├── servicecontrol/
│   ├── webaccess/
│   ├── usage/
│   ├── configviewer/
│   └── notifications/
└── frontend/
    ├── src/
    │   ├── App.vue            # shell del panel + pestañas/secciones
    │   ├── main.ts
    │   ├── components/        # componentes de sección (una carpeta por sección)
    │   ├── i18n/              # es.json, en.json
    │   ├── stores/            # estado (Pinia)
    │   └── lib/               # puente con el backend (wails bindings)
    └── vite.config.ts
```

## Code Style

Go: estándar + `gofmt`. Cada módulo expone una struct `Service` con métodos claros y un `New()` constructor; el shell lo cablea en `app.go`.

```go
// internal/shell/settings.go
package shell

type Settings struct {
    Language    string `json:"language"`     // "es" | "en"
    AutoStart   bool   `json:"autoStart"`
    RefreshMin  int    `json:"refreshMin"`   // intervalo de refresco (min)
}

func (s *Service) LoadSettings() (Settings, error) { /* ... */ }
func (s *Service) SaveSettings(st Settings) error    { /* ... */ }
```

Vue: `<script setup lang="ts">` + Composition API. Todo texto visible pasa por i18n (nunca strings hardcodeados en plantillas).

```vue
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
</script>

<template>
  <button class="btn">{{ t('actions.save') }}</button>
</template>
```

## Testing Strategy

- **Go**: tests table-driven por paquete (`_test.go`) con `testing` estándar. Cubren: carga/guardado de ajustes (round-trip), toggle de auto-inicio (con el registro mockeado), resolución de rutas y ciclo de vida de la ventana.
- **Vue**: Vitest + @vue/test-utils. Cubren: render de cada sección, cambio de idioma, persistencia del estado del panel.
- **Sin e2e** por ahora.
- Regla: una tarea no está "hecha" sin su test.

## Boundaries

- **Always do:**
  - Mantener sincronía entre README.md, specs, `tasks/` y estructura de directorios cuando algo cambie.
  - Todo texto de UI vía claves i18n.
  - Correr `go test ./...` y tests de frontend antes de dar por cerrado un cambio.
  - `gofmt` y lint de frontend.
- **Ask first:**
  - Agregar dependencias Go o npm.
  - Cambiar la librería de tray o el esquema de `Settings`.
  - Cambiar la estructura de carpetas del proyecto.
- **Never:**
  - Commitear secretos, tokens o la contraseña del servicio.
  - Hardcodear rutas de usuario (usar `os.UserConfigDir()` / homedir).
  - Escribir en la config de OpenCode (`opencode.json`) desde `shell`.

## Success Criteria

- [ ] `wails build` genera un único `.exe` portable que arranca en Windows 10+.
- [ ] El icono aparece en la bandeja; clic izquierdo abre el panel popover; Escape/clic-fuera lo cierra; cerrar minimiza al tray sin salir.
- [ ] Una segunda ejecución enfoca la instancia existente (single instance).
- [ ] Los ajustes (idioma, refresco, auto-inicio) persisten entre reinicios.
- [ ] El toggle de auto-inicio agrega/elimina la clave en `HKCU\...\Run` correctamente.
- [ ] El selector de idioma alterna es/en en toda la UI.
- [ ] `go test ./...` y `cd frontend && npm test` pasan en verde.

## Open Questions

1. **Librería de tray definitiva**: `energye/systray` (recomendada, cross-platform, usada por proyectos similares) vs `getlantern/systray`. Confirmar integración con Wails v2 y soporte Linux para la fase futura.
2. **Versión exacta de Wails v2 y Go** a fijar en `go.mod` (se confirma al hacer el scaffold).
3. **Ruta de ajustes**: `%APPDATA%\OpenCodeTray\settings.json` (Windows) — confirmar que es la preferida vs `~/.config/opencode-tray`.
