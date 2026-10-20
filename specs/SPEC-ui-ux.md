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

---

# Iteración 3 — Responsive (todos los tamaños)

> Traza al module id `ui-ux`. Afecta al shell, `resumen`, `service-control`, `web-access` y `config-viewer`. Frontend + configuración de ventana; sin backend nuevo.

## Objective

Que el panel se adapte de forma real a cualquier tamaño de ventana (desde ~340 px hasta monitores de 27" o más, y tanto en horizontal como en vertical), sin estirar widgets ni generar scroll horizontal.

## Diagnóstico

- `style.css` no tenía breakpoints de layout; `.app-content` y `.card` ocupaban el 100% del ancho.
- No hay Tailwind ni librería de layout: se usan tokens CSS propios.
- Los datos de la app no son tabulares (3 barras de consumo + lista MCP): **no se justifica un DataGrid**.

## Cambios

1. **Sin dependencias nuevas.** Se descarta Tailwind y cualquier DataGrid. Se usa **CSS Grid + Flexbox + container queries**, que cubren los mismos casos sin peso extra.
2. **Shell centrado y fluido.**
   - `--content-max: 1120px`; nav y contenido comparten `width: 100%`, `max-width` y `margin-inline: auto`.
   - Padding del contenido fluido con `clamp()`.
   - `min-width: 0` y `overflow-wrap` en textos largos (URLs, MCP).
3. **Resumen (dashboard).**
   - Contenedor `.dashboard` con `grid-template-columns: repeat(auto-fit, minmax(min(100%, 320px), 1fr))`.
   - 1 columna en compacto; 2 columnas cuando el ancho lo permite.
   - Las filas de consumo pasan a grid interno (`label | barra | %`) y usan **container query** cuando la card es muy estrecha (< 300 px): label y % arriba, barra a todo el ancho debajo.
4. **Servicio.** Dos cards en `.panel-grid` (control del servicio | configuración y MCP); 1 columna en compacto, 2 en ancho.
5. **Acceso web.** Cards para modo/URLs y para credenciales; formulario de contraseña con `flex-wrap` e input flexible.
6. **Ajustes.** `.settings-grid` con cards de idioma, general y notificaciones; 1/2/3 columnas según ancho.
7. **Modal.** Ancho máximo 560 px, alto acotado a `100dvh` y padding seguro; el formulario interno envuelve.
8. **Ventana.** `MinWidth: 340` y `MinHeight: 420` en `main.go` para impedir estados ilegibles; sigue siendo redimensionable sin máximo.
9. **Botones.** Se introduce la clase estándar `.btn` (30px / 12px / icono 14). El botón **Guardar** de Acceso web la usa, igual que el resto de acciones; en su formulario iguala la altura del input (`44px`) para que la fila quede alineada.

## Breakpoints de referencia

| Rango | Comportamiento |
|---|---|
| 340–599 px | 1 columna; nav envuelve; cards apiladas |
| 600–899 px | grids de 2 columnas cuando caben; contenido aún sin max-width efectivo |
| ≥ 900 px | contenido centrado con `max-width: 1120px`; Resumen y Servicio en 2 columnas |
| ≥ 1440 px | mismo layout centrado (sin estirar); 3 columnas solo si el grid lo permite (Ajustes) |
| Vertical alto | contenido alineado arriba y scroll en `.app-content`; sin estiramientos artificiales |

## Testing Strategy

- **Vitest** (estructura, no CSS): `Resumen` contiene `.dashboard`; `Servicio` contiene `.panel-grid`; `Ajustes` contiene `.settings-grid`; `AccesoWeb` usa cards; el modal mantiene su accesibilidad.
- **Manual**: redimensionar a 340×420, 380×520, 768×600, 1280×720, 1920×1080 y 2560×1440, además de 380×900 (vertical). Verificar: sin scroll horizontal, sin texto cortado, cards legibles.
- **Build**: `pnpm build`, `go test ./...` y `wails build` en verde.

## Boundaries

- **Always do:** CSS Grid/Flexbox, tokens del design system, `min-width: 0`, textos largos con wrap, a11y intacta.
- **Ask first:** dependencias nuevas (aquí, ninguna).
- **Never:** anchos fijos en contenedores principales, DataGrid para listas simples, ocultar contenido con `overflow: hidden` para “arreglar” el responsive.

## Success Criteria

- [ ] El contenido no se estira a todo el ancho en pantallas grandes; queda centrado con máximo 1120 px.
- [ ] Resumen muestra 1 columna en compacto y 2 en ancho.
- [ ] Servicio, Acceso web y Ajustes usan grids responsive (1/2/3 columnas según corresponda).
- [ ] Las barras de consumo se adaptan sin aplastarse (container query).
- [ ] No hay scroll horizontal en 340 px ni en 2560 px.
- [ ] La ventana no puede reducirse por debajo de 340×420.
- [ ] El botón **Guardar** de Acceso web usa la misma clase `.btn` que el resto de acciones y tiene la misma altura que el input de contraseña.
- [ ] Tests, `pnpm build` y `wails build` en verde.

---

# Iteración 4 — Carga asíncrona con skeletons

> Traza al module id `ui-ux`. Afecta a `resumen`, `service-control`, `web-access` y `shell` (Ajustes). Frontend, sin backend nuevo.

## Objective

Eliminar el bloqueo de render por una única carga global: cada widget debe mostrar su propio **skeleton** mientras espera sus datos y pintarse en cuanto estén listos, sin esperar al resto.

## Diagnóstico

- `Resumen` usaba un único `loading` y `Promise.all([ServiceStatus(), UsageFetch()])`: el widget de consumo (REST) no aparecía hasta que terminara el servicio (OpenCode), y los MCP esperaban a ambos.
- `Servicio`, `Acceso web` y `Ajustes` cargaban datos asíncronos sin indicador visual.

## Decisiones

1. **Skeleton, no spinner global.** El skeleton reserva el espacio del contenido final y evita saltos de layout; es preferible a un spinner centrado que sustituye toda la card.
2. **Componente reutilizable `Skeleton.vue`**: bloque con `width`/`height`/`radius`, decorativo (`aria-hidden`). El contenedor usa `role="status"`, `aria-live` implícito y un texto `.sr-only` con la clave i18n `common.loading`.
3. **Resumen con 3 cargas independientes**: `loadUsage`, `loadService` y `loadMcp` se lanzan en paralelo al montar. Cada bloque tiene su propio `loading` y `error`; si consumo termina antes, se pinta de inmediato aunque servicio/MCP sigan cargando.
4. **Servicio**: `loadStatus`, `loadMcp` y `loadJson` independientes y en paralelo. La lista MCP se pinta en cuanto llega, sin esperar al JSON; el skeleton del modal cubre el caso de abrirlo mientras el JSON carga.
5. **Acceso web**: `loadUrls` y `loadCredentials` independientes (usuario y contraseña se piden en paralelo); el cambio de contraseña solo recarga credenciales.
6. **Ajustes**: skeleton en la card General mientras se lee `GetSettings` (el resto de ajustes no depende del backend).
7. **Movimiento**: shimmer de 1.4s; `prefers-reduced-motion` ya lo desactiva globalmente.
8. **Sin dependencias nuevas.**

## Testing Strategy

- `Skeleton`: renderiza, es `aria-hidden` y aplica dimensiones.
- `Resumen`: con consumo resuelto y servicio pendiente, las barras se muestran y la card de servicio conserva su skeleton; al resolver el servicio, el skeleton desaparece.
- `Servicio`: muestra skeleton de estado mientras `ServiceStatus` está pendiente; los MCP se muestran aunque el JSON siga cargando.
- `Acceso web`: el botón Guardar usa `.btn`; muestra skeleton de URLs mientras carga.
- `Ajustes`: los switches aparecen tras `GetSettings` y el layout `.settings-grid` tiene 3 cards.
- `pnpm test`, `pnpm build`, `go test ./...` y `wails build` en verde.

## Success Criteria

- [ ] Consumo y Servicio se renderizan de forma independiente.
- [ ] Cada widget async muestra skeleton en lugar de bloquear la vista completa.
- [ ] Los errores se muestran en la card afectada, sin ocultar las demás.
- [ ] Los skeletons son accesibles (no anuncian contenido decorativo; sí el estado de carga).
- [ ] No hay dependencias nuevas.
