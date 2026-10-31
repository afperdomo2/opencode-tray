# Plan técnico — OpenCode Tray

> Phase 2 del flujo SDD. Fuente de verdad: `specs/CAPABILITY-MAP.md` + las 6 `specs/SPEC-*.md` aprobadas.
> Tras aprobar este plan, sigue Phase 3 (`tasks/todo.md`).

## 1. Resumen

App de bandeja (tray) para Windows (y Linux para desarrollo y prueba) con Wails v2 (Go) + Vue 3/Vite/TS que controla el servicio de OpenCode, el acceso web (local/red/remoto), muestra el consumo de OpenCode Go, permite ver la config (MCP/skills/agents) y emite notificaciones nativas. Se construye **por módulos, en orden de dependencias, con tests desde el inicio**.

## 2. Componentes y dependencias

```
shell
 └── service-control
       ├── web-access
       ├── usage          (paralelo con web-access y config-viewer)
       └── config-viewer
             ├── notifications   (lee usage + service-control)
             └── resumen         (frontend: compone usage + config-viewer + service-control)
```

- **Secuencial**: `shell` → `service-control` → resto.
- **Paralelizable** (tras `service-control`): `web-access`, `usage`, `config-viewer`.
- **Final**: `notifications` y `resumen` (frontend).

## 3. Orden de implementación (fases)

| Fase | Módulo | Qué entrega |
|---|---|---|
| 0 | Scaffold | `wails init -n opencode-tray -t vue-ts`, estructura, CI local de tests |
| 1 | `shell` | tray + panel popover + single-instance + ajustes + auto-inicio + i18n + esqueleto de secciones |
| 2 | `service-control` | status/start/stop/restart + `ForceStop` (workaround PID) + config Get/Set/Unset |
| 3a | `usage` | fetch consumo Go + token + `Severity` |
| 3b | `web-access` | modos local/red/remoto + usuario/contraseña |
| 3c | `config-viewer` | visor MCP/skills/agents + config JSON + abrir en editor |
| 4 | `notifications` | ritmo + agotamiento + servicio detenido |
| 5 | Pulido + release | README final, design system aplicado, `.exe` portable, instalador (posterior) |
| 6 | `resumen` | vista overview (landing) + ajustes de UI/UX (iconos, cabecera, idioma segmentado, acordeón de config, profundidad visual) |
| 7 | `ui-ux` | refinamientos: quitar menú Consumo, idioma a Ajustes, **botón `sm` estándar**, **unificar Servicio+Config**, JSON en modal, switches en Ajustes, **dashboard sin botones** |
| 8 | `ui-ux` | **responsive**: CSS Grid/Flexbox + container queries, contenido centrado (max 1120 px), grids en Resumen/Servicio/Acceso web/Ajustes, ventana mínima 340×420, clase `.btn` estándar |
| 9 | `ui-ux` | **skeletons**: carga independiente por widget (consumo/servicio/MCP/URLs/credenciales/ajustes), errores por card y `Skeleton.vue` accesible |
| 10 | `usage` + `ui-ux` | **token y caché**: lectura OpenCode v2 (`auth export`) + `auth.json` (`opencode`/`opencode-go`) + manual DPAPI; reintento automático/botón; caché nativo de consumo con `staleTime`, dedupe y polling según `RefreshMin` |
| 11 | `shell` + `ui-ux` | **flyout nativo**: ventana frameless fija anclada al icono (Win32, multi-monitor/taskbar/DPI), fuera de taskbar/Alt+Tab, alto según contenido, cierre por blur/Esc, nav inferior fija |
| 12 | `shell` | **cierre estable + tooltip**: guarda de blur post-apertura con verificación de foreground, foco reforzado y tooltip nativo con consumo (`OpenCode · 5h · 7d · 30d`) |
| 13 | `ui-ux` + `service-control` | **pestañas y acciones estables**: métricas fijas del nav, estado+URI juntos (descartando `stopped` como URL), etiquetas cortas en grids, tooltips nativos y scroll con espacio reservado |
| 14 | `shell` + `ui-ux` | **alto fijo y fila de servicio unificada**: el flyout mantiene 600 px en todas las pestañas; Resumen usa el mismo `status-row` que Servicio para estado+URI |
| 15 | `shell` + `web-access` + `ui-ux` | **tooltip mensual humanizado** (barra + iconos + reinicio), **URLs reales con abrir/copiar**, modos exclusivos con apagado del remoto, skeleton de 2 filas e **instalador Inno Setup per-user** |
| 16 | `shell` | **fix del tray**: registrar `energye/systray` en el hilo principal (`main`, antes de `wails.Run`) para que el clic izquierdo abra el panel y el clic derecho muestre el menú |
| 17 | `resumen` + `ui-ux` | **dashboard de widgets**: consumo con las tres barras, servicio y acceso en tiles, MCP en chips, widgets clicables con teclado, tintes de estado suaves y layout bento |
| 18 | `resumen` + `web-access` + `service-control` + `ui-ux` | **frescura de datos**: caché compartido con refresco suave/duro, skeleton solo en primera carga/refresco manual/mutación, invalidación entre pestañas, polling silencioso del servicio y una sola lectura de estado/config por llamada (Snapshot + TTL corto) |
| 19 | `shell` | **Linux/Ubuntu**: icono PNG por SO, auto-inicio XDG, posición del panel en la esquina inferior derecha, toolchain `webkit2_41` y documentación de limitaciones |
| 20 | `shell` | **instalación en Ubuntu**: paquete `.deb` con icono en el menú, instancia única (flock + socket `show`), `--show-panel` y resolución automática del CLI de OpenCode |

Las fases 3a/3b/3c se pueden hacer **en paralelo** (independientes entre sí).

## 4. Riesgos y mitigación

| Riesgo | Impacto | Mitigación |
|---|---|---|
| Bug de `opencode service stop` (no detiene) | Alto | `ForceStop` centralizado: stop → verificar → matar `pid` de `/api/info`. Tests dedicados. |
| Cambio del endpoint `/zen/go/v1/usage` (no es API pública documentada) | Medio | Aislar el parseo en `internal/usage/fetch.go`; tests con fixtures; fallback a mensaje "no se pudo leer" sin romper la app. |
| Librería de tray en Wails v2 (no nativa) | Medio | `energye/systray`; si falla en Linux, alternar a `getlantern/systray` (aislado en `internal/shell/tray.go`). |
| Leer la contraseña del servicio (puede estar oculta) | Bajo | Verificar `opencode service get password` vs leer `~/.config/opencode/service.json`; el contrato `Get("password")` no cambia. |
| WebView2 ausente (Windows) | Bajo | Wails lo gestiona; documentar requisito (viene con Win10/11). |
| Exponer secretos en UI/logs | Alto | Enmascarado centralizado (`configviewer/mask.go`), token nunca en claro, regla "Never" en specs. |
| Token de OpenCode v2 en SQLite y `auth.json` desactualizado | Alto | `opencode auth export` bajo demanda + soporte de `opencode`/`opencode-go` en `auth.json`; reintento automático una vez y botón manual. |
| `opencode auth export` lento (8–15 s) | Medio | Solo en reintento; caché en memoria, datos previos conservados y spinner "Leyendo token…". |

## 5. Checkpoints de verificación (entre fases)

Cada fase termina con su checkpoint antes de pasar a la siguiente:

- **Fase 0**: `wails dev` arranca y `go test ./...` + `pnpm test` corren (aunque estén vacíos).
- **Fase 1**: icono en tray, panel abre/cierra, ajustes persisten, auto-inicio escribe/borra `HKCU\...\Run`, i18n es/en funciona. Tests verdes.
- **Fase 2**: `Status` correcto; `ForceStop` deja el servicio muerto (prueba manual con el bug real).
- **Fase 3a/3b/3c**: consumo se muestra; modos alternan y muestran URLs; visor enmascara secretos y abre el editor.
- **Fase 4**: alertas disparan sin spam; umbrales configurables.
- **Fase 5**: `wails build` produce `.exe`; checklist de `ui-ux-pro-max` (contraste, foco, targets) aplicado.
- **Fase 8**: sin scroll horizontal en 340–2560 px; contenido centrado; ventana mínima respetada; `pnpm build` + tests verdes.
- **Fase 9**: cada widget async se pinta al resolver su propia carga; skeletons accesibles y sin saltos de layout; tests verdes.
- **Fase 10**: token leído de `opencode auth export`/`auth.json`/manual DPAPI; `GetSettings` nunca expone el secreto; consumo cacheado sin skeleton al remontar; reintento único + CTAs; `go test ./...`, `pnpm test`, `pnpm build` y `wails build` en verde.
- **Fase 11**: flyout anclado al icono, sin mover/redimensionar, fuera de taskbar/Alt+Tab, alto según contenido, cierre por blur/Esc; pruebas manuales en 4 posiciones de taskbar, multi-monitor y DPI 100/150/200 %.
- **Fase 12**: un clic abre y mantiene el panel (sin cierre por blur transitorio); clic fuera/Esc/tray lo cierran; el tooltip del tray muestra consumo tras `UsageFetch`; `go test ./...`, `pnpm test`, `pnpm build` y `wails build` en verde.
- **Fase 13**: pestañas sin movimiento, estado+URI juntos, botones cortos sin saltos, tooltips en botones icono y scroll sin reflujo; `pnpm test`, `pnpm build`, `go test ./...` y `wails build` en verde.
- **Fase 14**: mismo alto en todas las pestañas sin medir contenido, y estado+URI alineados en Resumen; tests y build en verde.
- **Fase 15**: tooltip con barra/porcentaje/reinicio actualizado sin abrir el panel; URLs local/LAN/remota con abrir/copiar; remoto exclusivo con confirmación; skeleton de 2 filas; instalador per-user generado; `go test ./...`, `pnpm test`, `pnpm build` y `wails build` en verde.
- **Fase 16**: clic izquierdo abre/cierra el panel, clic derecho abre el menú y "Salir" cierra la app; tray creado en el hilo principal; tests y build en verde.
- **Fase 17**: Resumen con 4 widgets, barras de consumo intactas, servicio/acceso con estado visual accesible, MCP sin estado textual por item, navegación por clic/teclado; tests y build en verde.
- **Fase 18**: navegar entre pestañas con datos frescos no hace petición ni muestra skeleton; refresco manual y mutaciones muestran skeleton desde el inicio; Resumen refleja el estado tras stop/start; sin textos de "Actualizando…"; `WebAccessState` hace 1 status + 1 config (y `/api/info` solo en red/remoto); tests y build en verde.
- **Fase 19**: `go test ./...` y `pnpm test` pasan en Linux; `wails build -tags webkit2_41` produce un binario que arranca en Ubuntu; el icono del tray se dibuja; el toggle de auto-inicio crea/borra `~/.config/autostart/opencode-tray.desktop`; en Windows no cambia nada.
- **Fase 20**: `./scripts/build-deb.sh` genera `dist/opencode-tray_<versión>_amd64.deb`; `dpkg-deb -c` muestra `/usr/bin/opencode-tray`, el `.desktop` y el icono; lanzada desde el menú abre el panel; una segunda ejecución sale en <50 ms sin duplicar el tray; el CLI se resuelve sin estar en el `PATH`; `apt remove` conserva los ajustes; tests y build en verde.

## 6. Estrategia de testing (desde el inicio)

- **Go**: `testing` estándar, table-driven. Todo I/O externo (exec, HTTP, archivos, notificaciones) se inyecta como interface/func para mockear. Ningún test toca `opencode` real ni red real.
- **Vue**: Vitest + @vue/test-utils para componentes de sección.
- **Sin e2e** (se evalúa después de la v1).
- **CI local**: `go test ./...` y `cd frontend && pnpm test` como pre-commit mental (y hook opcional después).

## 7. Convenciones transversales

- **Consistencia de archivos**: cualquier cambio que toque comandos, estructura o config global **debe** actualizar README.md, `tasks/plan.md`, `tasks/todo.md` y las `specs/SPEC-*.md` afectadas en el mismo cambio.
- **i18n**: todo texto de UI vía claves en `frontend/src/i18n/{es,en}.json`.
- **UI/UX**: sistema de diseño generado con `ui-ux-pro-max` y persistido en `design-system/opencode-tray/MASTER.md` (generado y aplicado en la Fase 5): paleta, tipografía, espaciado y checklist de accesibilidad (contraste ≥ 4.5:1, foco visible, targets ≥ 44px).
- **Seguridad**: secretos enmascarados, token/contraseña nunca en logs ni commits.

## 8. Distribución

- **v1**: `.exe` portable (un solo binario vía `wails build`).
- **Linux**: binario ELF con `wails build -tags webkit2_41`, para desarrollo y prueba local (sin instalador ni paquete `.deb`).
- **Posterior**: instalador NSIS. (Registrado para no olvidar.)
