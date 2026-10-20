# OpenCode Tray

App de bandeja (tray) para Windows que controla el servicio de OpenCode, gestiona el acceso web (local/red/remoto), muestra el consumo de OpenCode Go y permite ver la configuración (MCP/skills/agents).

> Estado: **funcional (Fase 11)** — todas las capacidades implementadas y con tests. La planificación vive en [`specs/`](specs/) y las tareas en [`tasks/`](tasks/).

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

- **Panel (flyout)**: ventana frameless anclada al icono del tray, sin marco, sin resize ni arrastre. Se abre con clic en el icono, se cierra con `Esc` o al perder el foco y mantiene un alto fijo de 600 px en todas las pestañas. No aparece en la barra de tareas ni en Alt+Tab. Las pestañas tienen tamaño fijo y el scroll reserva su espacio para no deformar el contenido. Al pasar el cursor por el icono, el tooltip nativo muestra el consumo (`OpenCode · 5h · 7d · 30d`). Detalle en [`specs/SPEC-shell.md`](specs/SPEC-shell.md).
- **Resumen**: dashboard de inicio (**solo lectura**; únicas acciones: reintentar el token o configurar la API key si falla) con widgets de consumo (barras con **color graduado** según proximidad al límite) y de servicio + config (estado y servidores MCP).
- **Servicio** (unifica el antiguo Config): estado + URI con puerto y control con etiquetas cortas (iniciar / detener / forzar / refrescar) + servidores MCP + configuración (editor / carpeta / **JSON en modal**). Los botones muestran su descripción completa al pasar el cursor.
- **Acceso web**: modos local / red (por IP) / remoto (túnel); usuario (`opencode`) y contraseña (enmascarada) con cambio de contraseña.
- **Ajustes**: idioma (es/en), auto-inicio con Windows, switches de notificaciones, tarjeta de **OpenCode Go** (estado del token, reintento desde OpenCode, API key manual cifrada con DPAPI) e intervalo de refresco (1/5/15/30/60 min).

La UI usa iconos SVG (sin emoji), botones `btn` en tamaño compacto estándar, control segmentado de idioma, switches estilizados y cards con profundidad (design system). El layout es responsive con CSS Grid/Flexbox y container queries, sin dependencias de UI. La navegación es un footer fijo de 4 pestañas (icono + texto) y el contenido es la única zona con scroll. Cada widget async (consumo, servicio, MCP, URLs, ajustes) carga con su propio **skeleton**. El consumo usa **caché en memoria**: al volver a Resumen no hay skeleton ni petición si los datos son frescos.

> **Notificaciones**: la **lógica** de alertas (ritmo de consumo, agotamiento por umbral, servicio detenido inesperadamente) está implementada y probada, pero el **envío de toasts está deshabilitado en runtime** — el mecanismo de toast de Wails (go-toast) entraba en conflicto con el bucle de la bandeja y la dejaba muda. Detalle y cómo reactivarlo en [`specs/SPEC-notifications.md`](specs/SPEC-notifications.md).

## Estructura

- [`specs/`](specs/) — mapa de capacidades + specs por módulo (fuente de verdad del "qué").
- [`tasks/`](tasks/) — plan técnico (`plan.md`) y tareas (`todo.md`).
- [`internal/`](internal/) — backend Go, un paquete por módulo: `shell`, `servicecontrol`, `webaccess`, `usage`, `configviewer`, `notifications`.
- [`frontend/`](frontend/) — UI Vue 3 (`src/components/` por sección, `src/lib/` con caché/ajustes/consumo, i18n es/en).
- [`design-system/opencode-tray/MASTER.md`](design-system/opencode-tray/MASTER.md) — sistema de diseño generado con `ui-ux-pro-max` (paleta, tipografía, espaciado, a11y).
- [`AGENTS.md`](AGENTS.md) — reglas de trabajo y convenciones.

## Requisitos

- Windows 10/11 con **WebView2** (incluido por defecto en Windows moderno).
- CLI de **OpenCode** instalada (`opencode` disponible en PATH), de la que se controla el servicio de fondo.

## Limitaciones conocidas

- **Flyout**: el hover solo muestra el tooltip nativo con consumo; no hay preview visual. El panel tampoco se puede redimensionar ni arrastrar por diseño.
- **Notificaciones**: envío de toasts deshabilitado en runtime (ver arriba). La lógica permanece y los toggles/evaluación funcionan, pero no se muestra ningún toast.
- **Fuentes web**: la UI carga IBM Plex Sans + JetBrains Mono desde Google Fonts (`display=swap`); **sin red degrada a las fuentes del sistema** y la app sigue siendo funcional.
- **Modo `wails dev` y la comunicación frontend ↔ backend**: en modo desarrollo (`wails dev`) el frontend a veces **no logra alcanzar el backend** (el puente `window.go` de Wails queda sin definir), por lo que las llamadas al Go fallan (aparece un `TypeError: Cannot read properties of undefined (reading 'main')` y las secciones muestran "Detenido" aunque el servicio corra). Es un **quirk del modo dev de Wails** (su proxy hacia el servidor de Vite), no un bug de esta app. Para **probar el comportamiento real** (panel de Servicio, consumo, etc.) usa el **build de producción**:

  ```bash
  wails build          # genera build/bin/opencode-tray.exe
  ```

  y ejecuta `build/bin/opencode-tray.exe`. En el build de producción todo funciona correctamente.

## Planificación

El proyecto se construye bajo un flujo **Spec-Driven Development (SDD)**: primero las specs (`specs/`), luego el plan y las tareas (`tasks/`), y solo después la implementación. Ver `specs/CAPABILITY-MAP.md` para el mapa de módulos y el orden de construcción.
