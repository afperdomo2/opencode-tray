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
  - **Hecho:** `ListMCP` ignora los marcadores Unicode de estado (`✓`, `✗`, `●`, `○`) para que el nombre nunca se mezcle con el estado.
  - Acceptance: `ListMCP/ListSkills/ListAgents` devuelven las listas (MCP con estado de conexión).
  - Verify: `go test` parseando salida de `opencode mcp list` (incluye `○ github disabled`).
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

- [x] **Task: Alerta de ritmo (burn rate) + dedupe**
  - Acceptance: dispara en 20/40/60/80 % si `consumido − esperado ≥ 15`, solo weekly/monthly, una vez por hito/período.
  - Verify: `go test` (lógica pura: dispara/no dispara, dedupe, reset tras `resetsAt`).
  - Files: `internal/notifications/burnrate.go`, `internal/notifications/dedupe.go`, `*_test.go`.

- [x] **Task: Alertas de agotamiento (90/98/100)**
  - Acceptance: 90 % naranja, 98 % rojo, 100 % crítico, en las tres ventanas, con dedupe.
  - Verify: `go test` por umbral.
  - Files: `internal/notifications/thresholds.go`, `*_test.go`.

- [x] **Task: Estado del servicio (detenido inesperadamente)**
  - Acceptance: notifica solo cuando pasa de corriendo → detenido sin acción del usuario.
  - Verify: `go test` con estados mockeados.
  - Files: `internal/notifications/servicewatch.go`, `*_test.go`.

- [x] **Task: Envío de toasts + ajustes**
  - **Estado post-Fase 4:** lógica y panel de ajustes listos (tests verdes), pero el **envío está deshabilitado en runtime** — `app.go` no inyecta sender (queda el no-op). Motivo: `wailsruntime.SendNotification` (go-toast) bloqueaba la bandeja. *Pendiente reactivar por una vía segura (ver `SPEC-notifications.md` → Estado).*
  - Acceptance: `Notify` envía vía `beeep`; tipos y umbrales configurables desde Ajustes.
  - Verify: `go test` con sender mockeado + Vitest del panel de Ajustes.
  - Files: `internal/notifications/notify.go`, `internal/notifications/notifications.go`, `*_test.go`, `frontend/src/components/notifications/`.

---

## Fase 5 — Pulido + release

- [x] **Task: Sistema de diseño (ui-ux-pro-max)**
  - **Hecho:** `design-system/opencode-tray/MASTER.md` generado (Dark Mode OLED) y **aplicado** en `frontend/src/style.css` (+ `Ajustes.vue`): tokens `--color-*`/`--space-*`/tipografía (IBM Plex Sans + JetBrains Mono, con fallback a fuentes del sistema). A11y verificado por script WCAG: contraste texto ≥ 5.65:1 (todas las parejas PASS), foco visible (`:focus-visible`), targets ≥ 44px (nav, acciones, toggles, select, inputs; el `.mini` inline queda exento por WCAG 2.5.8). `prefers-reduced-motion` respetado. El "Page Pattern" (FAQ Landing) de MASTER no aplica a esta app y no se aplicó.
  - Acceptance: `design-system/opencode-tray/MASTER.md` generado y aplicado (estilo, paleta, tipografía, espaciado); checklist de accesibilidad (contraste 4.5:1, foco visible, targets ≥44px) cumplido.
  - Verify: revisión visual + checklist `pro-rules.md` de la skill (aplicado el subconjunto web/desktop; contrastes verificados por script).
  - Files: `design-system/opencode-tray/MASTER.md`, `frontend/src/**`.

- [x] **Task: README final + pasada de consistencia**
  - **Hecho:** README reescrito con estado real (Fase 5), comandos, funcionalidad, estructura (incluye `design-system/`), requisitos y limitaciones (notificaciones deshabilitadas, fuentes web, quirk de `wails dev`). `plan.md` corregido (design system generado/aplicado en Fase 5). `SPEC-notifications.md` con nota de "Estado" (envío deshabilitado). `todo.md` sincronizado.
  - Acceptance: README documenta comandos, estructura y uso; specs/plan/todo reflejan el estado real (sin secciones huérfanas).
  - Verify: revisión manual cruzando README ↔ specs ↔ `tasks/`.
  - Files: `README.md`, `specs/SPEC-*.md`, `tasks/plan.md`, `tasks/todo.md`.

- [x] **Task: Build portable + release**
  - **Hecho:** `wails build` genera `build/bin/opencode-tray.exe` (portable único, ~11.5 MB) con el design system aplicado. Notas de release + checklist de validación en [`RELEASE.md`](../RELEASE.md). Verificación final: `gofmt` limpio, `go test ./...` (6 módulos ok), `pnpm test` (6 tests), `go build` OK.
  - Acceptance: `wails build` genera un `.exe` portable único y funcional.
  - Verify: ejecutar el `.exe` en Windows limpio (checklist en `RELEASE.md`).
  - Files: `build/bin/*.exe`, `RELEASE.md` (notas de release).

---

## Fase 6 — resumen (overview + UI/UX)

> Spec: [`specs/SPEC-resumen.md`](../specs/SPEC-resumen.md). Frontend, sin backend nuevo. Orden: Consumo → MCP/Config → Servicio.

- [x] **Task: Vista Resumen (dashboard)**
  - **Hecho:** `Resumen.vue` compone `ServiceStatus` + `UsageFetch` + `ConfigListMCP`/`ConfigJSON`; bloques Consumo → MCP/Config → Servicio; es la vista inicial.
  - Acceptance: `Resumen.vue` compone `ServiceStatus` + `UsageFetch` + `ConfigListMCP`/`ConfigJSON`; bloques en orden Consumo → MCP/Config → Servicio; es la vista inicial; estados cargando/error.
  - Verify: Vitest (servicio on/off, consumo ok/warn/crit, cargando, error) — en verde.
  - Files: `frontend/src/components/Resumen.vue`, `resumen.spec.ts`, `App.vue`.

- [x] **Task: Iconos SVG inline**
  - **Hecho:** `Icon.vue` con set outline (nav, acciones, ojo, editor/carpeta, chevron, idioma, campana); aplicados en nav/botones; decorativos `aria-hidden`, botón solo-icono con `aria-label`; **sin emoji**.
  - Acceptance: componente `Icon.vue` con set outline (nav, acciones, ojo, editor/carpeta, chevron, idioma); aplicados donde aportan; decorativos con `aria-hidden`, botones solo-icono con `aria-label`; **sin emoji**.
  - Verify: Vitest (`Icon.vue` renderiza por nombre) — en verde.
  - Files: `frontend/src/components/Icon.vue`, `*.vue`, `style.css`.

- [x] **Task: Cabecera + idioma segmentado**
  - **Hecho:** eliminado el título "OpenCode Tray" de la cabecera; `LanguageSwitch.vue` (segmentado ES|EN con icono) reemplaza al `<select>`; `main.go` con el fondo del nuevo design system.
  - Acceptance: se **elimina** el título "OpenCode Tray" de la cabecera; el idioma pasa a `LanguageSwitch.vue` (control segmentado ES|EN, con icono).
  - Verify: Vitest (cambia el locale es↔en) — en verde.
  - Files: `frontend/src/App.vue`, `frontend/src/components/LanguageSwitch.vue`, `language-switch.spec.ts`, `i18n/*`.

- [x] **Task: Acordeón de config (+ profundidad visual)**
  - **Hecho:** `Accordion.vue` accesible (`aria-expanded`/`aria-controls`); JSON de config colapsado por defecto en Resumen y Config; cards + sombras + hover (tokens del design system). `Section.vue` (placeholder) eliminado por quedar huérfano.
  - Acceptance: componente `Accordion.vue` accesible (`aria-expanded`); el JSON de config se muestra **colapsado por defecto** en Resumen y en Config; se añade profundidad (cards, sombras, hover) usando tokens del design system.
  - Verify: Vitest (colapsado por defecto; abre/cierra; `aria-expanded`) — en verde; `wails build` OK.
  - Files: `frontend/src/components/Accordion.vue`, `accordion.spec.ts`, `Config.vue`, `style.css`.

---

## Fase 7 — ui-ux (refinamientos)

> Spec: [`specs/SPEC-ui-ux.md`](../specs/SPEC-ui-ux.md). Frontend, sin backend nuevo.

- [x] **Task: Quitar el menú Consumo**
  - **Hecho:** el nav ya no incluye Consumo; `Consumo.vue` eliminado (su info vive en Resumen).
  - Acceptance: la navegación no incluye Consumo; se elimina `Consumo.vue` y su ruta (los datos quedan en Resumen).
  - Verify: Vitest (App no renderiza sección `consumo`); build OK.
  - Files: `frontend/src/App.vue`, eliminar `frontend/src/components/Consumo.vue`.

- [x] **Task: Mover el idioma a Ajustes (y quitar la cabecera)**
  - **Hecho:** `LanguageSwitch` movido a `Ajustes.vue`; eliminada la cabecera de `App.vue` (sin título ni selector).
  - Acceptance: `LanguageSwitch` se mueve a `Ajustes.vue`; se elimina la cabecera vacía de `App.vue`.
  - Verify: Vitest (cambia locale) — en verde.
  - Files: `frontend/src/App.vue`, `frontend/src/components/Ajustes.vue`, `LanguageSwitch.vue`.

- [x] **Task: Resumen — botones sm (con texto) + config (3 botones + modal)**
  - **Hecho:** los 4 botones de servicio y los 3 de config usan la variante **`sm`** (`.actions.sm`: 30px, 12px, icono 14) **con texto e icono** (no solo-icono); JSON en **modal** accesible (también en la sección Config). Se eliminó el acordeón (Fase 6) y `Accordion.vue`.
  - Acceptance: botones de Servicio y Config del Resumen en tamaño **`sm`** con texto; bloque Config con **3 botones** (editor/carpeta/JSON); el JSON se abre en **modal** (también en la sección Config).
  - Verify: Vitest (4 botones `sm` con texto, 3 botones en config, abre modal) — en verde.
  - Files: `Resumen.vue`, `Config.vue`, `Modal.vue`, `modal.spec.ts`, `resumen.spec.ts`, `style.css`.

- [x] **Task: Ajustes — switches estilizados**
  - **Hecho:** los checkboxes de Ajustes usan `Switch.vue` (pill con knob), accesible.
  - Acceptance: los checkboxes de Ajustes usan `Switch.vue` (pill con knob), accesible.
  - Verify: Vitest (`Switch` refleja checked y emite change) — en verde.
  - Files: `Ajustes.vue`, `Switch.vue`, `switch.spec.ts`, `style.css`.

- [x] **Task: Resumen — color de consumo graduado**
  - **Hecho:** barras y `%` coloreados por proximidad al límite: verde `<70`, ámbar `70–84`, naranja `85–97`, rojo `≥98` (se mantiene el rojo en 98, coherente con `usage`). Contraste verificado por script (≥5.59:1).
  - Acceptance: barra y `%` cambian de color de forma graduada; el número `%` acompaña al color (no solo color).
  - Verify: Vitest (clases `warn`/`high`/`crit` en barra y `%`) — en verde.
  - Files: `Resumen.vue`, `style.css`, `resumen.spec.ts`.

- [x] **Task: Estandarizar el tamaño de botón `sm`**
  - **Hecho:** `.actions button` adopta el tamaño `sm` (30px / 12px / icono 14) como **estándar** en toda la app; eliminada la variante `.actions.sm`.
  - Acceptance: todos los botones de acción (Servicio, Acceso web, Ajustes) usan el tamaño `sm`.
  - Verify: `vue-tsc`/build + revisión visual.
  - Files: `style.css`, `Resumen.vue`, `Servicio.vue`, `AccesoWeb.vue`, `Ajustes.vue`.

- [x] **Task: Unificar los menús Servicio y Config**
  - **Hecho:** `Config.vue` fusionado en `Servicio.vue` (servicio + MCP + JSON en modal); nav = Resumen · Servicio · Acceso web · Ajustes; eliminado `nav.config`.
  - Acceptance: existe una sola sección `Servicio` con control + configuración; el nav no tiene `config`.
  - Verify: build OK; sin referencias huérfanas.
  - Files: `Servicio.vue`, eliminar `Config.vue`, `App.vue`, `i18n/*`.

- [x] **Task: Dashboard (Resumen) sin botones**
  - **Hecho:** el Resumen unifica **Servicio + Config en un widget** de solo lectura (estado + MCP) y **elimina los botones** de acción.
  - Acceptance: el Resumen no tiene botones; muestra widgets de Consumo y de Servicio+Config; las acciones están en los menús.
  - Verify: Vitest (no hay `.actions` ni `<button>` en Resumen) — en verde.
  - Files: `Resumen.vue`, `resumen.spec.ts`.

---

## Fase 8 — ui-ux (responsive)

> Spec: [`specs/SPEC-ui-ux.md`](../specs/SPEC-ui-ux.md) → Iteración 3. Frontend + ventana; sin dependencias nuevas.

- [x] **Task: Shell responsive y contenido centrado**
  - **Hecho:** `--content-max: 1120px`, `--content-gutter: clamp(...)`, nav y contenido con `max-width` + `margin-inline:auto`; `min-width:0` y wrap en textos largos.
  - Acceptance: en pantallas grandes el contenido no se estira a todo el ancho; en 340 px no hay scroll horizontal.
  - Verify: revisión manual 340/768/1280/2560 px + `pnpm build`.
  - Files: `style.css`, `App.vue`.

- [x] **Task: Grids responsive en las vistas**
  - **Hecho:** `.dashboard` (Resumen), `.panel-grid` (Servicio y Acceso web) y `.settings-grid` (Ajustes) con `auto-fit/minmax`; cards de consumo con grid interno y container query.
  - Acceptance: 1 columna en compacto y 2–3 columnas cuando el ancho lo permite; barras de consumo legibles.
  - Verify: Vitest (`.dashboard`, `.panel-grid`, `.settings-grid`) + revisión visual.
  - Files: `Resumen.vue`, `Servicio.vue`, `AccesoWeb.vue`, `Ajustes.vue`, `style.css`.

- [x] **Task: Botón estándar `.btn`**
  - **Hecho:** clase `.btn` única (30px / 12px / icono 14) aplicada a acciones; el botón **Guardar** de Acceso web deja de usar el estilo nativo y en su formulario iguala la altura del input (`min-height: var(--target)`).
  - Acceptance: todos los botones de acción comparten el mismo diseño; Guardar incluido.
  - Verify: Vitest (`acceso-web.spec.ts` busca `.pw-change button.btn`).
  - Files: `style.css`, `Servicio.vue`, `AccesoWeb.vue`, `Ajustes.vue`.

- [x] **Task: Tamaño mínimo de ventana**
  - **Hecho:** `MinWidth: 340`, `MinHeight: 420` en `main.go`.
  - Acceptance: la ventana no se puede reducir a un tamaño ilegible.
  - Verify: `go build ./...` + revisión manual.
  - Files: `main.go`.

---

## Fase 9 — ui-ux (skeletons de carga)

> Spec: [`specs/SPEC-ui-ux.md`](../specs/SPEC-ui-ux.md) → Iteración 4. Frontend; sin dependencias nuevas.

- [x] **Task: Componente Skeleton accesible**
  - **Hecho:** `Skeleton.vue` con `width`/`height`/`radius`, `aria-hidden`; contenedores con `role="status"` y texto `.sr-only` (`common.loading`); shimmer respeta `prefers-reduced-motion`.
  - Acceptance: skeletons decorativos + estado de carga anunciado; sin saltos de layout.
  - Verify: `skeleton.spec.ts` — en verde.
  - Files: `Skeleton.vue`, `skeleton.spec.ts`, `style.css`, `i18n/*`.

- [x] **Task: Carga independiente en Resumen**
  - **Hecho:** `loadUsage`, `loadService` y `loadMcp` en paralelo con loading/error propios; consumo se pinta sin esperar al servicio.
  - Acceptance: si consumo resuelve antes, se muestra mientras servicio/MCP siguen en skeleton; errores por card.
  - Verify: Vitest (consumo visible con servicio pendiente; skeleton desaparece al resolver) — en verde.
  - Files: `Resumen.vue`, `resumen.spec.ts`.

- [x] **Task: Skeletons en Servicio, Acceso web y Ajustes**
  - **Hecho:** `loadStatus`/`loadMcp`/`loadJson` separados en Servicio (MCP visible sin esperar al JSON; skeleton en el modal); `loadUrls`/`loadCredentials` en Acceso web (usuario y contraseña en paralelo); skeleton de General mientras `GetSettings`.
  - Acceptance: cada bloque muestra skeleton/error/datos de forma independiente.
  - Verify: `servicio.spec.ts`, `acceso-web.spec.ts`, `ajustes.spec.ts` — en verde.
  - Files: `Servicio.vue`, `AccesoWeb.vue`, `Ajustes.vue`, `*.spec.ts`.

---

## Fase 10 — token de OpenCode + caché de consumo

> Specs: `SPEC-usage.md`, `SPEC-shell.md`, `SPEC-resumen.md`, `SPEC-ui-ux.md` (iteración futura de token/caché).

- [x] **Task: Persistencia cifrada del token manual (DPAPI)**
  - Acceptance: el token manual se guarda cifrado en `settings.json`; `GetSettings` nunca lo expone; `SaveSettings` no lo borra.
  - Verify: `go test ./internal/shell/...` (round-trip DPAPI, blob inválido, sanitizado).
  - Files: `internal/shell/settings.go`, `internal/shell/secret_windows.go`, `internal/shell/secret_other.go`, `*_test.go`, `app.go`.

- [x] **Task: Lectura del token de OpenCode v2**
  - Acceptance: soporta `opencode` y `opencode-go` en `auth.json`; relee el archivo en cada consulta; `ReloadToken()` usa `opencode auth export` y cachea el token solo en memoria.
  - Verify: `go test ./internal/usage/...` (parser de export, runner mockeado, candidatos, 401/403).
  - Files: `internal/usage/token.go`, `internal/usage/export.go`, `internal/usage/exec_windows.go`, `internal/usage/fetch.go`, `*_test.go`.

- [x] **Task: Reintento y estado del token en Wails**
  - Acceptance: `UsageTokenStatus`, `UsageReloadToken`, `UsageSetManualToken` y `UsageClearManualToken` expuestos; errores con prefijos `USAGE_*`.
  - Verify: `go test ./...` + `wails build` regenera bindings.
  - Files: `app.go`, `internal/usage/*.go`, `frontend/wailsjs/**` (generado).

- [x] **Task: Caché nativo de consultas**
  - Acceptance: `useQuery` con caché en módulo, `staleTime`, dedupe, datos previos y skeleton solo inicial.
  - Verify: `pnpm test` (`query.spec.ts`).
  - Files: `frontend/src/lib/query.ts`, `query.spec.ts`.

- [x] **Task: Resumen con caché y reintento de token**
  - Acceptance: no muestra skeleton al remontar dentro de `staleTime`; conserva datos al refrescar; CTA de reintento/configuración sin token.
  - Verify: `resumen.spec.ts` en verde.
  - Files: `Resumen.vue`, `App.vue`, `resumen.spec.ts`.

- [x] **Task: Ajustes — token e intervalo**
  - Acceptance: card OpenCode Go con estado, reintento, API key manual write-only, quitar y selector de intervalo; todo vía i18n.
  - Verify: `ajustes.spec.ts` en verde; `pnpm build`.
  - Files: `Ajustes.vue`, `frontend/src/lib/settings.ts`, `frontend/src/lib/usage.ts`, `i18n/*`, `style.css`, `ajustes.spec.ts`.

---

## Fase 11 — flyout nativo anclado

> Specs: `SPEC-shell.md` y `SPEC-ui-ux.md` (Iteración 5). Backend + frontend; sin dependencias nuevas.

- [x] **Task: Geometría de anclaje + Win32**
  - **Hecho:** `flyout.go` calcula la posición para taskbar arriba/abajo/izquierda/derecha con clamping; `flyout_windows.go` posiciona en píxeles físicos, aplica `WS_EX_TOOLWINDOW` y foco.
  - Acceptance: calcula la posición del panel para taskbar arriba/abajo/izquierda/derecha, con clamping; posiciona en píxeles físicos; aplica `WS_EX_TOOLWINDOW` y foco.
  - Verify: `go test ./internal/shell/...` (4 bordes, multi-monitor, clamping).
  - Files: `internal/shell/flyout.go`, `flyout_windows.go`, `flyout_other.go`, `flyout_test.go`.

- [x] **Task: Controlador de flyout**
  - **Hecho:** `ShowAt`, `Hide`, `HideFromBlur`, `ToggleAt` y `SetHeight` con guarda anti-carrera y clamp 320–600 px.
  - Acceptance: `ShowAt`, `Hide`, `HideFromBlur`, `ToggleAt`, `SetHeight`; guarda anti-carrera blur/tray y clamp 320–600 px.
  - Verify: tests con reloj inyectado y backend mock.
  - Files: `internal/shell/window.go`, `window_test.go`.

- [x] **Task: Configuración de ventana y bindings**
  - **Hecho:** ventana frameless/fija/topmost con `WindowClassName`; tray distingue clic (toggle) de menú (mostrar); bindings `SetFlyoutHeight`, `HideWindow` y `WindowBlurred`.
  - Acceptance: ventana frameless/fija/topmost con `WindowClassName`; tray distingue clic (toggle) de menú (mostrar); bindings `SetFlyoutHeight`, `HideWindow` y `WindowBlurred`.
  - Verify: `go test ./...` + `wails build`.
  - Files: `main.go`, `app.go`, `internal/shell/tray.go`, `frontend/wailsjs/**` (generado).

- [x] **Task: Nav inferior y alto según contenido**
  - **Hecho:** nav al final del DOM como footer fijo; `ResizeObserver` con debounce solicita el alto; `Esc`/`blur` cierran; el modal no propaga `Esc`.
  - Acceptance: nav al final del DOM, footer fijo, contenido con scroll; `ResizeObserver` con debounce solicita el alto; `Esc`/`blur` cierran; el modal no propaga `Esc`.
  - Verify: `pnpm test` (`app.spec.ts` y `modal.spec.ts`).
  - Files: `App.vue`, `Modal.vue`, `style.css`, `app.spec.ts`, `modal.spec.ts`.

- [x] **Task: Sincronizar README y specs**
  - **Hecho:** README documenta el flyout, el footer de navegación, la ausencia de resize/arrastre y la validación con build de producción.
  - Acceptance: README describe el flyout, la ausencia de resize/arrastre y la validación con build de producción.
  - Verify: revisión cruzada README ↔ specs ↔ tasks.
  - Files: `README.md`, `specs/SPEC-shell.md`, `specs/SPEC-ui-ux.md`, `tasks/plan.md`, `tasks/todo.md`.

---

## Fase 12 — cierre estable y tooltip de consumo

> Specs: `SPEC-shell.md` (Iteración 6). Backend `shell`; sin dependencias nuevas.

- [x] **Task: Guarda de blur post-apertura**
  - **Hecho:** `ShowAt` registra apertura; `HideFromBlur` espera la guarda y comprueba foreground antes de cerrar; timers cancelados al mostrar/ocultar.
  - Acceptance: un blur transitorio al abrir no cierra el panel; un blur real sí lo cierra; no quedan timers obsoletos.
  - Verify: `go test ./internal/shell/...` con reloj/timers inyectados.
  - Files: `internal/shell/window.go`, `window_test.go`.

- [x] **Task: Detección y refuerzo de foco**
  - **Hecho:** `FlyoutForeground` compara `GetForegroundWindow` con el HWND; `FocusFlyout` refuerza la activación si hace falta.
  - Acceptance: el panel queda en primer plano tras el clic del tray y recibe teclado/`Esc`.
  - Verify: `go test ./...` + prueba manual con build de producción.
  - Files: `internal/shell/flyout_windows.go`, `flyout_other.go`, `app.go`.

- [x] **Task: Tooltip nativo con consumo**
  - **Hecho:** `UsageFetch` actualiza el tooltip tras una lectura correcta; formato `OpenCode · 5h 18% · 7d 42% · 30d 27%`; se conserva el anterior si falla.
  - Acceptance: el hover sobre el icono muestra los tres porcentajes; no se rompe si aún no hay datos.
  - Verify: `go test ./...` (formateo y actualización) + prueba manual.
  - Files: `app.go`, `app_test.go`, `internal/shell/tray.go`, `tray_test.go`.

- [x] **Task: Sincronizar docs y validar**
  - **Hecho:** spec Iteración 6, plan/todo Fase 12 y README sincronizados; suite completa y build de producción.
  - Acceptance: documentación coherente; tests y `wails build` en verde.
  - Verify: `go test ./...`, `pnpm test`, `pnpm build`, `wails build`.
  - Files: `README.md`, `specs/SPEC-shell.md`, `tasks/plan.md`, `tasks/todo.md`.

---

## Fase 13 — pestañas estables, acciones compactas y scroll sin saltos

> Specs: `SPEC-ui-ux.md` (Iteración 7). Solo frontend; sin dependencias nuevas.

- [x] **Task: Pestañas con métricas fijas**
  - **Hecho:** altura fija de 52 px, peso constante y slot de icono fijo.
  - Acceptance: ninguna pestaña cambia de tamaño ni posición al activarse o cambiar de idioma.
  - Verify: revisión manual a 380 px + `pnpm build`.
  - Files: `frontend/src/style.css`.

- [x] **Task: Estado y URI juntos en Servicio**
  - **Hecho:** la URI se mueve a `.status-row`, alineada a la derecha del estado; el backend descarta `stopped` como URL.
  - Acceptance: estado y URI comparten fila cuando caben; un servicio detenido no muestra `stopped` como URI.
  - Verify: `servicio.spec.ts`, `servicecontrol/service_test.go`.
  - Files: `frontend/src/components/Servicio.vue`, `frontend/src/style.css`, `internal/servicecontrol/status.go`.

- [x] **Task: Etiquetas cortas y grids de acciones**
  - **Hecho:** servicio en 2×2 (`Iniciar`, `Detener`, `Forzar`, `Refrescar`) y configuración en 3 columnas (`Editor`, `Carpeta`, `JSON`), sin salto de texto.
  - Acceptance: cada etiqueta queda en una línea y las filas son estables en ES/EN.
  - Verify: `servicio.spec.ts`.
  - Files: `Servicio.vue`, `frontend/src/i18n/es.json`, `frontend/src/i18n/en.json`, `frontend/src/style.css`.

- [x] **Task: Tooltips en botones icono**
  - **Hecho:** `title` en botones cortos, cerrar modal y mostrar/ocultar contraseña; `aria-label` se mantiene.
  - Acceptance: todo botón solo-icono muestra tooltip.
  - Verify: `modal.spec.ts`, `acceso-web.spec.ts`.
  - Files: `Modal.vue`, `AccesoWeb.vue`, `Servicio.vue`.

- [x] **Task: Scroll sin reflujo**
  - **Hecho:** `scrollbar-gutter: stable`, barra fina y `overflow-x: hidden`.
  - Acceptance: el ancho del contenido no cambia al aparecer el scroll vertical.
  - Verify: revisión manual con secciones cortas y largas.
  - Files: `frontend/src/style.css`.

- [x] **Task: Sincronizar docs y validar**
  - **Hecho:** spec, plan, todo, design-system y README sincronizados; suite y build completos.
  - Acceptance: documentación coherente; tests y `wails build` en verde.
  - Verify: `pnpm test`, `pnpm build`, `go test ./...`, `wails build`.
  - Files: `README.md`, `design-system/opencode-tray/MASTER.md`, `specs/SPEC-ui-ux.md`, `tasks/plan.md`, `tasks/todo.md`.

---

## Fase 14 — alto fijo y fila de servicio unificada

> Specs: `SPEC-shell.md` (Iteración 7) y `SPEC-ui-ux.md` (Iteración 8). Sin dependencias nuevas.

- [x] **Task: Alto fijo del flyout**
  - **Hecho:** backend inicializa en `FlyoutMaxHeight` (600 px); se elimina la medición de contenido y el `ResizeObserver` del frontend. Verificado en producción: 475×750 físicos a 125 % (380×600 lógicos).
  - Acceptance: las 4 pestañas mantienen el mismo alto; solo el contenido hace scroll.
  - Verify: `window_test.go`, `app.spec.ts`, revisión manual.
  - Files: `main.go`, `internal/shell/window.go`, `internal/shell/window_test.go`, `frontend/src/App.vue`, `frontend/src/app.spec.ts`.

- [x] **Task: Estado y URI en Resumen**
  - **Hecho:** la URI se mueve dentro de `.status-row`, igual que en Servicio; test dedicado de estructura.
  - Acceptance: estado a la izquierda y URI con puerto a la derecha en la misma fila.
  - Verify: `resumen.spec.ts` + revisión visual.
  - Files: `frontend/src/components/Resumen.vue`, `frontend/src/components/resumen.spec.ts`.

- [x] **Task: Sincronizar docs y validar**
  - **Hecho:** specs, plan, todo, design-system y README sincronizados; suite completa y build de producción en verde.
  - Acceptance: documentación coherente; tests y `wails build` en verde.
  - Verify: `pnpm test`, `pnpm build`, `go test ./...`, `wails build`.
  - Files: `README.md`, `design-system/opencode-tray/MASTER.md`, `specs/SPEC-shell.md`, `specs/SPEC-ui-ux.md`, `tasks/plan.md`, `tasks/todo.md`.

---

## Fase 15 — tooltip mensual, URLs accionables e instalador

> Specs: `SPEC-shell.md` (Iteración 8), `SPEC-usage.md` (Iteración 8) y `SPEC-web-access.md` (Iteración 2). Sin dependencias nuevas.

- [x] **Task: Humanizador de reinicio + barra de consumo**
  - **Hecho:** `internal/usage/humanize.go` con `HumanReset`, `BatteryBar`, `TruncateUTF16` y `Tooltip`; español/inglés, zona local, 127 UTF-16.
  - Acceptance: hoy → relativo, mañana → `mañana a las HH:MM`, 2+ días → fecha; sin reset → `sin reinicio`.
  - Verify: `internal/usage/humanize_test.go`.
  - Files: `internal/usage/humanize.go`, `internal/usage/humanize_test.go`.

- [x] **Task: Tooltip vivo sin abrir el panel**
  - **Hecho:** watcher en `App` con fetch inicial, ticker de 60 s, caché compartida (60 s), reintento de token al arrancar y refresco al cambiar idioma.
  - Acceptance: el tooltip muestra datos desde el arranque y actualiza la cuenta atrás; un error conserva el último dato.
  - Verify: `app_test.go`, `go test ./...` + revisión manual.
  - Files: `app.go`, `app_test.go`, `internal/shell/singleinstance_bindings.go`, `internal/shell/singleinstance_windows.go`.

- [x] **Task: URLs reales y acciones de abrir/copiar**
  - **Hecho:** `webaccess.State` + `URLItem`; fuente `/api/info` (local/LAN/túnel) con fallback local; `IsAccessURL` valida antes de abrir; UI con badges, abrir, copiar y feedback.
  - Acceptance: puerto real (45278 en Windows), URL del túnel visible, acciones funcionan y se deshabilitan si el servicio está detenido.
  - Verify: `webaccess_test.go`, `acceso-web.spec.ts`, prueba live contra OpenCode v2.0.32.
  - Files: `internal/webaccess/{webaccess,modes,urls}.go`, `internal/webaccess/webaccess_test.go`, `app.go`, `frontend/src/components/AccesoWeb.vue`, `frontend/src/components/acceso-web.spec.ts`.

- [x] **Task: Modos exclusivos y confirmación de remoto**
  - **Hecho:** `ApplyMode(local|network)` apaga `remote`; `ApplyMode(remote)` fija loopback + `pair --remote --yes`; modal de confirmación y avisos de seguridad.
  - Acceptance: cambiar a local/red no deja OpenTunnel encendido; activar remoto no ocurre con un solo clic.
  - Verify: `webaccess_test.go`, `acceso-web.spec.ts`.
  - Files: `internal/webaccess/modes.go`, `frontend/src/components/AccesoWeb.vue`, i18n.

- [x] **Task: Skeleton de 2 filas y tooltips de modo**
  - **Hecho:** 2 filas esqueleto con badge/URL/acciones; `title` + `aria-label` + descripción visible del modo activo.
  - Acceptance: sin salto de layout y modos comprensibles sin hover.
  - Verify: `acceso-web.spec.ts`.
  - Files: `frontend/src/components/AccesoWeb.vue`, `frontend/src/style.css`, `frontend/src/i18n/{es,en}.json`.

- [x] **Task: Instalador Inno Setup per-user**
  - **Hecho:** `.iss` con instalación en `%LOCALAPPDATA%`, detección/descarga de WebView2, aviso de `opencode`, accesos directos y desinstalador; `wails.json` con `Info`; script `build-installer.ps1`.
  - Acceptance: genera `OpenCodeTray-Setup-<versión>.exe` sin admin; conserva ajustes al desinstalar.
  - Verify: compilar en Windows limpio; `iscc` + `wails build`.
  - Files: `build/windows/installer/opencode-tray.iss`, `scripts/build-installer.ps1`, `wails.json`.

- [x] **Task: Sincronizar docs y validar**
  - **Hecho:** README, RELEASE, specs, plan y todo actualizados; `go test ./...`, `pnpm test`, `pnpm build` y `wails build` en verde.
  - Acceptance: documentación coherente; build de producción disponible para revisión.
  - Verify: suite y build completos.
  - Files: `README.md`, `RELEASE.md`, `specs/SPEC-{shell,usage,web-access}.md`, `tasks/plan.md`, `tasks/todo.md`.

---

## Fase 16 — fix del clic y menú del tray

> Spec: `SPEC-shell.md` (Iteración 9). Sin dependencias nuevas.

- [x] **Task: Registrar el tray en el hilo principal**
  - **Hecho:** `main.go` llama a `app.initTray(trayIcon)` antes de `wails.Run`; `startup` ya no crea el tray. La ventana `SystrayClass` queda en el hilo que ejecuta el bucle de mensajes de Wails.
  - Acceptance: clic izquierdo abre/cierra el panel; clic derecho abre el menú; "Salir" cierra la app.
  - Verify: clic simulado con `PostMessage` + `IsWindowVisible` del flyout y de `#32768`; prueba manual.
  - Files: `main.go`, `app.go`, `internal/shell/tray.go`, `AGENTS.md`.

- [x] **Task: Documentar el gotcha**
  - **Hecho:** comentario en `NewTrayHandler` y nota en `AGENTS.md` → "Notas técnicas".
  - Acceptance: evitar que el tray vuelva a registrarse dentro de `startup`.
  - Verify: revisión de documentación + build.
  - Files: `internal/shell/tray.go`, `AGENTS.md`, `specs/SPEC-shell.md`, `tasks/plan.md`, `tasks/todo.md`.

---

## Fase 17 — dashboard de widgets en Resumen

> Spec: `SPEC-resumen.md` (Iteración 2). Sin dependencias nuevas.

- [x] **Task: Shell de widget y layout bento**
  - **Hecho:** `ResumenWidget.vue` con cabecera, icono, tono, chevron y clic/teclado; grid de 2 columnas (consumo y MCP a todo el ancho).
  - Acceptance: 4 widgets con jerarquía visual y adaptación a una columna en ventanas estrechas.
  - Verify: `resumen.spec.ts` + captura del flyout.
  - Files: `frontend/src/components/ResumenWidget.vue`, `frontend/src/components/Resumen.vue`, `frontend/src/style.css`.

- [x] **Task: Widgets de consumo, servicio, acceso y MCP**
  - **Hecho:** consumo con las tres barras actuales; servicio con tinte verde/neutro; acceso con icono y color por modo + chip Internet; MCP con chips/dots sin texto de estado.
  - Acceptance: información escaneable, color solo como refuerzo y URL de servicio sin esquema para que quepa.
  - Verify: tests de severidad, modo remoto, chips MCP y skeletons independientes.
  - Files: `frontend/src/components/Resumen{Consumo,Servicio,Acceso,Mcp}.vue`, i18n es/en.

- [x] **Task: Navegación por clic/teclado**
  - **Hecho:** Servicio → Servicio, Acceso → Acceso web, MCP → Servicio; `role="link"`, Enter/Space, foco visible y hover.
  - Acceptance: los tres widgets navegables funcionan con ratón y teclado; Consumo no navega.
  - Verify: `resumen.spec.ts` (click y keydown).
  - Files: `ResumenWidget.vue`, `Resumen.vue`, tests.

- [x] **Task: Documentar y validar**
  - **Hecho:** README, SPEC-resumen, CAPABILITY-MAP, plan y todo sincronizados; `pnpm test` (60 tests), `pnpm build`, `go test ./...` y `wails build`.
  - Acceptance: documentación y build de producción coherentes.
  - Verify: suite completa + captura visual.
  - Files: `README.md`, `specs/SPEC-resumen.md`, `specs/CAPABILITY-MAP.md`, `tasks/plan.md`, `tasks/todo.md`.

---

## Fase 18 — frescura de datos, skeletons y latencia

> Specs: `SPEC-resumen.md` (Iteración 3), `SPEC-web-access.md` (Iteración 3), `SPEC-service-control.md` (Snapshot) y `SPEC-ui-ux.md` (Iteración 3). Sin dependencias nuevas.

- [x] **Task: Caché compartido y dos tipos de refresco**
  - **Hecho:** `query.ts` separa `refetch()` (suave: conserva datos) de `refetchHard()` (duro: descarta y muestra skeleton). `clearQueries`/`invalidateQueries` permiten invalidar por lote. Claves unificadas en `lib/query-keys.ts`.
  - Acceptance: volver a una pestaña con datos frescos no hace petición ni pinta skeleton; un refresco duro sí muestra skeleton.
  - Verify: `query.spec.ts`.
  - Files: `frontend/src/lib/query.ts`, `frontend/src/lib/query-keys.ts`, `frontend/src/lib/query.spec.ts`.

- [x] **Task: Servicio con caché y skeletons correctos**
  - **Hecho:** `Servicio.vue` usa `useQuery` (estado 10 s, MCP 60 s); JSON diferido al abrir el modal; refresco manual con skeleton; mutaciones descartan estado desde el inicio.
  - Acceptance: primera carga y refresco manual/mutación con skeleton; revalidaciones silenciosas conservan datos; error con datos no borra el contenido.
  - Verify: `servicio.spec.ts`.
  - Files: `frontend/src/components/Servicio.vue`, `frontend/src/components/servicio.spec.ts`, i18n es/en.

- [x] **Task: Acceso web con caché, refresco y mutaciones**
  - **Hecho:** `AccesoWeb.vue` usa `useQuery` (estado 15 s, credenciales 60 s); botón de refresco en la cabecera; cambio de modo/contraseña con skeleton desde el inicio y conservación del modo activo.
  - Acceptance: volver a la pestaña no repite petición ni skeleton; refrescar y mutar sí muestran skeleton; el botón de modo no salta.
  - Verify: `acceso-web.spec.ts`.
  - Files: `frontend/src/components/AccesoWeb.vue`, `frontend/src/components/acceso-web.spec.ts`, `frontend/src/style.css`, i18n es/en.

- [x] **Task: Resumen sincronizado con las acciones**
  - **Hecho:** claves compartidas, polling silencioso del servicio cada 15 s, aviso de “servicio detenido” derivado del estado compartido y sin botón de refresco. Se eliminó el texto “Actualizando…” de Consumo.
  - Acceptance: tras detener el servicio en Servicio, al volver a Resumen el widget muestra “Detenido” (skeleton mientras llega el estado); el polling no pinta skeleton.
  - Verify: `resumen.spec.ts`.
  - Files: `frontend/src/components/Resumen.vue`, `frontend/src/components/ResumenConsumo.vue`, i18n es/en.

- [x] **Task: Backend — una lectura de estado y una de configuración**
  - **Hecho:** `servicecontrol.Snapshot()` lee `opencode service get` una vez; `webaccess.State` reutiliza un solo `service status` y solo consulta `/api/info` en red/remoto. `App` cachea el estado 1 s con singleflight e invalida en start/stop/force-stop/modo/contraseña.
  - Acceptance: `WebAccessState` deja de lanzar 6 procesos; Resumen no duplica `service status` en llamadas paralelas.
  - Verify: `go test ./...` (conteo de llamadas y TTL).
  - Files: `internal/servicecontrol/config.go`, `internal/servicecontrol/service_test.go`, `internal/webaccess/urls.go`, `internal/webaccess/webaccess.go`, `internal/webaccess/webaccess_test.go`, `app.go`, `app_test.go`.

- [x] **Task: Documentar y validar**
  - **Hecho:** README, specs, `CAPABILITY-MAP`, plan y todo sincronizados; `go test ./...`, `pnpm test`, `pnpm build` y `wails build` en verde.
  - Acceptance: documentación y build de producción coherentes.
  - Verify: suites + build completo.
  - Files: `README.md`, `specs/*.md`, `tasks/plan.md`, `tasks/todo.md`.

---

## Fase 19 — Linux/Ubuntu

> Spec: `SPEC-shell.md` (Iteración 10). Sin dependencias Go nuevas.

- [x] **Task: Toolchain de Linux**
  - **Hecho:** documentados `build-essential`, `pkg-config`, `libgtk-3-dev` y `libwebkit2gtk-4.1-dev`; Wails CLI v2.16 y el flag `-tags webkit2_41` (Ubuntu 26.04 solo trae WebKitGTK 4.1); `scripts/build-linux.sh` encapsula el build.
  - Acceptance: `wails build -tags webkit2_41` produce `build/bin/opencode-tray` en Ubuntu.
  - Verify: build real + arranque del binario.
  - Files: `README.md`, `scripts/build-linux.sh`, `AGENTS.md`.

- [x] **Task: Icono de bandeja por SO**
  - **Hecho:** `main.go` embebe `build/windows/icon.ico` y `build/linux/icon.png` (64 px) y `trayIcon()` elige por `runtime.GOOS`; `energye/systray` usa `image.Decode`, que no lee ICO.
  - Acceptance: el icono aparece en GNOME/AppIndicator sin errores de decodificación.
  - Verify: log de la app + bandeja.
  - Files: `main.go`, `build/linux/icon.png`.

- [x] **Task: Auto-inicio XDG**
  - **Hecho:** `internal/shell/autostart_xdg.go` implementa `runKey` sobre `~/.config/autostart/opencode-tray.desktop`; `autostart_other.go` queda como no-op solo para el resto de SO; textos i18n neutros ("Iniciar con el sistema" / "Start with system").
  - Acceptance: activar crea la entrada `.desktop` con `Exec` entrecomillado; desactivar la borra.
  - Verify: `autostart_xdg_test.go` (round-trip, borrado y comando vacío).
  - Files: `internal/shell/autostart_xdg.go`, `internal/shell/autostart_xdg_test.go`, `internal/shell/autostart_other.go`, `frontend/src/i18n/{es,en}.json`, `frontend/src/components/ajustes.spec.ts`.

- [x] **Task: Posición del panel sin Win32**
  - **Hecho:** `app.go` delega en `positionFlyout` (`placement_windows.go` → Win32; `placement_other.go` → `ScreenGetAll` + `WindowSetPosition` con `shell.CornerPosition`).
  - Acceptance: en X11/XWayland el panel aparece en la esquina inferior derecha; en Wayland el compositor decide sin romper nada.
  - Verify: `TestCornerPosition` y prueba manual.
  - Files: `placement_windows.go`, `placement_other.go`, `app.go`, `internal/shell/flyout.go`, `internal/shell/flyout_test.go`.

- [x] **Task: Arranque oculto en Linux**
  - **Hecho:** Wails v2.16 ignora `StartHidden` en Linux (`Hide` + `gtk_widget_show_all`); `App.startup` oculta la ventana en no-Windows para arrancar solo con el icono.
  - Acceptance: al lanzar el binario no aparece el panel; se abre desde "Mostrar panel".
  - Verify: `xwininfo` → `Map State: IsUnMapped` al arrancar y `IsViewable` tras el ítem de menú.
  - Files: `app.go`, `AGENTS.md`, `README.md`, `specs/SPEC-shell.md`.

- [x] **Task: README reordenado por plataforma**
  - **Hecho:** README con TOC y anclas, "Inicio rápido" (Ubuntu en 4 comandos / Windows en 2), sección "Probar la app" con checklist de humo y tabla de problemas frecuentes, builds separados por SO y limitaciones unificadas; se documenta el `export PATH="$HOME/go/bin"` que faltaba.
  - Acceptance: quien clona encuentra cómo compilar y ejecutar en Ubuntu en la primera pantalla, sin mezclar pasos de Windows.
  - Verify: enlaces relativos y anclas del TOC validados; `AGENTS.md` sin referencias a secciones viejas.
  - Files: `README.md`, `AGENTS.md`, `tasks/todo.md`.

- [x] **Task: Documentar limitaciones y validar**
  - **Hecho:** README (requisitos, build y limitaciones de Linux), `AGENTS.md`, `SPEC-shell.md` (Iteración 10), `CAPABILITY-MAP`, plan y todo sincronizados.
  - Acceptance: documentación coherente y build verde en Linux.
  - Verify: `go test ./...`, `pnpm test`, `pnpm build`, `wails build -tags webkit2_41`.
  - Files: `README.md`, `AGENTS.md`, `specs/SPEC-shell.md`, `specs/CAPABILITY-MAP.md`, `tasks/plan.md`, `tasks/todo.md`.
