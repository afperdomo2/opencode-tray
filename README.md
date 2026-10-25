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
powershell -ExecutionPolicy Bypass -File scripts/build-installer.ps1  # portable + instalador Inno Setup
```

## Funcionalidad

- **Panel (flyout)**: ventana frameless anclada al icono del tray, sin marco, sin resize ni arrastre. Se abre con clic en el icono, se cierra con `Esc` o al perder el foco y mantiene un alto fijo de 600 px en todas las pestañas. No aparece en la barra de tareas ni en Alt+Tab. Las pestañas tienen tamaño fijo y el scroll reserva su espacio para no deformar el contenido. Al pasar el cursor por el icono, el tooltip nativo muestra el **consumo mensual con barra** y el **reinicio humanizado** (`OpenCode · Mensual ▰▰▱▱▱ 42% · 🕒 en 4 h`), y se actualiza aunque el panel nunca se abra. Detalle en [`specs/SPEC-shell.md`](specs/SPEC-shell.md).
- **Resumen**: dashboard de inicio en formato **widgets (bento)**: consumo con las tres barras, servicio con estado/URL, acceso web con icono y color por modo (Local/Red/Remoto) y MCP con chips. Servicio, Acceso y MCP son **clicables** para saltar a su sección; Consumo mantiene los CTAs de token. Todo es de solo lectura salvo reintentar el token o configurar la API key.
- **Servicio** (unifica el antiguo Config): estado + URI con puerto y control con etiquetas cortas (iniciar / detener / forzar / refrescar) + servidores MCP + configuración (editor / carpeta / **JSON en modal**). Los botones muestran su descripción completa al pasar el cursor.
- **Acceso web**: modos local / red (por IP) / remoto (túnel), exclusivos: al volver a local/red se **desactiva el túnel remoto**. Muestra las URLs reales con su tipo (Local / LAN / Remoto), botón para **abrir en el navegador** y botón para **copiar** con confirmación. El modo remoto pide **confirmación explícita** (expone el servidor a Internet). Incluye usuario (`opencode`), contraseña enmascarada, cambio de contraseña y un botón para **refrescar** el estado y las URLs. En modo local la URL sale de `opencode service status` (puerto real); en red/remoto se obtienen de `/api/info` las IPs LAN reales y el enlace del túnel.
- **Ajustes**: idioma (es/en), auto-inicio con Windows, switches de notificaciones, tarjeta de **OpenCode Go** (estado del token, reintento desde OpenCode, API key manual cifrada con DPAPI) e intervalo de refresco (1/5/15/30/60 min).

La UI usa iconos SVG (sin emoji), botones `btn` en tamaño compacto estándar, control segmentado de idioma, switches estilizados y cards con profundidad (design system). El layout es responsive con CSS Grid/Flexbox y container queries, sin dependencias de UI. La navegación es un footer fijo de 4 pestañas (icono + texto) y el contenido es la única zona con scroll. Los datos de Servicio, MCP, Acceso web y Consumo comparten un **caché en memoria** con claves comunes: volver a una pestaña con datos frescos (10–15 s) no repite la petición ni muestra skeleton. El **skeleton** aparece solo en la primera carga, en el refresco manual y desde el inicio de una mutación (iniciar/detener el servicio, cambiar modo o contraseña). Las revalidaciones en segundo plano y el polling silencioso del servicio (cada 15 s con el panel visible) sustituyen los datos sin skeleton ni texto de “Actualizando…”. Un cambio hecho en Servicio/Acceso invalida el caché compartido, así que Resumen refleja el nuevo estado (con skeleton mientras llega). La lista de URLs usa 2 filas esqueleto con la misma altura que las filas reales.

## Build y despliegue (Windows)

### Requisitos de build

- Go 1.24+ y Wails CLI v2.16+.
- `pnpm` para el frontend.
- **Inno Setup 6.3+** solo si vas a generar el instalador.
- WebView2 Runtime (Windows 10/11 moderno ya lo incluye).

### Instalar Inno Setup

Con winget:

```powershell
winget install --id JRSoftware.InnoSetup -e
```

Con Chocolatey:

```powershell
choco install innosetup -y
```

O descargando el instalador desde <https://jrsoftware.org/isinfo.php>.

Comprueba que `ISCC.exe` está disponible:

```powershell
iscc /?
```

Si no se reconoce, añade al `PATH` una de estas carpetas (según la instalación):

```text
C:\Program Files (x86)\Inno Setup 6
%LOCALAPPDATA%\Programs\Inno Setup 6
```

El script también busca `ISCC.exe` en esas rutas aunque no esté en el `PATH`.

### Compilar

```powershell
# 1) .exe portable
wails build -clean -webview2 download

# 2) portable + instalador per-user
powershell -ExecutionPolicy Bypass -File scripts\build-installer.ps1

# 3) forzando una versión concreta
powershell -ExecutionPolicy Bypass -File scripts\build-installer.ps1 -Version 1.3.0
```

El script:

1. Lee `Info.productVersion` de `wails.json` (o usa `-Version`).
2. Avisa si `opencode-tray.exe` está en ejecución.
3. Ejecuta `wails build -clean -webview2 download`.
4. Ejecuta `ISCC.exe` con `build\windows\installer\opencode-tray.iss`.

También puedes compilar el instalador a mano:

```powershell
iscc /DMyAppVersion=1.3.0 build\windows\installer\opencode-tray.iss
```

La versión del `.exe` sale de `Info.productVersion` en `wails.json`; mantén ese valor y `-Version` sincronizados.

### Artefactos

```text
build/bin/opencode-tray.exe
build/bin/OpenCodeTray-Setup-<versión>.exe
```

Ambos están bajo `build/bin/`, que está en `.gitignore`: **no se commitean**. Sí se commitean la definición `build/windows/installer/opencode-tray.iss` y el script `scripts/build-installer.ps1`.

### Instalación

- Doble clic en `OpenCodeTray-Setup-<versión>.exe`.
- Instalación **por usuario**, sin UAC ni permisos de administrador.
- Directorio: `%LOCALAPPDATA%\Programs\OpenCode Tray`.
- Si falta WebView2, ofrece descargar e instalar el bootstrapper oficial de Microsoft (requiere Internet; si falla, muestra el enlace manual).
- Si `opencode` no está en el `PATH`, avisa pero **no bloquea** la instalación.
- Accesos directos opcionales (escritorio y ejecutar al terminar).
- El instalador **no** crea una entrada de auto-inicio: eso lo controla la app desde Ajustes (`HKCU\...\Run`).
- Los ajustes (`%APPDATA%\OpenCodeTray\settings.json`) se conservan al desinstalar.

### Desinstalar

Desde **Configuración → Aplicaciones → OpenCode Tray**, o ejecutando:

```text
%LOCALAPPDATA%\Programs\OpenCode Tray\unins000.exe
```

### Firma de código (opcional)

Sin certificado, Windows SmartScreen puede mostrar una advertencia. Si tienes certificado, firma el `.exe` portable y el instalador con `signtool` antes de publicarlos. Inno Setup permite integrar la firma con `SignTool` en el `.iss` o pasarla por línea de comandos con `/S<nombre>=<comando>`.

### Checklist de release

- [ ] `go test ./...`, `pnpm test` y `pnpm build` en verde.
- [ ] `Info.productVersion` en `wails.json` actualizado y `RELEASE.md` sincronizado.
- [ ] `wails build -clean -webview2 download` genera el portable.
- [ ] `scripts/build-installer.ps1` genera el instalador.
- [ ] Probar en una VM Windows limpia, sin Go/Wails instalados.
- [ ] Probar instalación con y sin WebView2, y con y sin `opencode` en PATH.
- [ ] Verificar accesos directos, ejecución, desinstalación y conservación de ajustes.
- [ ] Publicar el portable y el instalador (nunca `build/bin` dentro de git).

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
