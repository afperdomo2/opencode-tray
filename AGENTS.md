# AGENTS.md — OpenCode Tray

App de bandeja (tray) para Windows que controla el servicio de OpenCode, el acceso web (local/red/remoto), muestra el consumo de OpenCode Go, permite ver la configuración (MCP/skills/agents) y emite notificaciones nativas.

Stack: **Wails v2 + Go 1.24+ + Vue 3 / Vite / TypeScript**.

## Comandos

```bash
wails dev                 # desarrollo (hot reload)
wails build               # build (.exe portable)
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

## Flujo de trabajo (SDD)

Las specs son la fuente de verdad. Ante un cambio de requisitos, alcance o arquitectura: **actualizar primero la spec afectada**, luego el plan/tareas, luego implementar. Nunca implementar algo que no esté en una spec/tarea aprobada.

## Notas técnicas (gotchas)

- **`wails dev` no sirve para validar frontend ↔ backend**: en modo dev el puente `window.go` puede quedar sin definir (quirk del proxy de Vite de Wails), haciendo fallar todas las llamadas al backend. **Validar siempre con el build de producción** (`wails build` → `build/bin/opencode-tray.exe`). Ver README → "Limitaciones conocidas".
- **El tray se registra en `main`, antes de `wails.Run`**: Wails v2.16 ejecuta `OnStartup` en otra goroutine. Si `energye/systray` se registra allí, la ventana `SystrayClass` queda en un hilo sin bucle de mensajes: el icono aparece pero el clic izquierdo y el menú contextual no responden. `app.initTray(trayIcon)` debe seguir en `main`.
- **Instalador**: `build/windows/installer/opencode-tray.iss` y `scripts/build-installer.ps1` se trackean; `build/bin/` no. Para compilar el instalador se requiere **Inno Setup 6.3+** (`ISCC.exe`). Ver README → "Build y despliegue (Windows)".

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
