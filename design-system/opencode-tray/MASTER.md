# Design System Master File

> **LOGIC:** When building a specific page, first check `design-system/pages/[page-name].md`.
> If that file exists, its rules **override** this Master file.
> If not, strictly follow the rules below.

---

**Project:** OpenCode Tray
**Generated:** 2026-10-14 23:18:54
**Category:** Developer Tool / IDE

---

## Global Rules

### Color Palette

| Role | Hex | CSS Variable |
|------|-----|--------------|
| Primary | `#1E293B` | `--color-primary` |
| On Primary | `#FFFFFF` | `--color-on-primary` |
| Secondary | `#334155` | `--color-secondary` |
| On Secondary | `#FFFFFF` | `--color-on-secondary` |
| Accent/CTA | `#22C55E` | `--color-accent` |
| On Accent/CTA | `#0F172A` | `--color-on-accent` |
| Background | `#0F172A` | `--color-background` |
| Foreground | `#F8FAFC` | `--color-foreground` |
| Card | `#1B2336` | `--color-card` |
| Card Foreground | `#F8FAFC` | `--color-card-foreground` |
| Muted | `#272F42` | `--color-muted` |
| Muted Foreground | `#94A3B8` | `--color-muted-foreground` |
| Border | `#475569` | `--color-border` |
| Destructive | `#EF4444` | `--color-destructive` |
| On Destructive | `#000000` | `--color-on-destructive` |
| Ring | `#FFFFFF` | `--color-ring` |

**Color Notes:** Code dark + run green

### Typography

- **Heading Font:** JetBrains Mono
- **Body Font:** IBM Plex Sans
- **Mood:** code, developer, technical, precise, functional, hacker
- **Google Fonts:** [JetBrains Mono + IBM Plex Sans](https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500;600;700&display=swap)

**CSS Import:**
```css
@import url('https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500;600;700&display=swap');
```

### Spacing Variables

| Token | Value | Usage |
|-------|-------|-------|
| `--space-xs` | `4px` / `0.25rem` | Tight gaps |
| `--space-sm` | `8px` / `0.5rem` | Icon gaps, inline spacing |
| `--space-md` | `16px` / `1rem` | Standard padding |
| `--space-lg` | `24px` / `1.5rem` | Section padding |
| `--space-xl` | `32px` / `2rem` | Large gaps |
| `--space-2xl` | `48px` / `3rem` | Section margins |
| `--space-3xl` | `64px` / `4rem` | Hero padding |

### Shadow Depths

| Level | Value | Usage |
|-------|-------|-------|
| `--shadow-sm` | `0 1px 2px rgba(0,0,0,0.05)` | Subtle lift |
| `--shadow-md` | `0 4px 6px rgba(0,0,0,0.1)` | Cards, buttons |
| `--shadow-lg` | `0 10px 15px rgba(0,0,0,0.1)` | Modals, dropdowns |
| `--shadow-xl` | `0 20px 25px rgba(0,0,0,0.15)` | Hero images, featured cards |

---

## Component Specs

### Buttons

```css
/* Primary Button */
.btn-primary {
  background: #22C55E;
  color: #0F172A;
  padding: 12px 24px;
  border-radius: 8px;
  font-weight: 600;
  transition: all 200ms ease;
  cursor: pointer;
}

.btn-primary:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}

/* Secondary Button */
.btn-secondary {
  background: transparent;
  color: #F8FAFC;
  border: 2px solid #1E293B;
  padding: 12px 24px;
  border-radius: 8px;
  font-weight: 600;
  transition: all 200ms ease;
  cursor: pointer;
}
```

### Cards

```css
.card {
  background: #0F172A;
  border-radius: 12px;
  padding: 24px;
  box-shadow: var(--shadow-md);
  transition: all 200ms ease;
  cursor: pointer;
}

.card:hover {
  box-shadow: var(--shadow-lg);
  transform: translateY(-2px);
}
```

### Inputs

```css
.input {
  padding: 12px 16px;
  border: 1px solid #E2E8F0;
  border-radius: 8px;
  font-size: 16px;
  transition: border-color 200ms ease;
}

.input:focus {
  border-color: #1E293B;
  outline: none;
  box-shadow: 0 0 0 3px #1E293B20;
}
```

### Modals

```css
.modal-overlay {
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
}

.modal {
  background: white;
  border-radius: 16px;
  padding: 32px;
  box-shadow: var(--shadow-xl);
  max-width: 500px;
  width: 90%;
}
```

### Standard Action Button (`.btn`)

Todos los botones de acción de la app comparten una única clase, tamaño `sm`, texto e icono de 14px:

```css
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-xs);
  min-height: 30px;
  padding: var(--space-xs) var(--space-sm);
  background: var(--panel);
  color: var(--text);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font: inherit;
  font-size: 12px;
  cursor: pointer;
}
```

No usar estilos nativos del navegador para acciones (p. ej. el botón Guardar de Acceso web). Cuando un `.btn` comparte fila con un input, debe igualar su altura (`min-height: var(--target)`) para mantener la alineación.

### Flyout de Alto Fijo

- El panel mantiene **380×600 px** en todas las pestañas; el alto no depende del contenido.
- El frontend no mide ni solicita cambios de alto al cambiar de sección.
- Solo `.app-content` hace scroll; la nav inferior permanece fija.
- Estado y URI comparten `.status-row` tanto en Resumen como en Servicio, con la URI alineada a la derecha.

### Compact Action Grid

Las filas de acciones densas usan etiquetas cortas y una rejilla explícita para evitar saltos internos:

- **Servicio:** 2×2 (`Iniciar`, `Detener`, `Forzar`, `Refrescar`).
- **Configuración:** 3 columnas (`Editor`, `Carpeta`, `JSON`).
- Cada `.btn` de la rejilla usa `width: 100%`, `min-width: 0` y `white-space: nowrap`.
- Los textos completos viven en el tooltip nativo (`title`).

### Tooltips

- Todo botón solo-icono debe llevar `aria-label` y `title` con la misma acción.
- Los botones con etiqueta corta llevan `title` con la descripción completa.
- Se usa el tooltip nativo; no se añade componente ni dependencia.

### Scroll Estable

- `.app-content` usa `scrollbar-gutter: stable` para reservar siempre el ancho de la barra.
- La barra es fina y discreta (`scrollbar-width: thin` + `::-webkit-scrollbar`), visible solo al hacer scroll.
- `overflow-x: hidden`; el contenido nunca debe generar desplazamiento horizontal.
- Prohibido depender de `overflow-y: scroll` para evitar el reflujo: la barra permanente empeora la percepción del panel.

### Responsive Layout

El panel se adapta de ~340 px a monitores grandes sin estirar widgets:

- **Contenedor:** `max-width: 1120px`, centrado con `margin-inline: auto`; padding fluido con `clamp(14px, 2.5vw, 24px)`.
- **Grids:** `repeat(auto-fit, minmax(min(100%, 320px), 1fr))` para Resumen, Servicio, Acceso web y Ajustes.
- **Container queries:** las cards son contenedores (`container-type: inline-size`); el contenido interno se reorganiza cuando la card mide menos de 300 px.
- **Texto:** `min-width: 0` y `overflow-wrap: anywhere` en URLs, nombres MCP y títulos.
- **Ventana:** panel fijo 380×600; el contenido se adapta sin scroll horizontal.
- **No usar DataGrid** para listas simples ni Tailwind como dependencia.

### Loading Skeletons

Cada bloque asíncrono tiene su propio estado de carga (no un loading global):

```vue
<div v-if="loading" class="skeleton-stack" role="status">
  <span class="sr-only">{{ t('common.loading') }}</span>
  <Skeleton width="80%" height="14px" />
</div>
```

- `Skeleton` es decorativo (`aria-hidden`); el estado se anuncia con `role="status"` + `.sr-only`.
- Shimmer de 1.4 s, desactivado por `prefers-reduced-motion`.
- Mostrar skeleton solo en el widget que carga; el resto se renderiza en cuanto esté listo.
- En error, mostrar el mensaje en la card afectada y conservar las demás.

---

## Style Guidelines

**Style:** Dark Mode (OLED)

**Keywords:** Dark theme, low light, high contrast, deep black, midnight blue, eye-friendly, OLED, night mode, power efficient

**Best For:** Night-mode apps, coding platforms, entertainment, eye-strain prevention, OLED devices, low-light

**Key Effects:** Minimal glow (text-shadow: 0 0 10px), dark-to-light transitions, low white emission, high readability, visible focus

### Page Pattern

**Pattern Name:** FAQ/Documentation Landing

- **Conversion Strategy:** Reduce support tickets. Track search analytics. Show related articles. Contact escalation path.
- **CTA Placement:** Search bar prominent + Contact CTA for unresolved questions
- **Section Order:** Hero with search bar > Popular categories > FAQ accordion > Contact/support CTA

---

## Anti-Patterns (Do NOT Use)

- ❌ Light mode default
- ❌ Slow performance

### Additional Forbidden Patterns

- ❌ **Emojis as icons** — Use SVG icons (Heroicons, Lucide, Simple Icons)
- ❌ **Missing cursor:pointer** — All clickable elements must have cursor:pointer
- ❌ **Layout-shifting hovers** — Avoid scale transforms that shift layout
- ❌ **Low contrast text** — Maintain 4.5:1 minimum contrast ratio
- ❌ **Instant state changes** — Always use transitions (150-300ms)
- ❌ **Invisible focus states** — Focus states must be visible for a11y

---

## Pre-Delivery Checklist

Before delivering any UI code, verify:

- [ ] No emojis used as icons (use SVG instead)
- [ ] All icons from consistent icon set (Heroicons/Lucide)
- [ ] `cursor-pointer` on all clickable elements
- [ ] Hover states with smooth transitions (150-300ms)
- [ ] Light mode: text contrast 4.5:1 minimum
- [ ] Focus states visible for keyboard navigation
- [ ] `prefers-reduced-motion` respected
- [ ] Responsive: 375px, 768px, 1024px, 1440px
- [ ] No content hidden behind fixed navbars
- [ ] No horizontal scroll on mobile
