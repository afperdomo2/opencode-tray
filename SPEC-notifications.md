# Spec: notifications (OpenCode Tray)

> Traza al module id `notifications` del `CAPABILITY-MAP.md`. Depende de `shell`, `usage` y `service-control`. Último módulo en construirse (reacciona a los demás).

## Objective

`notifications` emite **toasts nativos** ante dos situaciones de consumo (definidas en el mapa) y, opcionalmente, ante cambios de estado del servicio. Todo es configurable y **sin spam** (dedupe por período).

**Tipo A — Ritmo ("burn rate"):** preventivo. Evalúa si consumes más rápido que un ritmo sostenible. Se dispara en **hitos de consumo 20 / 40 / 60 / 80 %** cuando `consumido − esperado ≥ margen` (margen por defecto **15 pts**). `esperado` = % de tiempo transcurrido de la ventana (lineal). Severidad **amarillo**.

> El ritmo solo aplica a ventanas de período fijo (**weekly / monthly**). La ventana **rolling (5 h)** es demasiado corta para "ritmo" y solo recibe alertas de agotamiento.

**Tipo B — Agotamiento:** reactivo. Umbrales fijos:
- **90 %** → naranja, "casi al límite".
- **98 %** → rojo, "al límite, cambia ya".
- **100 %** → crítico, "se acabó el saldo".

Aplica a las tres ventanas (rolling / weekly / monthly).

**Tipo C — Estado del servicio:** notificar cuando el servicio pasa de "corriendo" a "detenido" **sin acción del usuario** (detenido inesperadamente). No notifica starts/stops manuales.

**Contrato público:**

```go
type Level int // Info | Warn | Critical

func (s *Service) Evaluate(u usage.Usage, svc servicecontrol.Status) error
func (s *Service) Notify(title, msg string, lvl Level) error
func (s *Service) SetEnabled(kind string, on bool) error
func (s *Service) SetThresholds(t Thresholds) error
```

**Dedupe:** se persiste qué hitos/umbrales ya se notificaron por ventana y período; se reinicia cuando pasa `resetsAt`. Nunca se repite la misma alerta en refrescos consecutivos.

## Tech Stack

- Go 1.24+.
- Notificaciones: **`github.com/gen2brain/beeep`** (cross-platform: toast en Windows, `notify-send` en Linux — facilita la fase Linux).
- Lógica de ritmo/umbral pura (sin I/O) para ser 100 % testeable.

## Project Structure

```
internal/
└── notifications/
    ├── notifications.go  # struct Service + New()
    ├── burnrate.go       # lógica de ritmo (weekly/monthly)
    ├── thresholds.go     # agotamiento 90/98/100
    ├── servicewatch.go   # (opcional) estado del servicio
    ├── dedupe.go         # estado notificado por ventana/período
    ├── notify.go         # envío via beeep (inyectable para tests)
    └── *_test.go
```

## Code Style

- Lógica pura y determinista, separada del envío (el sender se inyecta):

```go
// burnrate.go
func burnRateAlerts(w Window, expectedPct float64, margin float64, milestones []float64) []Alert {
    var out []Alert
    for _, m := range milestones {
        if w.Percent >= m && w.Percent-expectedPct >= margin {
            out = append(out, Alert{Kind: "burnrate", Milestone: m})
        }
    }
    return out
}
```

- `expectedPct` se calcula fuera (inyectable) para tests deterministas.

## Testing Strategy

- **Go `testing`** (lógica pura, sin I/O):
  - `burnRateAlerts` dispara en el hito correcto solo si supera el margen; no dispara si va al ritmo.
  - Agotamiento: 90/98/100 disparan al cruzar; no antes.
  - Dedupe: no repite la misma alerta en la misma ventana/período; se reinicia tras `resetsAt`.
  - Ritmo **no** aplica a rolling; agotamiento sí a las tres.
  - Sender mockeado (beeep no se invoca en tests).
- **Vue (Vitest)**: panel de Ajustes de notificaciones (toggle por tipo, umbrales).

## Boundaries

- **Always do:**
  - Dedupe por ventana/período (nunca spam).
  - Nunca incluir token/contraseña en el texto de una notificación.
  - Lógica pura testeada; envío inyectable.
- **Ask first:**
  - Cambiar la librería de notificaciones o añadir sonido/acción de clic.
  - Añadir "No molestar" o programación de silencio.
- **Never:**
  - Disparar la misma alerta repetidamente en un mismo período.
  - Notificar información sensible.

## Success Criteria

- [ ] Alerta de ritmo: dispara en 20/40/60/80 % cuando `consumido − esperado ≥ 15`, una vez por hito y período, solo en weekly/monthly.
- [ ] Alerta de agotamiento: 90 % (naranja), 98 % (rojo), 100 % (crítico) en las tres ventanas.
- [ ] Sin notificaciones duplicadas entre refrescos consecutivos.
- [ ] Umbrales y tipos configurables desde Ajustes.
- [ ] Tests Go y Vitest en verde.

## Decisiones cerradas

1. **Tipo C incluido**: notificar solo "detenido inesperadamente" (de corriendo → detenido sin acción del usuario). Se mantiene la dependencia de `service-control`.
2. **Librería**: `beeep` (cross-platform).
3. **Ritmo solo en weekly/monthly**; rolling solo agotamiento.
