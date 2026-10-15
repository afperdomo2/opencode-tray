# Spec: ui-ux (OpenCode Tray)

> Refinamientos de UI/UX del panel (**iteración 2**). Traza al module id `ui-ux` del `CAPABILITY-MAP.md`. Afecta `shell` (navegación, cabecera, Ajustes), `resumen` y `config-viewer`. Frontend, sin backend nuevo.

## Objective

Simplificar y pulir la experiencia del panel: menos navegación redundante, controles más compactos y mejor presentación de la configuración.

### Cambios

1. **Quitar el menú "Consumo".** El consumo ya se muestra en **Resumen**; se elimina la pestaña/sección **Consumo** y su componente (`Consumo.vue`). Los datos siguen disponibles en Resumen.

2. **Mover el idioma a Ajustes.** El selector de idioma sale de la cabecera y se ubica en **Ajustes** (control segmentado ES | EN). Al quedarse sin título (Fase 6) y sin selector, la **cabecera se elimina**.

3. **Resumen → Servicio: botones compactos (tamaño `sm`).** Los cuatro botones (iniciar / detener / forzar detención / actualizar) pasan a un tamaño **`sm`** (menos padding, fuente e icono menores) **conservando el texto**. Objetivo: que quepan en **una sola fila** en el panel de 380 px; si el texto no lo permite, envuelven de forma compacta.

4. **Resumen → Config: 3 botones (tamaño `sm`) + JSON en modal.** El bloque Config del Resumen muestra **tres botones** compactos (*Abrir en editor*, *Abrir carpeta* y *Configuración (JSON)*). El tercero abre el JSON en un **modal**.

   > **Evaluación (modal vs. vista):** se opta por **modal**. Motivo: no saca al usuario del contexto de Resumen, reutiliza un único componente para Resumen y la sección Config, y evita reestructurar el sistema de vistas del panel. El JSON (potencialmente largo) va en una superficie desplazable dentro del modal. Se implementa accesible (`role="dialog"`, `aria-modal`, Esc, foco, `aria-label` en cerrar).

5. **Ajustes: checkboxes con mejor diseño.** Los `input[type=checkbox]` nativos se reemplazan por un **switch** (pill con knob deslizante) estilizado y accesible.

6. **Resumen: color de consumo graduado.** En las barras de consumo, el color refleja la **proximidad al límite**, de forma visible en la barra **y** en el número `%`:

   | Estado | Umbral | Color |
   |---|---|---|
   | `ok` | `< 70 %` | verde |
   | `warn` | `70 – 84 %` | ámbar |
   | `high` | `85 – 97 %` | naranja |
   | `crit` | `≥ 98 %` | rojo |

   Se mantiene el **rojo en 98 %** (coherente con el spec de `usage`/notificaciones) y se añaden escalones previos (ámbar/naranja) para dar aviso **antes** de llegar al límite. El número `%` acompaña al color, así que la severidad **no** se transmite solo por color.

7. **Estandarizar el tamaño de botón (`sm`).** El tamaño `sm` (usado en el Resumen) pasa a ser **el estándar de todos los menús/interfaces** (Servicio, Acceso web, Ajustes): `min-height 30px`, `font-size 12px`, icono `14px`, `padding 4px 8px`.

8. **Unificar los menús "Servicio" y "Config".** Se **fusionan en una sola sección** (`Servicio`): control del servicio + servidores MCP + configuración (abrir editor/carpeta y JSON). El nav queda: **Resumen · Servicio · Acceso web · Ajustes**.

9. **Dashboard (Resumen) = widgets sin botones.** El Resumen **unifica Servicio + Config en un widget** de solo lectura (estado + lista MCP) y **elimina los botones de acción**: las acciones viven en su menú. El dashboard es para **resúmenes/acceso**, no para operar.

## Tech Stack

- Vue 3 (`<script setup lang="ts">`) + TypeScript; `vue-i18n`; iconos SVG inline (`Icon.vue`).

## Project Structure

```
frontend/src/
├── components/
│   ├── Modal.vue        # NUEVO: overlay accesible (role=dialog, Esc, foco, backdrop)
│   ├── Switch.vue       # NUEVO: switch estilizado (checkbox nativo oculto, visual propio)
│   ├── Consumo.vue      # ELIMINADO (su contenido queda en Resumen)
│   ├── Config.vue       # ELIMINADO (fusionado en Servicio)
│   ├── Servicio.vue     # UNIFICADO: servicio + MCP + configuración (JSON en modal)
│   ├── Resumen.vue      # dashboard: widgets (Consumo · Servicio+Config) SIN botones
│   ├── Ajustes.vue      # idioma (LanguageSwitch) + switches
│   ├── LanguageSwitch.vue
│   └── *.spec.ts
└── App.vue              # sin cabecera; nav: Resumen · Servicio · Acceso web · Ajustes
```

## Code Style

- **Switch** (accesible, nativo oculto):

```vue
<label class="switch">
  <input type="checkbox" :checked="checked" @change="$emit('change', $event)" />
  <span class="track"><span class="thumb"></span></span>
</label>
```

- **Modal** accesible: `role="dialog"` + `aria-modal="true"` + `aria-labelledby`; cierre con **Esc** y clic en el backdrop; botón cerrar con `aria-label`; el foco entra al modal al abrir.

## Testing Strategy

- **Vitest** (@vue/test-utils):
  - `Modal`: está cerrado por defecto; abre/cierra con el botón y con **Esc**; `aria-modal="true"`.
  - `Switch`: refleja `checked`; emite `change`.
  - `Resumen`: los 4 botones de servicio están en una fila (`.actions.compact`); el bloque config tiene **3 botones**; el botón JSON abre el modal.
  - `Ajustes`: contiene el control de idioma y usa switches.
  - Regresión: `App.vue` ya no tiene la sección `consumo`.

## Boundaries

- **Always do:** i18n para todo texto; a11y (foco, targets, contraste ≥ 4.5:1); iconos **SVG** (no emoji); reusar tokens del design system.
- **Ask first:** agregar dependencias pnpm (aquí, ninguna nueva).
- **Never:** hardcodear colores fuera de tokens; dejar **código muerto** (eliminar `Consumo.vue` y sus referencias).

## Success Criteria

- [ ] No existe la pestaña **Consumo**; el consumo se ve en **Resumen**.
- [ ] El **idioma** vive en **Ajustes**; la cabecera (título/selector) desaparece.
- [ ] **Resumen → Servicio**: los botones usan tamaño **`sm`** y **muestran texto** (compactos).
- [ ] **Resumen → Config**: **3 botones** (tamaño `sm`); el JSON se abre en **modal** (también en la sección Config).
- [ ] **Ajustes**: los checkboxes son un **switch** estilizado y accesible.
- [ ] **Resumen**: las barras y el `%` cambian de color de forma graduada (verde → ámbar → naranja → rojo) según la proximidad al límite.
- [ ] **Botones**: el tamaño `sm` es el **estándar** en todos los menús.
- [ ] **Menús**: "Servicio" y "Config" están **unificados** (nav: Resumen · Servicio · Acceso web · Ajustes).
- [ ] **Dashboard**: Resumen muestra **widgets sin botones** (Consumo · Servicio+Config).
- [ ] Tests Vitest en verde; a11y mantenido.

## Decisiones cerradas

1. **Consumo** se elimina de la navegación; permanece en **Resumen**.
2. **Idioma** se mueve a **Ajustes** (segmentado ES | EN); se **elimina la cabecera**.
3. **JSON de config** → **modal** (no acordeón inline ni vista aparte).
4. **Tamaño de botón `sm`** (30px / 12px / icono 14) es el **estándar en toda la app**.
5. **Ajustes** → **switch** personalizado en lugar de checkbox nativo.
6. **Servicio y Config** se **fusionan** en un único menú (`Servicio`).
7. **Dashboard sin acciones**: el Resumen es solo **resumen/acceso** (widgets de solo lectura, **sin botones**); las acciones viven en su menú.
