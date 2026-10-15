# Release — OpenCode Tray v1.0.0

Primera entrega **portable** de OpenCode Tray: app de bandeja (tray) para Windows que controla el servicio de OpenCode, el acceso web, el consumo de OpenCode Go y la configuración.

## Artefacto

- `build/bin/opencode-tray.exe` — binario **portable único** (Windows 10/11 + WebView2). Sin instalador (portable primero; NSIS queda pendiente).

## Qué incluye

- **Servicio**: estado (detenido/corriendo), iniciar / detener / **forzar detención** (workaround del bug de `opencode service stop`) / refrescar.
- **Consumo**: barras de OpenCode Go (ventanas `rolling` / `weekly` / `monthly`) con estado ok / warn / crítico.
- **Acceso web**: modos local / red (por IP) / remoto (túnel); usuario (`opencode`) y contraseña (enmascarada) + cambio de contraseña.
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
- [ ] **Acceso web**: muestra las URLs y cambia la contraseña correctamente.
- [ ] **Config**: enmascara secretos y abre el editor.
- [ ] **Ajustes**: auto-inicio escribe/borra la clave del registro; idioma alterna es/en.
- [ ] Segunda instancia enfoca la existente (single-instance).
