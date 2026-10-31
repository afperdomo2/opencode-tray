# AGENTS.md — OpenCode Tray

App de bandeja (tray) para Windows (y Linux para desarrollo y prueba) que controla el servicio de OpenCode, el acceso web (local/red/remoto), muestra el consumo de OpenCode Go, permite ver la configuración (MCP/skills/agents) y emite notificaciones nativas.

Stack: **Wails v2.16 + Go 1.25+ + Vue 3 / Vite / TypeScript**.

## Comandos

```bash
wails dev                 # desarrollo (hot reload)
wails build               # Windows: .exe portable
./scripts/build-linux.sh  # Linux: ELF (wails build -tags webkit2_41)
go test ./...             # tests backend
cd frontend && pnpm test   # tests frontend (Vitest)
powershell -ExecutionPolicy Bypass -File scripts/build-installer.ps1  # portable + instalador Inno Setup
```

## Estructura

- `specs/` — `CAPABILITY-MAP.md` (índice de módulos) + `SPEC-<module-id>.md` por módulo. Fuente de verdad del **qué**.
- `tasks/` — `plan.md` (plan técnico) y `todo.md` (tareas). Fuente de verdad del **cómo**.
- `design-system/` — sistema de diseño generado con la skill `ui-ux-pro-max`.
- `internal/<módulo>` — backend Go: `shell`, `servicecontrol`, `webaccess`, `usage`, `configviewer`, `notifications`.
- `frontend/` — UI Vue 3.
- `build/windows/installer/opencode-tray.iss` — definición del instalador Inno Setup (se trackea).
- `scripts/build-installer.ps1` — build portable + instalador (se trackea).
- `scripts/build-linux.sh` — build para Linux (`-tags webkit2_41`); se trackea.
- `build/linux/icon.png` — icono de bandeja en Linux (64 px, PNG; `image.Decode` no lee ICO).

## Flujo de trabajo (SDD)

Las specs son la fuente de verdad. Ante un cambio de requisitos, alcance o arquitectura: **actualizar primero la spec afectada**, luego el plan/tareas, luego implementar. Nunca implementar algo que no esté en una spec/tarea aprobada.

## Notas técnicas (gotchas)

- **`wails dev` no sirve para validar frontend ↔ backend**: en modo dev el puente `window.go` puede quedar sin definir (quirk del proxy de Vite de Wails), haciendo fallar todas las llamadas al backend. **Validar siempre con el build de producción** (`wails build` → `build/bin/opencode-tray.exe`; `./scripts/build-linux.sh` → `build/bin/opencode-tray`). Ver README → "Limitaciones conocidas".
- **El tray se registra en `main`, antes de `wails.Run`**: Wails v2.16 ejecuta `OnStartup` en otra goroutine. Si `energye/systray` se registra allí, la ventana `SystrayClass` queda en un hilo sin bucle de mensajes: el icono aparece pero el clic izquierdo y el menú contextual no responden. `app.initTray(trayIcon)` debe seguir en `main`.
- **Instalador**: `build/windows/installer/opencode-tray.iss` y `scripts/build-installer.ps1` se trackean; `build/bin/` no. Para compilar el instalador se requiere **Inno Setup 6.3+** (`ISCC.exe`). Ver README → "Build — Windows".
- **`StartHidden` no funciona en Linux (Wails v2.16)**: `internal/frontend/desktop/linux/window.go` llama a `Hide` y acto seguido a `gtk_widget_show_all`, así que la ventana aparece al arrancar. `App.startup` la oculta explícitamente en no-Windows; no quitar esa guarda.
- **Linux/Ubuntu**: requiere `build-essential`, `pkg-config`, `libgtk-3-dev` y `libwebkit2gtk-4.1-dev`, más `CGO_ENABLED=1`. Ubuntu 24.04+ solo trae WebKitGTK 4.1 → compilar con `-tags webkit2_41` (lo hace `scripts/build-linux.sh`). El panel no se ancla al tray (esquina inferior derecha, y Wayland puede ignorarlo), no hay instancia única y el token manual se guarda en base64 (sin DPAPI). Detalle en `specs/SPEC-shell.md` → Iteración 10.

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
