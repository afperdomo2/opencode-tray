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
