package notifications

// noopNotify es el sender por defecto (no hace nada). app.go lo reemplaza por
// el sender real —notificaciones nativas de Wails (toast de Windows)— al
// iniciar, donde ya hay un ctx. Se hace así para no acoplar este módulo a Wails
// y mantenerlo testeable, y para evitar beeep, que en Windows usaba una segunda
// librería de bandeja (tadvi/systray) que entraba en conflicto con energye/systray.
func noopNotify(title, msg string, lvl Level) error { return nil }
