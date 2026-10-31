# OpenCode Tray

App de bandeja (tray) para **Windows** y **Linux** que controla el servicio de OpenCode, gestiona el acceso web (local / red / remoto), muestra el consumo de OpenCode Go y permite ver la configuración (MCP / skills / agents).

> **Estado:** funcional — todas las capacidades están implementadas y con tests.
> **Plataformas:** 🪟 Windows (distribución oficial: `.exe` + instalador) · 🐧 Ubuntu/Linux (desarrollo y prueba).
> **Planificación:** [`specs/`](specs/) (el *qué*) y [`tasks/`](tasks/) (el *cómo*).

Stack: **Wails v2.16 + Go 1.25+ + Vue 3 / Vite / TypeScript**.

## 📑 Contenido

- [Antes de empezar](#antes-de-empezar)
- [Inicio rápido](#inicio-rapido) — Ubuntu y Windows
- [Probar la app](#probar) — abrir el panel, checklist y problemas frecuentes
- [Funcionalidad](#funcionalidad) — qué hace cada pestaña
- [Desarrollo](#desarrollo) — dev, tests y estructura
- [Build — Windows](#build-windows) — `.exe`, instalador y release
- [Build — Linux/Ubuntu](#build-linux) — requisitos, compilar y ejecutar
- [Limitaciones conocidas](#limitaciones)
- [Documentación y planificación](#docs)

---

<a id="antes-de-empezar"></a>
## 🧭 Antes de empezar

| Necesitas | Windows | Ubuntu / Linux |
|---|---|---|
| Go | 1.25+ | 1.25+ |
| Wails CLI | v2.16+ | v2.16+ |
| Node / pnpm | Node 20+ y pnpm 10+ | Node 20+ y pnpm 10+ |
| Librerías del SO | WebView2 (viene con Win10/11) | GTK3 + **WebKitGTK 4.1** (ver build) |
| CLI de OpenCode | `opencode` en el `PATH` | `opencode` en el `PATH` |

Si `wails` no se reconoce, instálalo y añádelo al `PATH`:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0
export PATH="$HOME/go/bin:$PATH"        # queda en ~/go/bin
```

---

<a id="inicio-rapido"></a>
## 🚀 Inicio rápido

**¿Qué quieres hacer?**

| Objetivo | Cómo |
|---|---|
| Compilar y probar en **Ubuntu** (primera vez) | Pasos 1–4 de abajo |
| Recompilar tras un cambio | `./scripts/build-linux.sh` |
| **Instalar** como app en Ubuntu (`.deb`) | `./scripts/build-deb.sh` y luego `sudo apt install ./dist/*.deb` |
| Compilar el **`.exe`** en Windows | `wails build -clean -webview2 download` |
| Correr solo los tests | `go test ./...` y `cd frontend && pnpm test` |
| Desarrollo con hot reload | `wails dev` (lee las [limitaciones](#limitaciones)) |

### 🐧 Ubuntu / Linux

```bash
# 1) Dependencias del sistema (una sola vez)
sudo apt install -y build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev

# 2) Wails CLI (una sola vez)
go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0
export PATH="$HOME/go/bin:$PATH"

# 3) Compilar (genera build/bin/opencode-tray)
cd ~/www/opencode-tray
./scripts/build-linux.sh

# 4) Ejecutar (arranca solo con el icono del tray)
./build/bin/opencode-tray
```

El script detecta WebKitGTK 4.1 y aplica `-tags webkit2_41` por ti (Ubuntu 24.04+ ya no trae la 4.0) y fuerza `CGO_ENABLED=1`. También avisa si faltan las librerías de desarrollo.

### 🪟 Windows

```powershell
# 1) Compilar el portable
wails build -clean -webview2 download

# 2) Ejecutar
.\build\bin\opencode-tray.exe
```

Con instalador per-user: `powershell -ExecutionPolicy Bypass -File scripts\build-installer.ps1` (detalle en [Build — Windows](#build-windows)).

---

<a id="probar"></a>
## 🧪 Probar la app

### Cómo abrir el panel

- **Windows:** clic izquierdo en el icono de la bandeja (clic derecho = menú).
- **Ubuntu/GNOME:** abre el menú del icono y elige **Mostrar panel**. GNOME no siempre emite el clic izquierdo (`Activate`), por eso el panel también está en el menú. Si instalaste el `.deb`, también puedes lanzarla desde el **menú de aplicaciones** (abre el panel al arrancar).
- Se cierra con `Esc`, al perder el foco o con un segundo clic. **Cerrar = ocultar al tray**; para salir de verdad usa **Salir** en el menú.

### Checklist de humo (2 minutos)

| # | Comprueba | Resultado esperado |
|---|---|---|
| 1 | Icono en la bandeja | Aparece (PNG en Linux, ICO en Windows) |
| 2 | Consumo del mes | Tooltip en Windows/KDE; en GNOME, primera fila del menú (`OpenCode · Mensual ▰▰▱▱▱ 42% · 🕒 en 4 h`) |
| 3 | Abrir el panel | Resumen con 4 widgets: consumo, servicio, acceso web y MCP |
| 4 | Pestaña **Servicio** | Estado + URI real y botones iniciar / detener / forzar / refrescar |
| 5 | Pestaña **Acceso web** | URLs Local / LAN / Remoto con abrir y copiar |
| 6 | Pestaña **Ajustes** | Idioma es/en y el panel cambia de idioma al instante |

Si el tooltip muestra consumo, el token de OpenCode Go y la red están bien.

### Log y cierre

```bash
# Linux: lanzar en segundo plano y ver el log
./build/bin/opencode-tray > /tmp/opencode-tray.log 2>&1 &

# Cerrar
pkill -x opencode-tray
```

En Windows el `.exe` no abre consola; si necesitas ver errores, usa `wails dev` (sabiendo su limitación) o ejecuta el `.exe` desde una terminal.

### Problemas frecuentes

| Síntoma | Causa | Solución |
|---|---|---|
| `wails: command not found` | `~/go/bin` fuera del `PATH` | `export PATH="$HOME/go/bin:$PATH"` |
| `Package webkit2gtk-4.1 was not found` | faltan los headers | `sudo apt install libwebkit2gtk-4.1-dev` |
| Errores de cgo / `CGO_ENABLED=0` | compilación sin cgo | usa `./scripts/build-linux.sh` (lo fuerza) |
| El icono no aparece en GNOME | extensión AppIndicator desactivada | actívala en GNOME Tweaks o usa el menú del tray |
| El clic izquierdo no abre el panel | GNOME no emite `Activate` | usa **Mostrar panel** |
| El panel aparece en otra esquina | Wayland ignora la posición de la ventana | es esperado; en X11/XWayland se respeta |
| Las secciones dicen “Detenido” con el servicio activo | estás en `wails dev` | prueba con el build de producción |

---

<a id="funcionalidad"></a>
## ✨ Funcionalidad

- **Panel (flyout)**: ventana frameless anclada al icono del tray, sin marco, sin resize ni arrastre. Se abre desde el icono, se cierra con `Esc` o al perder el foco y mantiene un alto fijo de 600 px en todas las pestañas. No aparece en la barra de tareas ni en Alt+Tab. Las pestañas tienen tamaño fijo y el scroll reserva su espacio para no deformar el contenido. Al pasar el cursor por el icono, el tooltip nativo muestra el **consumo mensual con barra** y el **reinicio humanizado**, y se actualiza aunque el panel nunca se abra. Detalle en [`specs/SPEC-shell.md`](specs/SPEC-shell.md).
- **Resumen**: dashboard de inicio en formato **widgets (bento)**: consumo con las tres barras, servicio con estado/URL, acceso web con icono y color por modo (Local/Red/Remoto) y MCP con chips. Servicio, Acceso y MCP son **clicables** para saltar a su sección; Consumo mantiene los CTAs de token. Todo es de solo lectura salvo reintentar el token o configurar la API key.
- **Servicio** (unifica el antiguo Config): estado + URI con puerto y control con etiquetas cortas (iniciar / detener / forzar / refrescar) + servidores MCP + configuración (editor / carpeta / **JSON en modal**). Los botones muestran su descripción completa al pasar el cursor.
- **Acceso web**: modos local / red (por IP) / remoto (túnel), exclusivos: al volver a local/red se **desactiva el túnel remoto**. Muestra las URLs reales con su tipo (Local / LAN / Remoto), botón para **abrir en el navegador** y botón para **copiar** con confirmación. El modo remoto pide **confirmación explícita** (expone el servidor a Internet). Incluye usuario (`opencode`), contraseña enmascarada, cambio de contraseña y un botón para **refrescar** el estado y las URLs. En modo local la URL sale de `opencode service status` (puerto real); en red/remoto se obtienen de `/api/info` las IPs LAN reales y el enlace del túnel.
- **Ajustes**: idioma (es/en), auto-inicio con el sistema (registro en Windows, XDG Autostart en Linux), switches de notificaciones, tarjeta de **OpenCode Go** (estado del token, reintento desde OpenCode, API key manual cifrada con DPAPI en Windows y base64 en Linux) e intervalo de refresco (1/5/15/30/60 min).

La UI usa iconos SVG (sin emoji), botones `btn` en tamaño compacto estándar, control segmentado de idioma, switches estilizados y cards con profundidad (design system). El layout es responsive con CSS Grid/Flexbox y container queries, sin dependencias de UI. La navegación es un footer fijo de 4 pestañas (icono + texto) y el contenido es la única zona con scroll. Los datos de Servicio, MCP, Acceso web y Consumo comparten un **caché en memoria** con claves comunes: volver a una pestaña con datos frescos (10–15 s) no repite la petición ni muestra skeleton. El **skeleton** aparece solo en la primera carga, en el refresco manual y desde el inicio de una mutación (iniciar/detener el servicio, cambiar modo o contraseña). Las revalidaciones en segundo plano y el polling silencioso del servicio (cada 15 s con el panel visible) sustituyen los datos sin skeleton ni texto de “Actualizando…”. Un cambio hecho en Servicio/Acceso invalida el caché compartido, así que Resumen refleja el nuevo estado (con skeleton mientras llega). La lista de URLs usa 2 filas esqueleto con la misma altura que las filas reales.

---

<a id="desarrollo"></a>
## 🛠️ Desarrollo

```bash
wails dev                  # hot reload (lee la limitación del puente más abajo)
go test ./...              # tests backend (Go)
cd frontend && pnpm test    # tests frontend (Vitest): 69 tests
cd frontend && pnpm build   # type-check (vue-tsc) + build del frontend
```

- `gofmt` y los tests son obligatorios antes de cerrar un cambio (ver [`AGENTS.md`](AGENTS.md)).
- **No valides frontend ↔ backend con `wails dev`**: en modo dev el puente `window.go` puede quedar sin definir y todas las llamadas al backend fallan (`TypeError: Cannot read properties of undefined (reading 'main')`). Es un quirk del proxy de Vite de Wails, no un bug de la app. Para probar el comportamiento real usa el build de producción.

### Estructura del repo

| Ruta | Qué hay |
|---|---|
| `main.go` / `app.go` | Entry point de Wails, tray, wiring de módulos y ventana |
| `internal/shell` | Tray, ciclo de vida del panel, ajustes, auto-inicio, i18n |
| `internal/servicecontrol` | Estado y control del servicio de OpenCode |
| `internal/webaccess` | Modos de acceso, URLs y credenciales |
| `internal/usage` | Consumo de OpenCode Go, token y tooltip |
| `internal/configviewer` | MCP / skills / agents y JSON de configuración |
| `internal/notifications` | Lógica de alertas (envío deshabilitado) |
| `frontend/` | UI Vue 3 (`components/` por sección, `lib/` caché y ajustes, i18n es/en) |
| `build/` | Iconos y recursos por plataforma (`build/bin` es gitignored) |
| `scripts/` | `build-linux.sh`, `build-deb.sh`, `build-installer.ps1` (Windows) |
| `internal/cli` | Resolución del CLI de OpenCode (PATH, `~/.opencode/bin`, etc.) |

---

<a id="build-windows"></a>
## 🪟 Build — Windows

### Requisitos

- Go 1.25+ y Wails CLI v2.16+.
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

Comprueba que `ISCC.exe` está disponible con `iscc /?`. Si no se reconoce, añade al `PATH` una de estas carpetas:

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

---

<a id="build-linux"></a>
## 🐧 Build — Linux/Ubuntu

> Se puede **compilar y ejecutar sin instalar** o generar un **`.deb`** para usarla como cualquier app del escritorio. La distribución oficial sigue siendo el `.exe` de Windows.

### Requisitos (una sola vez)

```bash
sudo apt install -y build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev
go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0
export PATH="$HOME/go/bin:$PATH"
```

- `CGO_ENABLED=1` (Wails usa cgo en Linux; el script lo fuerza).
- **WebKitGTK 4.1** (`webkit2_41`): Ubuntu 24.04+ ya no empaqueta WebKitGTK 4.0.

### Compilar y ejecutar

```bash
./scripts/build-linux.sh            # wails build -clean -tags webkit2_41
./build/bin/opencode-tray           # arranca con el icono del tray
```

El binario queda en `build/bin/opencode-tray` (gitignored). Para recompilar tras un cambio basta con repetir el script (build incremental ~10 s).

### Instalar como aplicación (`.deb`)

```bash
./scripts/build-deb.sh                                   # → dist/opencode-tray_<versión>_amd64.deb
sudo apt install ./dist/opencode-tray_1.3.0_amd64.deb
```

El paquete instala tres cosas:

| Ruta | Para qué |
|---|---|
| `/usr/bin/opencode-tray` | Ejecutable |
| `/usr/share/applications/opencode-tray.desktop` | **Icono en el menú de aplicaciones** (buscable como “OpenCode Tray”) |
| `/usr/share/icons/hicolor/256x256/apps/opencode-tray.png` | Icono del menú y del dock |

- Se lanza **desde el menú de aplicaciones** o con `opencode-tray` en la terminal; no hace falta abrir una terminal para usarla.
- El lanzador del menú abre el **panel directamente** (`--show-panel`); si la app ya está corriendo, la segunda ejecución pide al proceso vivo que muestre el panel y termina (instancia única).
- El **auto-inicio** (Ajustes → “Iniciar con el sistema”) arranca oculta, solo con el icono del tray.
- **No necesita symlink ni tocar el `PATH`**: la app resuelve el CLI de OpenCode sola (`$OPENCODE_BIN`, `PATH`, `~/.opencode/bin`, `~/.local/bin`, `/usr/local/bin`, `/usr/bin`, `/snap/bin`).
- La versión del paquete sale de `Info.productVersion` en `wails.json`.

Desinstalar (los ajustes de `~/.config/OpenCodeTray` se conservan):

```bash
sudo apt remove opencode-tray
```

### Diferencias respecto a Windows

- **Icono de bandeja**: se embebe `build/linux/icon.png` (64 px) porque la bandeja decodifica PNG/JPEG y no ICO. Windows sigue usando `build/windows/icon.ico`.
- **Panel**: aparece **centrado** en el monitor principal y se puede **arrastrar** desde la franja superior (en Windows va anclado al icono del tray y no se mueve). Se centra solo la primera vez: conserva la posición que elijas. Bajo Wayland el compositor decide la posición inicial.
- **Arranque oculto**: Wails ignora `StartHidden` en Linux; la app oculta la ventana en `startup` para arrancar solo con el icono.
- **Clic en el icono**: en GNOME/AppIndicator el clic abre el menú (el escritorio nunca emite `Activate`), así que el panel se abre con **Mostrar panel**, la primera entrada del menú. En KDE/XFCE el clic izquierdo sí abre el panel.
- **Auto-inicio**: crea `~/.config/autostart/opencode-tray.desktop` (XDG).
- **Tooltip**: GNOME no dibuja el tooltip del icono; por eso la primera fila del menú muestra el consumo y el reinicio (se actualiza solo). En Windows/KDE el tooltip nativo sigue funcionando.
- **Secretos**: sin DPAPI, el token manual de OpenCode Go se guarda en **base64** (no cifrado). No lo uses en máquinas compartidas.
- **Instancia única**: implementada con `flock` + socket en `$XDG_RUNTIME_DIR`; una segunda ejecución pide mostrar el panel y sale.
- **Notificaciones**: dependen del daemon de notificaciones del escritorio (el envío está deshabilitado en runtime, igual que en Windows).

---

<a id="limitaciones"></a>
## ⚠️ Limitaciones conocidas

- **Flyout**: no hay preview visual al pasar el cursor (solo tooltip nativo) y no se puede redimensionar. En Windows tampoco se arrastra (va anclado al tray); en Linux sí, desde la franja superior.
- **Notificaciones**: el **envío** de toasts está deshabilitado en runtime — el mecanismo de Wails entraba en conflicto con el bucle de la bandeja y la dejaba muda. La lógica (ritmo, umbrales, servicio detenido) está implementada y probada; detalle y cómo reactivarlo en [`specs/SPEC-notifications.md`](specs/SPEC-notifications.md).
- **Fuentes web**: la UI carga IBM Plex Sans + JetBrains Mono desde Google Fonts (`display=swap`); sin red degrada a las fuentes del sistema y la app sigue siendo funcional.
- **Linux**: sin anclaje real al icono (esquina inferior derecha; Wayland puede ignorarla) y sin cifrado DPAPI para el token manual. Detalle en [`specs/SPEC-shell.md`](specs/SPEC-shell.md) → Iteraciones 10 y 11.
- **`wails dev`**: el puente frontend ↔ backend puede fallar en modo dev (ver [Desarrollo](#desarrollo)). Para validar comportamiento real usa siempre el build de producción:

  ```bash
  ./scripts/build-linux.sh    # Linux → build/bin/opencode-tray
  wails build                 # Windows → build/bin/opencode-tray.exe
  ```

---

<a id="docs"></a>
## 📚 Documentación y planificación

El proyecto sigue **Spec-Driven Development (SDD)**: primero las specs, luego el plan y las tareas, y solo después la implementación.

| Documento | Contenido |
|---|---|
| [`specs/CAPABILITY-MAP.md`](specs/CAPABILITY-MAP.md) | Mapa de módulos y orden de construcción |
| [`specs/SPEC-*.md`](specs/) | Una spec por módulo (fuente de verdad del *qué*) |
| [`tasks/plan.md`](tasks/plan.md) / [`tasks/todo.md`](tasks/todo.md) | Plan técnico y tareas (el *cómo*) |
| [`design-system/opencode-tray/MASTER.md`](design-system/opencode-tray/MASTER.md) | Sistema de diseño (paleta, tipografía, a11y) |
| [`RELEASE.md`](RELEASE.md) | Notas de release y checklist de publicación |
| [`AGENTS.md`](AGENTS.md) | Reglas de trabajo, comandos y gotchas técnicos |
