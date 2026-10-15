# Spec: resumen (OpenCode Tray)

> Traza al module id `resumen` del `CAPABILITY-MAP.md`. **Vista de overview (landing)** + **ajustes de UI/UX del panel**. Depende de `shell`, `service-control`, `usage` y `config-viewer`. Es la última pieza (compone lo que ya existe).

## Objective

`resumen` es la **vista de aterrizaje por defecto** del panel: un **dashboard compacto** que centraliza el estado clave de un vistazo, más una pasada de **UI/UX** para que el panel se sienta como una app de escritorio (Windows) y no como una web plana.

### Parte A — Vista Resumen (frontend, sin módulo Go nuevo)

Compone datos ya existentes (`ServiceStatus`, `UsageFetch`, `ConfigListMCP`, `ConfigJSON`, `ServiceStart/Stop/ForceStop`). **Orden fijo de bloques**, cada uno como una **card** con jerarquía clara:

1. **Consumo (primero).** Resumen compacto de las tres ventanas (`rolling` / `weekly` / `monthly`) con su % y severidad (ok / warn / crítico). Es el dato que más se consulta.
2. **MCP / Config.** Lista compacta de servidores MCP (nombre + estado) y, debajo, un **acordeón** con el **JSON de configuración** — **colapsado por defecto** (si no, la vista crece demasiado).
3. **Servicio (último).** Estado (corriendo / detenido) + acciones rápidas (iniciar / detener / actualizar).

### Parte B — Ajustes de UI/UX del panel

1. **Iconos (SVG inline, nunca emoji).** En navegación (una por sección), botones de acción (iniciar / detener / forzar / actualizar), mostrar/ocultar contraseña, abrir editor / carpeta, y el chevron del acordeón. Aportan claridad y usabilidad.
2. **Cabecera.** **Eliminar el título "OpenCode Tray"** (redundante con el título de la ventana). La cabecera queda para el **control de idioma**.
3. **Idioma.** Reemplazar el `<select>` por un **control segmentado** (ES | EN), intuitivo y "bonito" para solo dos opciones, con icono (globo / idiomas).
4. **Config.** El **JSON de configuración** se muestra en un **acordeón** (colapsado por defecto) tanto en la vista Resumen como en la sección Config.
5. **Profundidad visual ("menos plano").** Superficies/cards, sombras sutiles (tokens del design system), jerarquía tipográfica y estados hover: sensación de app de escritorio.

**Contrato:** **sin cambios de backend**; reusa bindings existentes. Es una capa de composición + presentación en el frontend.

## Tech Stack

- Vue 3 (`<script setup lang="ts">`) + TypeScript.
- Iconos: **SVG inline** (sin dependencia nueva de pnpm). Set estilo *outline*.
- `vue-i18n` para todo el texto.

## Project Structure

```
frontend/src/
├── components/
│   ├── Resumen.vue          # NUEVO: vista overview (Consumo → MCP/Config → Servicio)
│   ├── Icon.vue             # NUEVO: set de iconos SVG inline por nombre
│   ├── Accordion.vue        # NUEVO: disclosure accesible (aria-expanded)
│   ├── LanguageSwitch.vue   # NUEVO: control segmentado ES|EN
│   ├── Servicio.vue · Consumo.vue · AccesoWeb.vue · Config.vue · Ajustes.vue   # + iconos
│   └── *.spec.ts
├── App.vue                  # cabecera sin título; usa LanguageSwitch; resumen → Resumen
└── i18n/{es,en}.json        # claves nuevas
```

## Code Style

- **Iconos** vía componente: `<Icon name="globe" />` → `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" ...>`. Decorativos → `aria-hidden="true"`; botones solo-icono → `aria-label`.
- **Acordeón** accesible:

```vue
<button class="acc-head" :aria-expanded="open" :aria-controls="panelId" @click="open = !open">
  <Icon name="chevron-down" :class="{ open }" aria-hidden="true" />
  <span>{{ title }}</span>
</button>
<div :id="panelId" v-show="open">…</div>
```

(El skill `ui-ux-pro-max` no devolvió match para "accordion" en su BD; se usa el patrón estándar de *disclosure* accesible.)

## Testing Strategy

- **Vitest** + `@vue/test-utils`, mockeando los bindings de Wails:
  - `Resumen.vue`: estados cargando / servicio corriendo / detenido / consumo ok-warn-crítico / error.
  - `Accordion.vue`: colapsado por defecto; abre y cierra al clic; `aria-expanded` correcto.
  - `LanguageSwitch.vue`: cambia el locale (es ↔ en).
  - `Icon.vue`: renderiza `<svg>` por nombre; decorativo → oculto a lectores.

## Boundaries

- **Always do:** iconos **SVG** (nunca emoji); todo texto vía i18n; a11y (contraste ≥ 4.5:1, foco visible, targets ≥ 44px); reusar **tokens** del design system (`--color-*`, `--space-*`).
- **Ask first:** agregar una **librería de iconos** (pnpm). Se evita con SVG inline; si se prefiere librería, se consulta.
- **Never:** hardcodear colores fuera de los tokens del design system.

## Success Criteria

- [ ] `Resumen` es la vista inicial y muestra **Consumo → MCP/Config (acordeón) → Servicio**, de forma compacta y bien seccionada.
- [ ] **Iconos** en navegación, botones de acción y controles relevantes; **sin emoji**.
- [ ] Cabecera **sin título**; idioma como **control segmentado** ES | EN.
- [ ] El **JSON de config** está **colapsado por defecto** (acordeón) en Resumen y en Config.
- [ ] El panel tiene **profundidad visual** (cards, sombras, hover): no se siente plano.
- [ ] Tests Vitest en verde; a11y (contraste/foco/targets) mantenido.

## Decisiones cerradas

1. **Resumen = landing por defecto.** Se **conservan** las secciones detalladas (Servicio, Consumo, Acceso web, Config, Ajustes) para control directo; Resumen solo agrega/centraliza.
2. **Iconos SVG inline**, set outline (inspirado en Phosphor/Lucide). Sin dependencia nueva.
3. **Orden del Resumen:** Consumo → MCP/Config → Servicio.
4. Esta spec cubre además **ajustes de UI/UX transversales** del panel (cabecera, idioma, acordeón, profundidad), no solo la vista Resumen.
