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

## Planificación

El proyecto se construye bajo un flujo **Spec-Driven Development (SDD)**: primero las specs (`specs/`), luego el plan y las tareas (`tasks/`), y solo después la implementación. Ver `specs/CAPABILITY-MAP.md` para el mapa de módulos y el orden de construcción.
