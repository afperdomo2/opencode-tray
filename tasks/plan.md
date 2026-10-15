# Plan técnico — OpenCode Tray

> Phase 2 del flujo SDD. Fuente de verdad: `specs/CAPABILITY-MAP.md` + las 6 `specs/SPEC-*.md` aprobadas.
> Tras aprobar este plan, sigue Phase 3 (`tasks/todo.md`).

## 1. Resumen

App de bandeja (tray) para Windows con Wails v2 (Go) + Vue 3/Vite/TS que controla el servicio de OpenCode, el acceso web (local/red/remoto), muestra el consumo de OpenCode Go, permite ver la config (MCP/skills/agents) y emite notificaciones nativas. Se construye **por módulos, en orden de dependencias, con tests desde el inicio**.

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

## 5. Checkpoints de verificación (entre fases)

Cada fase termina con su checkpoint antes de pasar a la siguiente:

- **Fase 0**: `wails dev` arranca y `go test ./...` + `pnpm test` corren (aunque estén vacíos).
- **Fase 1**: icono en tray, panel abre/cierra, ajustes persisten, auto-inicio escribe/borra `HKCU\...\Run`, i18n es/en funciona. Tests verdes.
- **Fase 2**: `Status` correcto; `ForceStop` deja el servicio muerto (prueba manual con el bug real).
- **Fase 3a/3b/3c**: consumo se muestra; modos alternan y muestran URLs; visor enmascara secretos y abre el editor.
- **Fase 4**: alertas disparan sin spam; umbrales configurables.
- **Fase 5**: `wails build` produce `.exe`; checklist de `ui-ux-pro-max` (contraste, foco, targets) aplicado.

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
- **Posterior**: instalador NSIS. (Registrado para no olvidar.)
