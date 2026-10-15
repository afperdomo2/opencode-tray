# OpenCode Tray

App de bandeja (tray) para Windows que controla el servicio de OpenCode, gestiona el acceso web (local/red/remoto), muestra el consumo de OpenCode Go y permite ver la configuración (MCP/skills/agents).

> Estado: **funcional (Fase 5)** — todas las capacidades implementadas y con tests. La planificación vive en [`specs/`](specs/) y las tareas en [`tasks/`](tasks/).

## Stack

Wails v2 + Go 1.24+ + Vue 3 / Vite / TypeScript.

## Comandos

```bash
wails dev                 # desarrollo (hot reload)
wails build               # build (.exe portable) → build/bin/opencode-tray.exe
go test ./...             # tests backend
cd frontend && pnpm test   # tests frontend (Vitest)
cd frontend && pnpm build  # type-check (vue-tsc) + build del frontend
```

## Funcionalidad

- **Servicio**: estado (detenido/corriendo), iniciar / detener / **forzar detención** (workaround del bug de `opencode service stop`) / refrescar.
- **Consumo**: barras del consumo de OpenCode Go (ventanas `rolling` / `weekly` / `monthly`) con estado ok / warn / crítico.
- **Acceso web**: modos local / red (por IP) / remoto (túnel); usuario (`opencode`) y contraseña (enmascarada) con cambio de contraseña.
- **Config**: visor de servidores MCP y del JSON de configuración global (secretos enmascarados) + abrir en editor externo.
- **Ajustes**: auto-inicio con Windows, idioma (es/en) y toggles de notificaciones.

> **Notificaciones**: la **lógica** de alertas (ritmo de consumo, agotamiento por umbral, servicio detenido inesperadamente) está implementada y probada, pero el **envío de toasts está deshabilitado en runtime** — el mecanismo de toast de Wails (go-toast) entraba en conflicto con el bucle de la bandeja y la dejaba muda. Detalle y cómo reactivarlo en [`specs/SPEC-notifications.md`](specs/SPEC-notifications.md).

## Estructura

- [`specs/`](specs/) — mapa de capacidades + specs por módulo (fuente de verdad del "qué").
- [`tasks/`](tasks/) — plan técnico (`plan.md`) y tareas (`todo.md`).
- [`internal/`](internal/) — backend Go, un paquete por módulo: `shell`, `servicecontrol`, `webaccess`, `usage`, `configviewer`, `notifications`.
- [`frontend/`](frontend/) — UI Vue 3 (`src/components/` por sección, i18n es/en).
- [`design-system/opencode-tray/MASTER.md`](design-system/opencode-tray/MASTER.md) — sistema de diseño generado con `ui-ux-pro-max` (paleta, tipografía, espaciado, a11y).
- [`AGENTS.md`](AGENTS.md) — reglas de trabajo y convenciones.

## Requisitos

- Windows 10/11 con **WebView2** (incluido por defecto en Windows moderno).
- CLI de **OpenCode** instalada (`opencode` disponible en PATH), de la que se controla el servicio de fondo.

## Limitaciones conocidas

- **Notificaciones**: envío de toasts deshabilitado en runtime (ver arriba). La lógica permanece y los toggles/evaluación funcionan, pero no se muestra ningún toast.
- **Fuentes web**: la UI carga IBM Plex Sans + JetBrains Mono desde Google Fonts (`display=swap`); **sin red degrada a las fuentes del sistema** y la app sigue siendo funcional.
- **Modo `wails dev` y la comunicación frontend ↔ backend**: en modo desarrollo (`wails dev`) el frontend a veces **no logra alcanzar el backend** (el puente `window.go` de Wails queda sin definir), por lo que las llamadas al Go fallan (aparece un `TypeError: Cannot read properties of undefined (reading 'main')` y las secciones muestran "Detenido" aunque el servicio corra). Es un **quirk del modo dev de Wails** (su proxy hacia el servidor de Vite), no un bug de esta app. Para **probar el comportamiento real** (panel de Servicio, consumo, etc.) usa el **build de producción**:

  ```bash
  wails build          # genera build/bin/opencode-tray.exe
  ```

  y ejecuta `build/bin/opencode-tray.exe`. En el build de producción todo funciona correctamente.

## Planificación

El proyecto se construye bajo un flujo **Spec-Driven Development (SDD)**: primero las specs (`specs/`), luego el plan y las tareas (`tasks/`), y solo después la implementación. Ver `specs/CAPABILITY-MAP.md` para el mapa de módulos y el orden de construcción.
