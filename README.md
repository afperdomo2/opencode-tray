# OpenCode Tray

App de bandeja (tray) para Windows que controla el servicio de OpenCode, el acceso web (local/red/remoto), muestra el consumo de OpenCode Go, permite ver la configuración (MCP/skills/agents) y emite notificaciones nativas.

> Estado: **en desarrollo** (fase scaffold). La planificación vive en [`specs/`](specs/) y las tareas en [`tasks/`](tasks/).

## Stack

Wails v2 + Go 1.24+ + Vue 3 / Vite / TypeScript.

## Comandos

```bash
wails dev                 # desarrollo (hot reload)
wails build               # build (.exe portable)
go test ./...             # tests backend
cd frontend && pnpm test   # tests frontend (Vitest)
```

## Estructura

- [`specs/`](specs/) — mapa de capacidades + specs por módulo (fuente de verdad del "qué").
- [`tasks/`](tasks/) — plan técnico y tareas.
- [`internal/`](internal/) — backend Go (un paquete por módulo).
- [`frontend/`](frontend/) — UI Vue 3.
- [`AGENTS.md`](AGENTS.md) — reglas de trabajo y convenciones.

## Limitaciones conocidas

- **Modo `wails dev` y la comunicación frontend ↔ backend**: en modo desarrollo (`wails dev`) el frontend a veces **no logra alcanzar el backend** (el puente `window.go` de Wails queda sin definir), por lo que las llamadas al Go fallan (aparece un `TypeError: Cannot read properties of undefined (reading 'main')` y las secciones muestran "Detenido" aunque el servicio corra). Es un **quirk del modo dev de Wails** (su proxy hacia el servidor de Vite), no un bug de esta app. Para **probar el comportamiento real** (panel de Servicio, consumo, etc.) usa el **build de producción**:

  ```bash
  wails build          # genera build/bin/opencode-tray.exe
  ```

  y ejecuta `build/bin/opencode-tray.exe`. En el build de producción todo funciona correctamente.

## Planificación

El proyecto se construye bajo un flujo **Spec-Driven Development (SDD)**: primero las specs (`specs/`), luego el plan y las tareas (`tasks/`), y solo después la implementación. Ver `specs/CAPABILITY-MAP.md` para el mapa de módulos y el orden de construcción.
