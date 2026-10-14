# Tareas — OpenCode Tray

> Phase 3 del flujo SDD. Ordenadas por dependencia (Fase 0 → 5). Cada tarea: criterio de aceptación, verificación y archivos.
> Regla transversal: **ninguna tarea está "hecha" sin su test**, y todo cambio que toque comandos/estructura actualiza README.md + specs + este archivo.

---

## Fase 0 — Scaffold

- [x] **Task: Scaffold del proyecto Wails + Vue TS**
  - Acceptance: proyecto Wails v2 con plantilla `vue-ts`, git inicializado, `go test ./...` y `pnpm test` configurados (aunque vacíos), README inicial.
  - Verify: `wails dev` arranca; `wails build` genera `.exe`; `go test ./...` y `cd frontend && pnpm test` corren.
  - Files: `go.mod`, `main.go`, `app.go`, `frontend/**`, `README.md`, `.gitignore` (generados por `wails init -n opencode-tray -t vue-ts`).

---

## Fase 1 — shell

- [x] **Task: Icono de bandeja + ciclo de vida del panel**
  - Acceptance: icono en el tray; clic izquierdo abre el panel popover; Escape/clic-fuera lo cierra; cerrar minimiza al tray; una segunda ejecución enfoca la existente (single-instance).
  - Verify: correr la app y ejecutar las acciones manualmente.
  - Files: `internal/shell/tray.go`, `internal/shell/window.go`, `app.go`.

- [x] **Task: Persistencia de ajustes**
  - Acceptance: `LoadSettings`/`SaveSettings` leen y escriben `%APPDATA%\OpenCodeTray\settings.json`; valores por defecto correctos.
  - Verify: `go test ./internal/shell/...` (round-trip de ajustes).
  - Files: `internal/shell/settings.go`, `internal/shell/settings_test.go`.

- [x] **Task: Toggle "Iniciar con Windows"**
  - Acceptance: activar/desactivar agrega/elimina la clave en `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`.
  - Verify: test con registro mockeado + comprobación manual del `regedit`.
  - Files: `internal/shell/autostart_windows.go`, `internal/shell/autostart_test.go`.

- [x] **Task: i18n español/inglés**
  - Acceptance: selector de idioma alterna es/en en toda la UI; textos vía claves, sin strings hardcodeados.
  - Verify: test de Vitest que renderiza una sección en ambos idiomas.
  - Files: `frontend/src/i18n/{es,en}.json`, `frontend/src/i18n/index.ts`, `frontend/src/main.ts`.

- [x] **Task: Esqueleto de 6 secciones**
  - Acceptance: el panel muestra las secciones Resumen, Servicio, Acceso web, Consumo, Config, Ajustes como contenedores vacíos (rutas/estado listos).
  - Verify: test de Vitest que monta el shell y verifica las 6 secciones.
  - Files: `frontend/src/App.vue`, `frontend/src/stores/`, `frontend/src/components/` (placeholders).

---

## Fase 2 — service-control

- [x] **Task: Wrapper de exec + Status/Info**
  - Acceptance: `run()` ejecuta el CLI con timeout; `Status()` y `Info()` parsean `opencode service status` y el JSON de `/api/info` (incluye `pid`).
  - Verify: `go test` con runner mockeado (sin `opencode` real).
  - Files: `internal/servicecontrol/exec.go`, `internal/servicecontrol/status.go`, `internal/servicecontrol/service.go`, `*_test.go`.

- [x] **Task: Start/Stop/Restart**
  - Acceptance: `Start/Stop/Restart` invocan los comandos correctos y propagan errores.
  - Verify: `go test` table-driven con runner mockeado.
  - Files: `internal/servicecontrol/lifecycle.go`, `*_test.go`.

- [x] **Task: ForceStop (workaround PID)**
  - Acceptance: `ForceStop()` = `Stop()` → verificar → si sigue vivo, matar `pid` de `/api/info` (`taskkill /F` en Windows).
  - Verify: `go test` (Stop falla → mata; ya detenido → no mata) + prueba manual con el bug real.
  - Files: `internal/servicecontrol/lifecycle.go`, `internal/servicecontrol/kill_windows.go`, `*_test.go`.

- [x] **Task: Config Get/Set/Unset**
  - Acceptance: `Get/Set/Unset` leen y escriben hostname, port y password del servicio vía `opencode service set/get/unset`.
  - Verify: `go test` de argumentos formados + verificación manual de `service get hostname`.
  - Files: `internal/servicecontrol/config.go`, `*_test.go`.

---

## Fase 3a — usage

- [x] **Task: Fetch + parseo del consumo Go**
  - Acceptance: `Fetch()` consulta `GET https://opencode.ai/zen/go/v1/usage` y parsea rolling/weekly/monthly (percent + resetsAt), tolerando `null`.
  - Verify: `go test` con `httptest.Server` (respuesta válida, `percent:null`, 401, 403/EntitlementError).
  - Files: `internal/usage/fetch.go`, `internal/usage/usage.go`, `*_test.go`.

- [x] **Task: Resolución de token**
  - Acceptance: manual (Ajustes) > auto desde `~/.local/share/opencode/auth.json` (`opencode-go.key`) > "sin token".
  - Verify: `go test` con auth.json temporal (prioridad y ausencia).
  - Files: `internal/usage/token.go`, `*_test.go`.

- [x] **Task: Severity para el tray**
  - Acceptance: `Severity()` mapea consumo → verde/amarillo/rojo según las reglas (≥90 amarillo, ≥98 rojo, error rojo).
  - Verify: `go test` por banda.
  - Files: `internal/usage/severity.go`, `*_test.go`.

---

## Fase 3b — web-access

- [x] **Task: Aplicación de modos (local/red/remoto)**
  - Acceptance: `ApplyMode` forma el hostname correcto (127.0.0.1 / 0.0.0.0) o activa remoto (`pair --remote`), usando el contrato de `service-control`.
  - Verify: `go test` mockeando `servicecontrol.Service`.
  - Files: `internal/webaccess/modes.go`, `internal/webaccess/webaccess.go`, `*_test.go`.

- [x] **Task: URLs de acceso**
  - Acceptance: `ComputeURLs()` devuelve IP(s) LAN + puerto para el modo red.
  - Verify: `go test` con IPs inyectadas.
  - Files: `internal/webaccess/urls.go`, `*_test.go`.

- [x] **Task: Credenciales (usuario + contraseña)**
  - Acceptance: `Username()` = "opencode"; `CurrentPassword()` lee la contraseña; `ChangePassword()` la cambia y reinicia el servicio; UI con mostrar/ocultar.
  - Verify: `go test` (cambio de contraseña) + Vitest (toggle mostrar/ocultar).
  - Files: `internal/webaccess/credentials.go`, `*_test.go`, `frontend/src/components/web-access/`.

---

## Fase 3c — config-viewer

- [x] **Task: Listado MCP/skills/agents**
  - Acceptance: `ListMCP/ListSkills/ListAgents` devuelven las listas (MCP con estado de conexión).
  - Verify: `go test` parseando salida de `opencode mcp list`.
  - Files: `internal/configviewer/mcp.go`, `internal/configviewer/catalog.go`, `*_test.go`.

- [x] **Task: Config JSON + enmascarado de secretos**
  - Acceptance: `ConfigJSON()` lee `~/.config/opencode/opencode.json(c)` y enmascara keys/tokens/passwords antes de exponerlos.
  - Verify: `go test` con `opencode.jsonc` de prueba (secretos → "••••").
  - Files: `internal/configviewer/config.go`, `internal/configviewer/mask.go`, `*_test.go`.

- [x] **Task: Abrir en editor / abrir carpeta**
  - Acceptance: `OpenInEditor` usa `code` si está en PATH, si no el editor por defecto; `OpenConfigDir` abre el directorio.
  - Verify: `go test` con PATH fakeado.
  - Files: `internal/configviewer/editor.go`, `*_test.go`.

---

## Fase 4 — notifications

- [ ] **Task: Alerta de ritmo (burn rate) + dedupe**
  - Acceptance: dispara en 20/40/60/80 % si `consumido − esperado ≥ 15`, solo weekly/monthly, una vez por hito/período.
  - Verify: `go test` (lógica pura: dispara/no dispara, dedupe, reset tras `resetsAt`).
  - Files: `internal/notifications/burnrate.go`, `internal/notifications/dedupe.go`, `*_test.go`.

- [ ] **Task: Alertas de agotamiento (90/98/100)**
  - Acceptance: 90 % naranja, 98 % rojo, 100 % crítico, en las tres ventanas, con dedupe.
  - Verify: `go test` por umbral.
  - Files: `internal/notifications/thresholds.go`, `*_test.go`.

- [ ] **Task: Estado del servicio (detenido inesperadamente)**
  - Acceptance: notifica solo cuando pasa de corriendo → detenido sin acción del usuario.
  - Verify: `go test` con estados mockeados.
  - Files: `internal/notifications/servicewatch.go`, `*_test.go`.

- [ ] **Task: Envío de toasts + ajustes**
  - Acceptance: `Notify` envía vía `beeep`; tipos y umbrales configurables desde Ajustes.
  - Verify: `go test` con sender mockeado + Vitest del panel de Ajustes.
  - Files: `internal/notifications/notify.go`, `internal/notifications/notifications.go`, `*_test.go`, `frontend/src/components/notifications/`.

---

## Fase 5 — Pulido + release

- [ ] **Task: Sistema de diseño (ui-ux-pro-max)**
  - Acceptance: `design-system/opencode-tray/MASTER.md` generado y aplicado (estilo, paleta, tipografía, espaciado); checklist de accesibilidad (contraste 4.5:1, foco visible, targets ≥44px) cumplido.
  - Verify: revisión visual + checklist `pro-rules.md` de la skill.
  - Files: `design-system/opencode-tray/MASTER.md`, `frontend/src/**`.

- [ ] **Task: README final + pasada de consistencia**
  - Acceptance: README documenta comandos, estructura y uso; specs/plan/todo reflejan el estado real (sin secciones huérfanas).
  - Verify: revisión manual cruzando README ↔ specs ↔ `tasks/`.
  - Files: `README.md`, `specs/SPEC-*.md`, `tasks/plan.md`, `tasks/todo.md`.

- [ ] **Task: Build portable + release**
  - Acceptance: `wails build` genera un `.exe` portable único y funcional.
  - Verify: ejecutar el `.exe` en Windows limpio.
  - Files: `build/bin/*.exe`, notas de release.
