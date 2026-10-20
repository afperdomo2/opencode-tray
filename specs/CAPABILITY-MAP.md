# Capability Map: OpenCode Tray

> Estado: **APROBADO** ✅ (Phase 0 del flujo SDD).

## Módulos

| Module id | Responsabilidad | Depende de |
|---|---|---|
| `shell` | App Wails: icono de bandeja, ciclo de vida de la ventana del panel, una sola instancia, persistencia de ajustes (config local), toggle de auto-inicio con Windows, i18n (es/en) | — |
| `service-control` | Estado del servicio de fondo de OpenCode, `start` / `stop` / `restart`, y **force-stop** (workaround: leer `pid` de `/api/info` y matar el proceso) | `shell` |
| `web-access` | Modos de acceso web: **local** / **red (por IP)** / **remoto (túnel)**; mostrar y cambiar el usuario (`opencode`) y la contraseña | `shell`, `service-control` |
| `usage` | Consultar y mostrar el consumo de OpenCode Go (ventanas `rolling` / `weekly` / `monthly`), resolver el token vigente (OpenCode v2 export → `auth.json` legado → manual DPAPI) y releerlo bajo demanda | `shell` |
| `config-viewer` | Ver MCP servers, skills, agents y el JSON de configuración global (solo lectura + abrir en editor externo) | `shell` |
| `notifications` | Notificaciones nativas (toast) de **dos tipos**: (A) ritmo de consumo vs. ritmo objetivo y (B) agotamiento por umbral | `shell`, `usage`, `service-control` |
| `resumen` | **Vista de overview (landing)** del panel: dashboard compacto (consumo → MCP/config → servicio) + ajustes de UI/UX del panel (iconos, cabecera, idioma, acordeón, profundidad). **Frontend, sin paquete Go.** | `shell`, `service-control`, `usage`, `config-viewer` |
| `ui-ux` | **Refinamientos de UI/UX (iteración 2)**: quitar el menú Consumo, mover el idioma a Ajustes, botones de servicio compactos en Resumen, JSON de config en modal, switches estilizados en Ajustes. **Frontend, sin paquete Go.** | `shell`, `resumen`, `config-viewer` |

## Orden de construcción

```
shell  →  service-control  →  ┬─ web-access
                              ├─ usage
                              └─ config-viewer
                                            →  notifications
                                            →  resumen (frontend)
```

`web-access`, `usage` y `config-viewer` son **independientes entre sí** (paralelos). `notifications` y `resumen` van al final (`notifications` reacciona a `usage`/`service-control`; `resumen` **compone** `usage` + `config-viewer` + `service-control` en la UI).

## Contratos en los límites

- `web-access` depende de `service-control`: usa `configure(hostname, port, password)` y `start/stop` (contrato en la spec de `service-control`).
- `notifications` depende de `usage` (datos de ritmo/umbral) y `service-control` (estado), pero solo los **lee**; no los controla.
- `usage` y `config-viewer` solo dependen de `shell`.

## Decisiones cerradas (grilling — Ronda 1)

1. **Notificaciones (dos tipos):**
   - **(A) Ritmo (burn rate)**: en hitos de consumo 20/40/60/80%, dispara si `consumido − esperado ≥ margen` (esperado = % tiempo transcurrido de la ventana; margen por defecto **15 pts**). Severidad amarillo.
   - **(B) Agotamiento**: umbrales fijos **90%** (naranja, "casi al límite"), **98%** (rojo, "al límite, cambia ya") y **100%** (crítico, "se acabó").
   - Ambos configurables en Ajustes.
2. **UI**: panel popover compacto anclado al tray, con secciones (Resumen, Servicio, Acceso web, Consumo, Config, Ajustes).
3. **i18n**: español + inglés, con selector de idioma.
4. **Visor de config**: solo lectura + "abrir en editor" (VS Code si `code` está en PATH, si no, editor por defecto de Windows).
5. **Tests desde el inicio**: unitarios Go (`testing`) + componentes Vue (Vitest). Sin e2e por ahora.
6. **Distribución**: `.exe` portable primero; instalador NSIS después (pendiente, no olvidar).
7. **Nombre**: "OpenCode Tray".

## Hechos a verificar en implementación (no son decisiones)

- Mecanismo exacto para **leer/cambiar la contraseña** del servicio (`~/.config/opencode/service.json` vs `opencode service get password`).
- Vía de **auto-inicio** en Windows (`HKCU\...\Run`).

## Preocupaciones transversales (aplican a todos los módulos)

1. **UI/UX** — Skill `ui-ux-pro-max` para el sistema de diseño (estilo, paleta, tipografía, espaciado) y accesibilidad (contraste 4.5:1, foco visible, targets ≥ 44px, modo claro/oscuro).
2. **Testing desde el inicio** — Ninguna tarea "hecha" sin su test.
3. **Consistencia entre archivos** — README.md, `specs/SPEC-*.md`, `tasks/plan.md`, `tasks/todo.md` y estructura de directorios siempre en sincronía (regla "Always do").

## Convenciones de archivo

- Mapa: `specs/CAPABILITY-MAP.md` (este archivo).
- Spec por módulo: `specs/SPEC-<module-id>.md`.
- Plan: `tasks/plan.md`. Tareas: `tasks/todo.md`.
- Sistema de diseño (`ui-ux-pro-max`): `design-system/<slug>/MASTER.md`.
