# Spec: shell (OpenCode Tray)

> Módulo base. Todo lo demás depende de él. Traza al module id `shell` del `CAPABILITY-MAP.md`.

## Objective

`shell` es la fundación de **OpenCode Tray**: una app de bandeja (tray) para **Windows y Linux** escrita con Wails que da soporte estructural al resto de módulos. No implementa lógica de OpenCode por sí misma; provee el andamiaje (ventana, tray, ajustes, idioma, arranque) al que se enchufan `service-control`, `web-access`, `usage`, `config-viewer` y `notifications`.

**Usuario**: un desarrollador que usa OpenCode a diario y quiere controlar su servicio, acceso web y consumo desde la bandeja, sin abrir la terminal.

**Qué hace `shell`:**
1. Icono en la bandeja del sistema. Hoy es **estático**; la variación por estado (servicio corriendo/detenido, alerta de consumo) queda como fase futura (ver Iteración 13).
2. Al hacer clic izquierdo, abre un **panel popover compacto** anclado al tray. Permanece abierto al hacer clic fuera; se oculta con la **X** de su cabecera, con `Escape` o con un segundo clic en el icono.
3. **Una sola instancia**: si se lanza dos veces, se enfoca la existente.
4. **Cerrar = minimizar al tray** (el proceso no sale; salir solo desde el menú del tray).
5. **Persistencia de ajustes** en un archivo JSON local (no en la config de OpenCode).
6. **Toggle "Iniciar con Windows"** (registro `HKCU\...\Run`; en Linux, entrada XDG Autostart).
7. **i18n español/inglés** con selector de idioma.
8. Define el **esqueleto de secciones** del panel (Resumen, Servicio, Acceso web, Consumo, Config, Ajustes) como contenedores vacíos que los demás módulos rellenan.

## Tech Stack

| Capa | Tecnología | Versión |
|---|---|---|
| Shell desktop | Wails | v2 (última estable) |
| Backend | Go | 1.25+ (`go.mod`) |
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
cd frontend && pnpm install
cd frontend && pnpm dev

# Tests backend (Go)
go test ./...

# Tests frontend
cd frontend && pnpm test          # vitest run
cd frontend && pnpm test:watch

# Iconos (solo si cambia brand/)
./scripts/generate-icons.sh       # Chrome/Chromium + Pillow
```

## Project Structure

```
opencode-tray/
├── README.md                  # puerta de entrada; enlaza a specs/ y tasks/
├── specs/
│   ├── CAPABILITY-MAP.md      # índice de módulos (ya existe)
│   ├── SPEC-shell.md          # esta spec
│   ├── SPEC-service-control.md
│   ├── SPEC-web-access.md
│   ├── SPEC-usage.md
│   ├── SPEC-config-viewer.md
│   ├── SPEC-notifications.md
│   └── SPEC-resumen.md
├── tasks/
│   ├── plan.md                # plan técnico
│   └── todo.md                # lista de tareas
├── design-system/
│   └── opencode-tray/
│       └── MASTER.md          # sistema de diseño (ui-ux-pro-max)
├── brand/
│   ├── icon.svg               # tile de app (fuente única)
│   └── mark.svg               # glifo del prompt (bandeja)
├── build/                     # iconos generados por plataforma (se commitean)
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
    │   ├── lib/               # puente con el backend + query cache nativo
    │   │   ├── query.ts       # caché de datos (stale-while-revalidate)
    │   │   ├── settings.ts    # ajustes compartidos
    │   │   └── usage.ts       # consulta y reintento de token
    │   └── stores/            # (reservado; hoy no se usa Pinia)
    └── vite.config.ts
```

## Code Style

Go: estándar + `gofmt`. Cada módulo expone una struct `Service` con métodos claros y un `New()` constructor; el shell lo cablea en `app.go`.

```go
// internal/shell/settings.go
package shell

type Settings struct {
    Language    string `json:"language"`              // "es" | "en"
    AutoStart   bool   `json:"autoStart"`
    RefreshMin  int    `json:"refreshMin"`            // intervalo de refresco (min)
    UsageToken  string `json:"usageToken,omitempty"`  // blob DPAPI (nunca se envía al frontend)
}

func (s *Service) LoadSettings() (Settings, error) { /* ... */ }
func (s *Service) SaveSettings(st Settings) error    { /* ... */ }
func (s *Service) SetRefreshMin(min int) error       { /* ... */ }
```

El token manual de OpenCode Go se cifra con **Windows DPAPI** (`CryptProtectData`) y se guarda como blob base64 en `settings.json`. `GetSettings` lo excluye siempre; el frontend solo recibe el estado (`TokenStatus`), nunca el secreto. El token de OpenCode (auth.json/SQLite) no se persiste en los ajustes.
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
  - Agregar dependencias Go o de paquetes (pnpm).
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
- [ ] El token manual de OpenCode Go persiste cifrado (DPAPI) y nunca se expone vía `GetSettings`.
- [ ] El intervalo de refresco (`RefreshMin`) se puede cambiar desde Ajustes y el polling lo respeta.
- [ ] `go test ./...` y `cd frontend && pnpm test` pasan en verde.

## Open Questions

1. **Librería de tray definitiva**: `energye/systray` (recomendada, cross-platform, usada por proyectos similares) vs `getlantern/systray`. Confirmar integración con Wails v2 y soporte Linux para la fase futura.
2. **Versión exacta de Wails v2 y Go** a fijar en `go.mod` (se confirma al hacer el scaffold).
3. **Ruta de ajustes**: `%APPDATA%\OpenCodeTray\settings.json` (Windows) — confirmar que es la preferida vs `~/.config/opencode-tray`.

---

# Iteración 5 — Flyout nativo anclado al icono

> Traza al module id `shell`. Backend + configuración de ventana. Sustituye el comportamiento de ventana flotante por un flyout tipo Windows 11.
>
> **Nota:** el alto variable de esta iteración fue sustituido por el alto fijo de la Iteración 7.

## Objective

Que el panel se sienta como un flyout nativo del sistema: aparece pegado al icono del tray, no se puede mover ni redimensionar, crece con su contenido y se cierra al perder el foco.

## Cambios

1. **Ventana frameless y fija.**
   - `Frameless: true`, `DisableResize: true`, `AlwaysOnTop: true`.
   - Ancho fijo **380 px**; alto variable entre **320 px** y **600 px**.
   - Sin título, sin botones de ventana y sin arrastre.
   - Fuera de la barra de tareas y de Alt+Tab (`WS_EX_TOOLWINDOW`).
   - Esquinas redondeadas/sombra nativas en Windows 11; cuadrado aceptable en Windows 10.

2. **Anclaje al icono.**
   - El clic izquierdo captura la posición del cursor y ancla el panel al borde de la taskbar.
   - Soporta taskbar abajo, arriba, izquierda y derecha, y múltiples monitores.
   - Se recalcula la posición en cada apertura y al cambiar el alto.
   - Posicionamiento en píxeles físicos vía Win32 (el `WindowSetPosition` de Wails usa coordenadas relativas al work area y no escala DPI).

3. **Alto según contenido.**
   - El frontend mide el contenido y solicita el alto al backend.
   - El panel crece hacia arriba manteniendo fijo el borde inferior.
   - El alto se acota y se actualiza con debounce para evitar saltos durante cargas asíncronas.

4. **Cierre nativo.**
   - `Esc`, pérdida de foco y clic fuera ocultan el panel.
   - Clic en el icono con el panel abierto lo cierra.
   - Guarda anti-carrera: el blur producido por el mismo clic que va al tray no reabre el panel.
   - El ítem “Mostrar panel” del menú abre el panel anclado al cursor.

5. **Herramientas Win32.**
   - Localizar el HWND mediante `WindowClassName` propio.
   - `GetCursorPos`, `MonitorFromPoint`, `GetMonitorInfo`, `SetWindowPos`, `SetForegroundWindow`.
   - Sin dependencias Go nuevas (se usa `golang.org/x/sys/windows`, ya presente).

## Fuera de alcance

- Preview al pasar el cursor (hover).
- Tooltip dinámico con el consumo.
- Mica/Acrylic.

## Testing Strategy

- **Go**: geometría de anclaje table-driven (4 bordes de taskbar, clamping, multi-monitor); controlador de ventana con reloj inyectado para la guarda de blur; clamp de alto.
- **Manual**: DPI 100/150/200 %, taskbar en las 4 posiciones, dos monitores, Windows 10/11.
- **Build**: `go test ./...`, `pnpm test`, `pnpm build` y `wails build`.

## Success Criteria

- [ ] El panel aparece pegado al icono del tray y no se puede mover ni redimensionar.
- [ ] No aparece en la barra de tareas ni en Alt+Tab.
- [ ] El alto se adapta al contenido y crece hacia arriba.
- [ ] `Esc`, clic fuera y clic en el tray cierran el panel.
- [ ] Funciona con taskbar en las 4 posiciones y con varios monitores/DPI.
- [ ] Tests Go y Vitest en verde; `wails build` OK.

---

# Iteración 6 — Cierre estable y tooltip de consumo

> Traza al module id `shell`. Corrige el cierre inmediato del flyout y añade consumo al tooltip nativo del tray.

## Objective

Que un clic en el icono deje el panel abierto y enfocado, sin que el blur transitorio del cambio de foco lo cierre; y que al pasar el cursor por el icono se vea el consumo actual.

## Cambios

1. **Guarda de blur posterior a la apertura.**
   - `ShowAt` registra el instante de apertura.
   - Un blur dentro de la ventana de gracia (~400 ms) no oculta de inmediato: se comprueba si el flyout sigue siendo la ventana en primer plano.
   - Si el foco volvió al flyout, el panel permanece; si se perdió de verdad, se oculta al vencer la guarda.
   - Los temporizadores pendientes se cancelan al mostrar/ocultar para evitar cierres obsoletos.

2. **Detección nativa de foco.**
   - `GetForegroundWindow` + `GetAncestor(GA_ROOT)` comparan la ventana activa con el HWND del flyout.
   - Si el sistema no concede el foco al mostrar, `FocusFlyout` refuerza la activación.

3. **Tooltip dinámico de consumo.**
   - Tras cada `UsageFetch` correcto se actualiza el tooltip del tray con los tres porcentajes.
   - Formato compacto y neutral: `OpenCode · 5h 18% · 7d 42% · 30d 27%`.
   - Sin datos o con error se conserva el tooltip anterior.

## Fuera de alcance

- Preview visual (popover) al pasar el cursor; el hover usa el tooltip nativo.
- Historial o gráficas en el tooltip.

## Testing Strategy

- **Go**: controlador con reloj y temporizador inyectados: blur transitorio con foreground, blur real, fuera de la guarda, cancelación de timers; formateo del tooltip y actualización tras `UsageFetch`.
- **Manual**: clic abre y permanece; clic fuera cierra; clic en tray con panel abierto cierra; `Esc` cierra; hover muestra consumo.
- **Build**: `go test ./...`, `pnpm test`, `pnpm build` y `wails build`.

## Success Criteria

- [ ] Un clic en el tray muestra el panel y este permanece abierto y enfocado.
- [ ] Un clic fuera, `Esc` o un segundo clic en el tray lo cierran.
- [ ] El tooltip muestra el consumo tras la primera lectura correcta.
- [ ] Tests Go y Vitest en verde; `wails build` OK.

---

# Iteración 7 — Alto fijo del panel

> Traza a `shell` + `ui-ux`. Deja sin efecto el punto 3 y el criterio de alto de la Iteración 5.

## Objective

Que el flyout conserve el mismo alto en todas las pestañas, como en Ajustes, para evitar que la ventana cambie de tamaño al navegar o cuando termina una carga asíncrona.

## Cambios

1. **Alto fijo global.**
   - `FlyoutMaxHeight` (600 px) pasa a ser el alto fijo del panel.
   - `main.go` y `NewWindowController` inicializan la ventana con ese alto.
   - El frontend deja de medir contenido, de observar `ResizeObserver` y de llamar a `SetFlyoutHeight` al cambiar de sección.
   - Solo `.app-content` hace scroll cuando el contenido excede el alto disponible.

2. **Anclaje estable.**
   - La ventana mantiene el borde inferior fijo y se reancla al icono en cada apertura.
   - `SetHeight` permanece disponible y acotado, pero el flujo normal no lo usa.

## Fuera de alcance

- Alto configurable por el usuario.
- Persistencia de un alto distinto entre sesiones.

## Testing Strategy

- **Go**: el controlador arranca con `FlyoutWidth × FlyoutMaxHeight`; el clamp de `SetHeight` sigue cubierto.
- **Vitest**: App no solicita cambios de alto al cambiar de pestaña.
- **Manual**: las 4 pestañas conservan el mismo alto; las secciones largas hacen scroll.
- **Build**: `go test ./...`, `pnpm test`, `pnpm build` y `wails build`.

## Success Criteria

- [ ] El panel mantiene 600 px de alto en todas las pestañas.
- [ ] La ventana no crece ni se encoge al navegar o al cargar datos.
- [ ] El contenido largo hace scroll sin mover la nav inferior.
- [ ] Tests y build en verde.

---

# Iteración 8 — Tooltip mensual y reinicio humanizado

> Traza a `shell` + `usage`. Sustituye el punto 3 de la Iteración 6.

## Objective

Que el hover del icono del tray comunique de un vistazo el **consumo mensual** y **cuándo se reinicia**, sin abrir el panel.

## Cambios

1. **Formato del tooltip.**
   - `OpenCode · Mensual ▰▰▱▱▱ 42% · 🕒 en 4 h`
   - `OpenCode · Mensual ▰▰▱▱▱ 42% · 🗓 mañana a las 12:30`
   - `OpenCode · Mensual ▰▰▱▱▱ 42% · 🗓 el 12 jun 26`
   - `OpenCode · Mensual ▰▰▱▱▱ 42% · ↻ sin reinicio`
   - La barra representa **% consumido** (5 celdas, `floor(pct/20)`), no batería restante.

2. **Humanizador (`internal/usage/humanize.go`).**
   - Hoy: `en 47 min` / `en 4 h`.
   - Mañana: `mañana a las 12:30`.
   - 2+ días: `el 12 jun 26`.
   - Sin reset: `sin reinicio`; vencido: `ahora`.
   - Español 24 h, inglés 12 h (`tomorrow at 12:30 PM`, `Jun 12, 2026`).
   - Sin librerías nuevas; reloj inyectable para tests.

3. **Iconos Unicode.**
   - Preferidos: `🕒` (hoy), `🗓` (mañana/futuro), `↻` (sin reset), `⚠` (error).
   - Fallback documentado: `◷`, `▦`, `↻`, `!` si el shell no renderiza emoji.

4. **Límites nativos.**
   - Una sola línea y máximo 127 unidades UTF-16 (`szTip`).
   - Truncado sin partir pares sustitutos.

5. **Refresco sin abrir el panel.**
   - Watcher en Go: fetch inicial a los ~2 s + ticker de 60 s.
   - El ticker recalcula la cuenta atrás; el fetch real respeta `RefreshMin`.
   - Caché compartida con `UsageFetch` (TTL 60 s) para no duplicar peticiones.
   - Un reintento automático de token al arrancar si no hay token.
   - Al cambiar idioma se reconstruye el tooltip.
   - Un error conserva el último tooltip válido.

## Fuera de alcance

- Popover visual al hacer hover: `energye/systray` no expone evento hover.
- Iconos gráficos o color dentro del tooltip nativo.

## Testing Strategy

- **Go**: tabla del humanizador (es/en, hoy/mañana/fecha, sin reset, vencido), barra de consumo, truncado UTF-16 y tooltip completo.
- **Manual**: hover con la app recién abierta y el panel cerrado; cambio de idioma; error de red.
- **Build**: `go test ./...`, `pnpm test`, `pnpm build`, `wails build`.

## Success Criteria

- [ ] El tooltip muestra datos sin abrir el panel.
- [ ] La cuenta atrás se actualiza cada minuto.
- [ ] Cambiar idioma cambia el tooltip.
- [ ] Un error no pisa el último tooltip válido.
- [ ] Tests y build en verde.

---

# Iteración 9 — Tray registrado en el hilo principal

> Traza a `shell`. Corrige el fallo por el que el clic izquierdo no abría el panel y el clic derecho no mostraba el menú.

## Causa

Wails v2.16 ejecuta `OnStartup` en una **goroutine nueva**. Al registrar `energye/systray` dentro de `startup`, la ventana `SystrayClass` se creaba en un hilo que no tenía bucle de mensajes, mientras el bucle de Wails corría en el hilo principal. Los mensajes de clic se quedaban en la cola de un hilo que nunca los procesaba: el icono se veía, pero no respondía.

## Cambios

1. **Registro temprano.**
   - `main.go` llama a `app.initTray(trayIcon)` **antes** de `wails.Run`.
   - La ventana del tray se crea en el hilo principal, que luego ejecuta el bucle de mensajes de Wails.
2. **Callbacks seguros durante el arranque.**
   - `initTray` conecta `OnClick`/`OnShow`/`OnQuit` con guardas para `a.win` y `a.ctx` nulos.
   - `startup` ya no crea el tray; solo monta la ventana y arranca el watcher del tooltip.
3. **Documentación.**
   - `internal/shell/tray.go` explica el requisito de hilo.
   - `AGENTS.md` recoge el gotcha para evitar regresiones.

## Testing Strategy

- **Manual**: clic izquierdo abre el panel; segundo clic lo cierra; clic derecho muestra el menú (Mostrar panel / Salir).
- **Automatizado**: clic simulado con `PostMessage` al `SystrayClass` y verificación de `IsWindowVisible` del flyout y de la ventana `#32768` del menú.
- **Build**: `go test ./...`, `pnpm test`, `pnpm build`, `wails build`.

## Success Criteria

- [ ] El clic izquierdo abre y cierra el panel.
- [ ] El clic derecho abre el menú contextual.
- [ ] "Salir" cierra la app.
- [ ] Tests y build en verde.


---

# Iteración 10 — soporte de compilación y ejecución en Linux

> Traza a `shell`. Habilita compilar y probar la app en Ubuntu sin cambiar el comportamiento en Windows.

## Objective

Que OpenCode Tray se pueda compilar y ejecutar en Linux (Ubuntu 24.04+ con GNOME) para desarrollo y prueba, manteniendo Windows como plataforma de distribución principal.

## Plataformas

| Capacidad | Windows | Linux |
|---|---|---|
| Compilación | `wails build` | `wails build -tags webkit2_41` (Ubuntu 26.04 solo trae WebKitGTK 4.1) |
| Toolchain | MSVC/WebView2 | `gcc` + `pkg-config` + `libgtk-3-dev` + `libwebkit2gtk-4.1-dev` |
| Icono de bandeja | `build/windows/tray.ico` | `build/linux/icon.png` (64 px; `image.Decode` no lee ICO) |
| Bandera de bandeja | `StatusNotifierItem` vía DBus (GNOME AppIndicator) | ídem |
| Auto-inicio | `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` | `~/.config/autostart/opencode-tray.desktop` (XDG) |
| Anclaje del panel | Win32: posición junto al icono, taskbar, DPI y multi-monitor | mejor esfuerzo: esquina inferior derecha del monitor principal |
| Instancia única | mutex con nombre Win32 | **no soportada** (se documenta como limitación) |
| Secretos (token de Go) | DPAPI (`CryptProtectData`) | base64 (sin llavero del SO; se documenta como limitación) |
| Notificaciones | toasts (hoy deshabilitadas) | D-Bus, dependen del daemon del escritorio |
| Instalador | Inno Setup per-user | no aplica (binario ELF) |

## Cambios

1. **Icono por SO** (`main.go`): se embeben `build/windows/tray.ico` (tile verde) y `build/linux/icon.png`; `trayIcon()` elige según `runtime.GOOS`. El `.exe` y el instalador usan `build/windows/icon.ico` (tile oscuro).
2. **Auto-inicio XDG** (`internal/shell/autostart_xdg.go`): implementa `runKey` sobre `~/.config/autostart` escribiendo una entrada `.desktop` con `Exec=<ruta del ejecutable>`. El contrato de `AutoStart` (`SetEnabled` / `Enabled`) no cambia.
3. **Posición del panel** (`placement_windows.go`, `placement_other.go`): el backend de ventana de `app.go` delega en `positionFlyout`, que en Windows usa Win32 y en el resto usa `ScreenGetAll` + `WindowSetPosition` con `shell.CornerPosition` (geometría pura y testeada).
4. **Arranque oculto**: Wails en Linux ignora `StartHidden` (llama a `Hide` y después a `gtk_widget_show_all`), así que `startup` oculta la ventana explícitamente en no-Windows para que la app nazca solo con el icono de bandeja.
5. **i18n**: la clave `common.autoStart` pasa a texto neutro ("Iniciar con el sistema" / "Start with system").
6. **Documentación**: README con requisitos y limitaciones de Linux; `AGENTS.md` con los comandos.

## Restricciones conocidas (Linux)

- **Wayland** puede ignorar `WindowSetPosition`: el panel aparecerá donde decida el compositor. En X11/XWayland se coloca en la esquina inferior derecha.
- En GNOME el clic izquierdo sobre el icono puede no emitir `Activate`: el panel se abre desde **Mostrar panel** en el menú del tray.
- Sin `DPAPI`: el token manual de OpenCode Go se guarda en base64 (no cifrado).
- El auto-inicio XDG depende de que el escritorio respete `~/.config/autostart`.

## Testing Strategy

- **Go**: `autostart_xdg_test.go` (round-trip de la entrada `.desktop` en un directorio temporal, borrado y comando vacío) y `TestCornerPosition` en `flyout_test.go` (tamaño normal, tamaños inválidos y panel mayor que la pantalla).
- **Build**: `go test ./...`, `pnpm test`, `pnpm build`, `wails build -tags webkit2_41` en Ubuntu.
- **Manual**: lanzar el binario, abrir el panel desde el tray, comprobar servicio/acceso web/consumo y el toggle de auto-inicio.

## Success Criteria

- [ ] `go test ./...` y `pnpm test` pasan en Linux.
- [ ] `wails build -tags webkit2_41` produce `build/bin/opencode-tray` ejecutable en Ubuntu.
- [ ] El icono del tray se dibuja (PNG decodificado) y el menú abre el panel.
- [ ] El toggle de auto-inicio crea y elimina `~/.config/autostart/opencode-tray.desktop`.
- [ ] En Windows se sigue usando el registro y el `.ico` sin cambios de comportamiento.

---

# Iteración 11 — instalación en Ubuntu (`.deb`), instancia única y resolución del CLI

> Traza a `shell`. Permite usar la app como cualquier aplicación del escritorio: icono en el menú, paquete `.deb`, una sola instancia y sin pasos manuales de entorno.

## Objective

Que en Ubuntu se pueda instalar con `apt`, lanzar desde el menú de aplicaciones y usar a diario, sin terminal y sin configurar el `PATH` a mano.

## Cambios

1. **Paquete `.deb`** (`scripts/build-deb.sh`):
   - `/usr/bin/opencode-tray`
   - `/usr/share/applications/opencode-tray.desktop` (`Exec=opencode-tray --show-panel`, `Icon=opencode-tray`, `StartupWMClass=opencode-tray`)
   - `/usr/share/icons/hicolor/256x256/apps/opencode-tray.png`
   - `Depends: libgtk-3-0t64 | libgtk-3-0, libwebkit2gtk-4.1-0`; `Recommends: gnome-shell-extension-appindicator`.
   - Versión desde `Info.productVersion` de `wails.json`; salida en `dist/` (gitignored).
2. **Resolución del CLI** (`internal/cli`): `Resolve()` busca `$OPENCODE_BIN`, luego `PATH`, y después `~/.opencode/bin`, `~/.local/bin`, `/usr/local/bin`, `/usr/bin` y `/snap/bin`. `servicecontrol`, `usage` y `configviewer` lo usan en sus constructores. Motivo: una app lanzada desde el menú o el auto-inicio no hereda el `PATH` del shell, donde suele vivir `opencode`.
3. **Instancia única en Linux** (`internal/shell/singleinstance_other.go`): `flock` sobre `$XDG_RUNTIME_DIR/opencode-tray-instance.lock` (se libera solo si el proceso muere) y `RequestShow`/`ServeShowRequests` con un socket Unix en `$XDG_RUNTIME_DIR/opencode-tray-instance.sock`.
4. **`--show-panel`** (`main.go` + `App.showOnStart`): el lanzador del menú abre el panel al arrancar; el auto-inicio (sin flag) arranca oculto. Si ya hay una instancia, la segunda envía `show` por el socket y termina.

## Comportamiento al lanzar

| Situación | Resultado |
|---|---|
| Primer arranque desde el menú | Arranca y abre el panel |
| Ya corriendo (auto-inicio) + clic en el menú | La segunda instancia pide `show` y sale; el panel se abre |
| Auto-inicio de sesión | Arranca oculto, solo icono del tray |
| Segunda instancia en terminal | Sale en silencio (<50 ms) |

## Fuera de alcance

- Auto-actualización del paquete (se reinstala el `.deb`).
- Firma GPG del paquete y repositorio APT (instalación local).
- Empaquetado para otras distros (`.rpm`, AppImage).

## Testing Strategy

- **Go**: `internal/cli` (precedencia `OPENCODE_BIN`/`PATH`/rutas conocidas, directorios ignorados, fallback a `opencode`) y `internal/shell` (lock excluyente, liberación y re-adquisición; `ServeShowRequests` + `RequestShow`; sin servidor devuelve `false`).
- **Manual/E2E**: `.deb` generado y validado con `desktop-file-validate` y `dpkg-deb -c`; binario ejecutado con un `PATH` gráfico simulado (sin `~/.opencode/bin`) verificando con `strace` que se ejecuta `/home/<user>/.opencode/bin/opencode`; segunda ejecución con `--show-panel` termina en <50 ms y abre el panel de la primera (`xwininfo` → `IsViewable`).
- **Build**: `go test ./...`, `pnpm test`, `pnpm build`, `./scripts/build-deb.sh`.

## Success Criteria

- [ ] `sudo apt install ./dist/opencode-tray_<versión>_amd64.deb` instala sin dependencias rotas.
- [ ] El icono aparece en el menú de aplicaciones y abre el panel al lanzarlo.
- [ ] La app encuentra `opencode` desde el menú y el auto-inicio sin symlinks ni cambios en el `PATH`.
- [ ] Una segunda ejecución no duplica el icono del tray.
- [ ] `sudo apt remove opencode-tray` deja `~/.config/OpenCodeTray` intacto.

---

# Iteración 12 — pulido de la experiencia en Ubuntu

> Traza a `shell`. Corrige el cierre del panel al usar un `<select>`, añade el consumo al menú del tray y hace el panel centrado y arrastrable.

## Objective

Que la app se sienta usable a diario en Ubuntu/GNOME, donde el escritorio impone límites al tray: sin tooltip nativo, sin clic izquierdo y sin anclaje al icono.

## Problemas y solución

1. **El panel se ocultaba al cambiar el intervalo de refresco.** El `<select>` nativo abría una ventana GTK que robaba el foco; el `blur` del frontend llamaba a `WindowBlurred` y el panel se ocultaba (parecía que la app se cerraba).
   - **Solución:** `SegmentedControl.vue` (control segmentado `1 · 5 · 15 · 30 · 60 min`) sin ventanas nativas. Test que verifica que Ajustes ya no contiene ningún `<select>`.
2. **El clic izquierdo del icono abre el menú en GNOME.** La extensión AppIndicator nunca emite `Activate`, así que `energye/systray` no puede distinguir el clic.
   - **Solución:** no es arreglable en GNOME; se documenta y `Mostrar panel` queda como primera entrada del menú. En KDE/XFCE el clic funciona.
3. **No hay tooltip al pasar el cursor.** La propiedad `ToolTip` del StatusNotifierItem se publica correctamente (verificado por D-Bus), pero GNOME no la dibuja.
   - **Solución:** fuera de Windows la primera fila del menú (deshabilitada) replica el texto del tooltip y se actualiza con `SetTooltip`. En Windows y KDE se mantiene el tooltip nativo.
4. **La ventana no se podía mover.** Frameless sin región de arrastre.
   - **Solución:** `CenterPosition` centra el panel la primera vez y `--wails-draggable: drag` en una franja superior de 22 px (solo no-Windows). Después de la primera apertura no se vuelve a posicionar, para conservar la posición elegida.

## Cambios técnicos

- `frontend/src/components/SegmentedControl.vue` (nuevo) + estilos `.seg` / `.seg-opt` / `.seg-unit`; `Ajustes.vue` lo usa y `onRefresh` pasa a recibir el número.
- `internal/shell/tray.go`: ítem de estado deshabilitado fuera de Windows; `SetTooltip` actualiza tooltip nativo **y** la fila del menú (protegido con mutex).
- `internal/shell/flyout.go`: `CenterPosition` (pura y testeada).
- `placement_other.go`: centra una sola vez por proceso (`atomic.Bool`).
- `App.Platform()` expuesto al frontend; `App.vue` muestra `.drag-strip` si no es Windows.
- Versión 1.3.1 (`.deb` con la corrección).

## Testing Strategy

- **Vitest**: `segmented-control.spec.ts` (render, emisión, opción activa, etiquetas accesibles); `ajustes.spec.ts` (sin `<select>`, `SetRefreshMin(5)` al pulsar `5`); `app.spec.ts` (franja de arrastre solo con `Platform() !== 'windows'`).
- **Go**: `TestCenterPosition` (centrado, tamaños inválidos, panel mayor que la pantalla).
- **Manual**: cambiar el intervalo sin que el panel desaparezca; arrastrar el panel y reabrir (mantiene la posición); `GetLayout` del menú muestra la fila de consumo deshabilitada y actualizada.
- **Build**: `go test ./...`, `pnpm test`, `pnpm build`, `./scripts/build-deb.sh`.

## Success Criteria

- [ ] Cambiar el intervalo de refresco no oculta el panel.
- [ ] El menú muestra el consumo como primera fila (solo lectura) y se actualiza.
- [ ] El panel aparece centrado y se puede arrastrar; reabrirlo conserva la posición.
- [ ] En Windows no cambia nada (sin franja, sin fila de consumo, tooltip nativo, panel anclado).

---

# Iteración 13 — identidad visual e iconos propios

> Traza a `shell`. Sustituye el icono por defecto de Wails por una identidad propia, con fuentes SVG y generación reproducible.

## Objective

Que la app tenga un icono reconocible en la bandeja, el ejecutable, el instalador y el menú de aplicaciones, en lugar del placeholder “W” de Wails. La fuente de verdad son dos SVG versionados; los PNG/ICO se generan con un script manual y se commitean.

## Cambios

1. **Fuente única SVG** (`brand/`):
   - `brand/icon.svg`: tile de app cuadrado de 512, radio 112, fondo `#1E293B`, borde interior `#94A3B8` al 28 % y glifo `>_` en `#22C55E`.
   - `brand/mark.svg`: solo el glifo, con `currentColor` para reutilizarlo en verde u oscuro.
2. **Generación reproducible** (`scripts/generate-icons.sh`):
   - Rasteriza con Chrome/Chromium headless y normaliza/ensambla con Python 3 + Pillow.
   - Genera `build/appicon.png` (1024), `build/windows/icon.ico` y `build/windows/tray.ico` (16/24/32/48/64/128/256), `build/linux/icon-256.png` (256) y `build/linux/icon.png` (64).
   - Supersampling 4× para tamaños ≤ 64 px; ICO con BMP 32-bit para tamaños chicos y PNG para 128/256.
   - Idempotente: los PNG se guardan sin metadatos y dos corridas producen los mismos bytes.
3. **Uso por plataforma**:
   - `.exe` e instalador: `build/windows/icon.ico` (tile oscuro). Inno Setup sigue con `SetupIconFile=..\icon.ico`.
   - Bandeja de Windows: `build/windows/tray.ico` (tile verde sólido + glifo oscuro), embebido en `main.go`.
   - Bandeja de Linux/macOS: `build/linux/icon.png` (64 px, tile verde), embebido en `main.go`.
   - Menú de aplicaciones y `.deb`: `build/linux/icon-256.png`.
4. **Test de assets** (`icon_test.go`): valida que el PNG embebido sea 64×64, que ambos ICO declaren los siete tamaños y que los PNG de app/menú tengan el tamaño esperado.
5. **El icono sigue siendo estático.** La variación por estado (corriendo/detenido/alerta) no se implementa en esta iteración; la frase de la Objective original queda acotada a esa fase futura.

## Decisiones de diseño

- **Prompt `>_`**: comunica “terminal/CLI” sin copiar el logo pixelado de OpenCode; sobrevive a 16 px con trazo de 72/512 (≈2,25 px).
- **Tile oscuro para la app** (`#1E293B`) y **tile verde para la bandeja** (`#22C55E`): el verde sólido se separa tanto de barras claras como oscuras; el tile oscuro conserva la identidad en el menú y el `.exe`.
- Glifo oscuro (`#0F172A`) sobre verde: contraste ≈8:1; verde sobre tile oscuro: ≈7:1.

## Fuera de alcance

- Iconos por estado del servicio o alerta de consumo.
- Favicon del frontend (el panel no es una pestaña del navegador).
- Iconos para macOS (no hay build `darwin` en el repo).
- Branding del instalador más allá del icono (`SetupIconFile`).

## Testing Strategy

- **Go**: `icon_test.go` (PNG embebido 64×64, ICO de bandeja y de app con 16/24/32/48/64/128/256, PNG de app/menú con su tamaño).
- **Script**: dos corridas consecutivas y comparación de `sha256`; verificación con Pillow de que el ICO se puede leer.
- **Manual**: revisión de la hoja de contacto a 16/24/32/64 px sobre fondo claro y oscuro; arranque en Linux y comprobación del icono real en la bandeja y el menú.
- **Build**: `go test ./...`, `./scripts/build-linux.sh` y `./scripts/build-deb.sh`.

## Success Criteria

- [ ] `brand/icon.svg` y `brand/mark.svg` son la única fuente; no se editan PNG/ICO a mano.
- [ ] `./scripts/generate-icons.sh` regenera los cinco assets y es idempotente.
- [ ] `go test ./...` valida tamaños y formato de los iconos.
- [ ] El `.exe`, el instalador, el menú de aplicaciones y la bandeja muestran el icono nuevo (sin la “W”).
- [ ] En Linux la bandeja usa el tile verde y el menú el tile oscuro.

---

# Iteración 14 — versión y build visibles

> Traza a `shell`. Permite saber qué versión y qué compilación concreta está instalada, sin abrir una terminal.

## Objective

Que en **Ajustes → Acerca de** se vea la versión de la app y la compilación instalada (commit, fecha y plataforma), para distinguir builds entre sí al probar o reportar un problema.

## Cambios

1. **Fuente única de versión**: `Info.productVersion` de `wails.json` sigue alimentando los metadatos del `.exe`, el instalador y el `.deb`. El binario **embebe `wails.json`** (`//go:embed`) y lee ese valor en runtime, de modo que incluso un `wails build` directo (sin scripts) muestra la versión correcta. Fallback: `dev`.
2. **Commit y fecha**: Wails compila con `-buildvcs=false`, así que Go no embebe el commit. `scripts/build-linux.sh` y `scripts/build-installer.ps1` calculan `git rev-parse --short HEAD`, `git status --porcelain` y la fecha UTC, y los inyectan con `-ldflags "-X main.buildCommit=… -X main.buildDate=… -X main.buildDirty=…"`. Sin git (tarball) se omiten sin fallar; la UI muestra “No disponible”.
3. **Contrato con el frontend**: `App.GetAppVersion()` devuelve `{ version, commit, date, dirty, platform }`. `platform` es `runtime.GOOS/runtime.GOARCH`; `date` va en RFC3339 UTC y el frontend la formatea con `Intl.DateTimeFormat`.
4. **UI**: componente `AcercaDe.vue` al final de Ajustes (debajo del grid, a ancho completo) con nombre, versión, compilación (commit corto + `*` si hay cambios sin commitear), fecha, plataforma y botón **Copiar información** (`ClipboardSetText`, confirmación de 1,5 s).
5. **i18n**: textos en `es.json`/`en.json` bajo el namespace `about`.
6. **Versión 1.4.0** en `wails.json` y `RELEASE.md`.

## Decisiones

- **Versión embebida, no inyectada**: `wails.json` es la única fuente; los scripts solo añaden commit/fecha. Evita que un build directo muestre `dev` o una versión distinta a los metadatos del `.exe`.
- **El parámetro `-Version` del instalador** se mantiene como etiqueta del instalador; la app siempre reporta `wails.json` y el script ya avisa si difieren.
- **Commit corto + marca dirty**: es lo que distingue dos builds de la misma versión durante las pruebas.

## Fuera de alcance

- Auto-actualización o verificación de versiones nuevas.
- Mostrar la versión en el tooltip del tray o en el menú contextual.
- Firma de código y metadatos de versión en Linux (el `.deb` ya lleva `Version`).
- Comando `--version` en la CLI.

## Testing Strategy

- **Go** (`buildinfo_test.go`): parseo de `Info.productVersion`, fallback `dev`, commit corto, marca dirty y plataforma.
- **Vitest** (`acerca-de.spec.ts`): render de versión/commit/fecha/plataforma, marca dirty, copiar con éxito y error, ausencia de fecha.
- **Ajustes** (`ajustes.spec.ts`): la tarjeta aparece al final y no altera el grid de ajustes.
- **Build**: `go test ./...`, `pnpm test`, `pnpm build`, `./scripts/build-linux.sh` y `./scripts/build-deb.sh`.

## Success Criteria

- [ ] Ajustes muestra `v1.4.0` al final, incluso con `wails build` directo.
- [ ] Con `./scripts/build-linux.sh` se ven commit y fecha; al modificar un archivo aparece la marca de cambios sin commitear.
- [ ] El botón copiar produce un texto con versión, compilación y plataforma.
- [ ] `go test ./...` y `pnpm test` en verde.
- [ ] El `.deb` y el instalador se generan con 1.4.0.

---

# Iteración 15 — cierre explícito con cabecera y botón X

> Traza a `shell` + `ui-ux`. Deja sin efecto el cierre por pérdida de foco/clic fuera de las Iteraciones 5 y 6.

## Objective

Que el panel se comporte como una ventana de escritorio normal: al abrir otra aplicación o hacer clic fuera **no se cierra**. El cierre pasa a ser una acción explícita del usuario mediante un botón **X** visible en una cabecera compacta.

## Cambios

1. **Sin cierre automático por foco.**
   - Se elimina el listener `window.blur` del frontend y el binding `WindowBlurred`.
   - Se eliminan del controlador `HideFromBlur`, `hideIfUnfocused`, la guarda de blur (`blurGuard`, `showBlurGuard`) y la detección de foreground (`IsForeground`/`FlyoutForeground`).
   - `ToggleAt` vuelve a ser un toggle puro: si está visible oculta, si está oculto muestra de inmediato (sin guarda de blur).

2. **Cabecera compacta con X.**
   - Nueva `<header class="app-header">` antes de `<main>` con el título (`app.title`) a la izquierda y un botón de cerrar a la derecha.
   - El botón usa el icono `x`, `aria-label`/`title` desde `common.close` y un target de al menos 28 px.
   - La X llama a `HideWindow`: **oculta al tray**, no termina el proceso (salir sigue en el menú del tray).
   - En Linux la cabecera es la zona de arrastre (`--wails-draggable: drag`); la X declara `no-drag` para que el clic no mueva la ventana. En Windows se mantiene el anclaje al icono y no hay arrastre.
   - Se elimina la antigua franja `.drag-strip`/`.drag-grip`.

3. **Atajos que se conservan.**
   - `Escape` sigue ocultando el panel; los modales siguen deteniendo la propagación.
   - El clic izquierdo en el icono del tray sigue alternando abrir/ocultar.
   - `Alt+F4` sigue ocultando al tray vía `HideWindowOnClose`.

4. **Orden de capas.** La ventana continúa `AlwaysOnTop`, sin entrada en la barra de tareas ni en Alt+Tab: al no cerrarse al perder el foco, debe seguir accesible por encima del resto.

## Fuera de alcance

- Barra de título nativa o botones de minimizar/maximizar.
- Comportamiento de z-order normal (el panel deja de ser topmost).
- Cerrar el proceso desde la X.

## Testing Strategy

- **Go**: el controlador ya no tiene estado de blur; test de que `Hide` + `ToggleAt` vuelve a mostrar de inmediato; `Toggle`, `ShowAt`, `Hide` y `SetHeight` siguen cubiertos.
- **Vitest**: la estructura es `header`, `main`, `nav`; la X y `Escape` llaman a `HideWindow`; un evento `blur` no oculta; la cabecera solo es arrastrable fuera de Windows.
- **Manual**: abrir el panel, cambiar a otra ventana y comprobar que sigue abierto; cerrar con X, reabrir desde el tray, `Escape`, arrastre desde la cabecera en Linux y `Escape` dentro de un modal.
- **Build**: `go test ./...`, `pnpm test`, `pnpm build`, `./scripts/build-linux.sh` y `wails build`.

## Success Criteria

- [ ] Hacer clic fuera del panel no lo cierra.
- [ ] La cabecera muestra el título y una X accesible; la X oculta el panel al tray.
- [ ] `Escape` y el clic en el tray siguen ocultando el panel.
- [ ] En Linux la ventana se arrastra desde la cabecera y la X no la arrastra.
- [ ] `go test ./...` y `pnpm test` en verde.
