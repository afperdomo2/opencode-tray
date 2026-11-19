# AGENTS.md — OpenCode Tray

App de bandeja (tray) para Windows (y Linux para desarrollo y prueba) que controla el servicio de OpenCode, el acceso web (local/red/remoto), muestra el consumo de OpenCode Go, permite ver la configuración (MCP/skills/agents) y emite notificaciones nativas.

Stack: **Wails v2.16 + Go 1.25+ + Vue 3 / Vite / TypeScript**.

## Comandos

```bash
wails dev                 # desarrollo (hot reload)
wails build               # Windows: .exe portable
./scripts/build-linux.sh  # Linux: ELF (wails build -tags webkit2_41)
./scripts/build-deb.sh    # Linux: paquete .deb en dist/ (instala icono en el menú)
./scripts/generate-icons.sh  # iconos PNG/ICO desde brand/ (Chrome + Pillow)
go test ./...             # tests backend
cd frontend && pnpm test   # tests frontend (Vitest)
powershell -ExecutionPolicy Bypass -File scripts/build-installer.ps1  # portable + instalador Inno Setup
```

## Estructura

- `specs/` — `CAPABILITY-MAP.md` (índice de módulos) + `SPEC-<module-id>.md` por módulo. Fuente de verdad del **qué**.
- `tasks/` — `plan.md` (plan técnico) y `todo.md` (tareas). Fuente de verdad del **cómo**.
- `design-system/` — sistema de diseño generado con la skill `ui-ux-pro-max`.
- `brand/` — fuentes SVG de los iconos (`icon.svg` tile de app + `mark.svg` glifo). **Fuente única**.
- `docs/screenshots/` — capturas del panel en WebP lossless usadas en `README.md` (se commitean).
- `internal/<módulo>` — backend Go: `shell`, `servicecontrol`, `webaccess`, `usage`, `configviewer`, `notifications`.
- `buildinfo.go` — versión y compilación expuestas al frontend (`GetAppVersion`); embebe `wails.json`.
- `frontend/` — UI Vue 3.
- `build/windows/installer/opencode-tray.iss` — definición del instalador Inno Setup (se trackea).
- `scripts/build-installer.ps1` — build portable + instalador (se trackea).
- `scripts/build-linux.sh` — build para Linux (`-tags webkit2_41`); se trackea.
- `scripts/build-deb.sh` — empaqueta `.deb` en `dist/` (`/usr/bin`, `.desktop` e icono); se trackea.
- `scripts/generate-icons.sh` — regenera los PNG/ICO desde `brand/` (Chrome + Pillow); se trackea.
- `build/linux/icon.png` — icono de bandeja en Linux (64 px, PNG; `image.Decode` no lee ICO).
- `build/windows/tray.ico` — icono de bandeja en Windows (multi-tamaño, tile verde).
- `build/windows/icon.ico` — icono del `.exe` y del instalador (multi-tamaño, tile oscuro).
- `build/linux/icon-256.png` y `build/linux/opencode-tray.desktop` — icono y lanzador del menú para el `.deb`.
- `dist/` — paquetes generados; gitignored.

## Flujo de trabajo (SDD)

Las specs son la fuente de verdad. Ante un cambio de requisitos, alcance o arquitectura: **actualizar primero la spec afectada**, luego el plan/tareas, luego implementar. Nunca implementar algo que no esté en una spec/tarea aprobada.

## Notas técnicas (gotchas)

- **`wails dev` no sirve para validar frontend ↔ backend**: en modo dev el puente `window.go` puede quedar sin definir (quirk del proxy de Vite de Wails), haciendo fallar todas las llamadas al backend. **Validar siempre con el build de producción** (`wails build` → `build/bin/opencode-tray.exe`; `./scripts/build-linux.sh` → `build/bin/opencode-tray`). Ver README → "Limitaciones conocidas".
- **El tray se registra en `main`, antes de `wails.Run`**: Wails v2.16 ejecuta `OnStartup` en otra goroutine. Si `energye/systray` se registra allí, la ventana `SystrayClass` queda en un hilo sin bucle de mensajes: el icono aparece pero el clic izquierdo y el menú contextual no responden. `app.initTray(trayIcon)` debe seguir en `main`.
- **Instalador**: `build/windows/installer/opencode-tray.iss` y `scripts/build-installer.ps1` se trackean; `build/bin/` no. Para compilar el instalador se requiere **Inno Setup 6.3+** (`ISCC.exe`). Ver README → "Build — Windows".
- **Versión**: `Info.productVersion` de `wails.json` es la fuente única; `buildinfo.go` embebe ese archivo y lo expone con `GetAppVersion()` (tarjeta **Ajustes → Acerca de**). Wails compila con `-buildvcs=false`, así que el commit, la fecha y la marca dirty se inyectan con `-ldflags` desde `scripts/build-linux.sh` y `scripts/build-installer.ps1`; no quitar ese paso si se quieren distinguir builds.
- **Iconos**: no editar a mano `build/appicon.png`, `build/windows/{icon,tray}.ico` ni `build/linux/icon*.png`. Se generan con `./scripts/generate-icons.sh` desde `brand/` y se commitean. `build/linux/icon.png` y `build/windows/tray.ico` se embeben con `//go:embed`: hay que regenerarlos antes de compilar. El `.exe` y el instalador usan `build/windows/icon.ico` (Wails/Inno Setup); la bandeja usa `tray.ico`/`icon.png` (tile verde). `icon_test.go` valida tamaños y formato.
- **`StartHidden` no funciona en Linux (Wails v2.16)**: `internal/frontend/desktop/linux/window.go` llama a `Hide` y acto seguido a `gtk_widget_show_all`, así que la ventana aparece al arrancar. `App.startup` la oculta explícitamente en no-Windows; no quitar esa guarda.
- **Linux/Ubuntu**: requiere `build-essential`, `pkg-config`, `libgtk-3-dev` y `libwebkit2gtk-4.1-dev`, más `CGO_ENABLED=1`. Ubuntu 24.04+ solo trae WebKitGTK 4.1 → compilar con `-tags webkit2_41` (lo hace `scripts/build-linux.sh`). El panel no se ancla al tray: abre centrado y se arrastra desde la franja superior (Wayland puede ignorar el centrado). El token manual se guarda en base64 (sin DPAPI). Detalle en `specs/SPEC-shell.md` → Iteraciones 10, 11 y 12.
- **Nada de `<select>` nativo en la UI**: en Linux/WebKitGTK abre una ventana GTK que roba el foco; el `blur` del frontend hacía que el panel se ocultara al cambiar el intervalo de refresco. Usa `SegmentedControl.vue` o controles in-page.
- **CLI de OpenCode**: se resuelve con `internal/cli.Resolve()` (`$OPENCODE_BIN` → `PATH` → `~/.opencode/bin` → `~/.local/bin` → `/usr/local/bin` → `/usr/bin` → `/snap/bin`). **No hardcodees `"opencode"`** en los constructores: desde el menú o el auto-inicio el `PATH` gráfico no incluye el shell del usuario.
- **Instancia única (Linux)**: `flock` en `$XDG_RUNTIME_DIR/opencode-tray-instance.lock` y socket `.sock` para que una segunda ejecución pida mostrar el panel. El lanzador del `.deb` usa `--show-panel`; el auto-inicio arranca oculto.

## Reglas

**Always do:**
- Nada está "hecho" sin su test (`go test ./...` / `pnpm test`).
- Mantener sincronía de archivos: si un cambio toca comandos, estructura o config global, actualizar en el mismo cambio `README.md`, `specs/SPEC-*.md`, `tasks/plan.md` y `tasks/todo.md`.
- Todo texto de UI vía claves i18n (`frontend/src/i18n/{es,en}.json`), nunca hardcodeado.
- Enmascarar secretos antes de mostrarlos en la UI.
- Comentarios del código en **español**, nunca en inglés (identificadores —variables, funciones, tipos— en inglés; comentarios y strings de UI en español salvo i18n).
- `gofmt` y lint de frontend antes de cerrar un cambio.

**Ask first:**
- Agregar dependencias Go o de paquetes (pnpm).
- Cambiar el esquema de `Settings`, la librería de tray o la de notificaciones.
- Cambiar la estructura de carpetas.
- Añadir edición inline de config (hoy es solo lectura).

**Never:**
- Commitear secretos, tokens o la contraseña del servicio.
- Editar a mano `service.json`, la DB de OpenCode o `opencode.json` (usar el CLI de OpenCode).
- Hardcodear rutas de usuario (usar `os.UserConfigDir()` / homedir).
- Loggear el token de OpenCode Go ni mostrarlo completo.
- Usar métodos **variádicos** en interfaces del backend que Wails expone (p. ej. `Run(args ...string)`): rompen la generación de bindings **en silencio** (no escribe `frontend/wailsjs/go`). Usar `[]string` en su lugar.
