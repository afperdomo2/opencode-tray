# Release — OpenCode Tray v1.3.0

App de bandeja (tray) para Windows que controla el servicio de OpenCode, el acceso web, el consumo de OpenCode Go y la configuración.

## Novedades v1.3.0

- **Tooltip del tray**: consumo mensual con barra (`▰▰▱▱▱`), porcentaje y reinicio humanizado (`en 4 h`, `mañana a las 12:30`, `el 12 jun 26`), con iconos Unicode y refresco cada minuto aunque el panel esté cerrado.
- **Acceso web**: URLs reales desde `/api/info` con el puerto correcto (45278 en Windows), badges Local/LAN/Remoto y botones para **abrir en el navegador** y **copiar** con confirmación.
- **Modos exclusivos**: al pasar a Local/Red se desactiva OpenTunnel; activar Remoto pide confirmación explícita porque expone el servidor a Internet.
- **Skeleton** de 2 filas en la lista de URLs.
- **Instalador per-user con Inno Setup** (`scripts/build-installer.ps1`), con detección de WebView2 y aviso si falta `opencode` en PATH.
- **Fix de build**: la generación de bindings ya no falla en silencio si la app está abierta (mutex de instancia única desactivado con `-tags bindings`).
- **Fix crítico del tray**: el icono se registra en el hilo principal antes de `wails.Run`; el clic izquierdo abre/cierra el panel y el clic derecho vuelve a mostrar el menú (Mostrar panel / Salir).

## Novedades v1.2.0

- **Menú simplificado**: se quitan las pestañas **Consumo** y **Config**; nav = **Resumen · Servicio · Acceso web · Ajustes**.
- **Servicio unificado**: incluye el control del servicio **y** la configuración (MCP + JSON en modal).
- **Dashboard sin acciones**: el Resumen muestra **widgets** (Consumo · Servicio+Config) de **solo lectura**; las acciones viven en cada menú.
- **Idioma movido a Ajustes** (control segmentado ES | EN); se elimina la cabecera del panel.
- **Botones compactos (`sm`)**: tamaño **estándar** en toda la app (con texto e icono).
- **Config JSON en modal** accesible.
- **Ajustes**: los checkboxes se reemplazan por **switches** estilizados.
- **Color graduado** en las barras de consumo: verde → ámbar → naranja → rojo según proximidad al límite.

## Novedades v1.1.0

- **Vista Resumen (overview)** como pantalla inicial: Consumo → MCP/Config → Servicio, en cards.
- **Iconos SVG** en navegación y botones (sin emoji).
- **Idioma** como control segmentado (ES | EN); se elimina el título redundante de la cabecera.
- **Config JSON** en acordeón (colapsado por defecto) y **profundidad visual** (cards, sombras, hover).

## Artefacto

- `build/bin/opencode-tray.exe` — binario **portable único** (Windows 10/11 + WebView2). Instalador per-user opcional: `build/bin/OpenCodeTray-Setup-1.3.0.exe` (Inno Setup).

## Qué incluye

- **Servicio**: estado (detenido/corriendo), iniciar / detener / **forzar detención** (workaround del bug de `opencode service stop`) / refrescar.
- **Consumo**: barras de OpenCode Go (ventanas `rolling` / `weekly` / `monthly`) con estado ok / warn / crítico.
- **Acceso web**: modos local / red (por IP) / remoto (túnel), exclusivos; URLs reales con **abrir** y **copiar**; usuario (`opencode`) y contraseña (enmascarada) + cambio de contraseña.
- **Config**: visor de servidores MCP y del JSON de configuración global (secretos enmascarados) + abrir en editor externo.
- **Ajustes**: auto-inicio con Windows, idioma (es/en), toggles de notificaciones.
- i18n español/inglés, una sola instancia (single-instance), panel oscuro con sistema de diseño (Dark Mode OLED).

## Limitación conocida

- **Notificaciones (toasts)**: la **lógica** de alertas (ritmo de consumo, agotamiento, servicio detenido) está implementada y probada, pero el **envío de toasts está deshabilitado en runtime** — el mecanismo de toast de Wails (go-toast) entraba en conflicto con el bucle de la bandeja y la dejaba muda. Los toggles y la evaluación funcionan; no se muestra ningún toast. Detalle y cómo reactivarlo: `specs/SPEC-notifications.md` → Estado.
- **Fuentes web**: la UI carga IBM Plex Sans + JetBrains Mono desde Google Fonts; sin red degrada a las fuentes del sistema (funcional).

## Cómo ejecutar

1. Ejecutar `build/bin/opencode-tray.exe` (aparece el icono en la bandeja).
2. **Clic izquierdo** en el icono → abre el panel. **Clic derecho** → menú (Mostrar panel / Salir).
3. Requiere la CLI de OpenCode (`opencode` en PATH) para controlar el servicio de fondo.

## Release check (validar en Windows limpio)

- [ ] `opencode-tray.exe` arranca y muestra el icono en la bandeja (sin ventana visible al inicio).
- [ ] Clic izquierdo abre el panel; clic derecho muestra el menú (Mostrar panel / Salir).
- [ ] **Servicio**: muestra el estado real; "Forzar detención" deja el servicio muerto (bug real).
- [ ] **Consumo**: muestra las tres barras (o mensaje claro si no hay token/suscripción).
- [ ] **Tooltip**: al pasar el cursor muestra barra, % mensual y reinicio humanizado; se actualiza sin abrir el panel.
- [ ] **Acceso web**: muestra las URLs con badge y permite abrir/copiar; remoto pide confirmación y se apaga al volver a local/red.
- [ ] **Instalador**: `OpenCodeTray-Setup-1.3.0.exe` instala por usuario, detecta WebView2 y desinstala limpio.
- [ ] **Config**: enmascara secretos y abre el editor.
- [ ] **Ajustes**: auto-inicio escribe/borra la clave del registro; idioma alterna es/en.
- [ ] Segunda instancia enfoca la existente (single-instance).
