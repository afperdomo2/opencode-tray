package shell

// Geometría del flyout anclado al icono del tray. La lógica es pura para poder
// probarla sin Win32: recibe rectángulos en píxeles físicos y devuelve la
// posición absoluta de la ventana.

const (
	// FlyoutWindowClass identifica la ventana de Wails en Win32.
	FlyoutWindowClass = "OpenCodeTrayFlyout"

	// FlyoutWidth es el ancho fijo del panel.
	FlyoutWidth = 380
	// FlyoutMinHeight y FlyoutMaxHeight acotan el alto según contenido.
	FlyoutMinHeight = 320
	FlyoutMaxHeight = 600
	// flyoutMargin separa el panel del borde del área de trabajo.
	flyoutMargin = 8
	// flyoutIconOffset aproxima el centro del icono para alinear el panel.
	flyoutIconOffset = 24
)

// Point es una coordenada en píxeles físicos.
type Point struct {
	X int
	Y int
}

// Rect es un rectángulo en píxeles físicos.
type Rect struct {
	Left   int
	Top    int
	Right  int
	Bottom int
}

func (r Rect) width() int  { return r.Right - r.Left }
func (r Rect) height() int { return r.Bottom - r.Top }

type taskbarEdge int

const (
	taskbarBottom taskbarEdge = iota
	taskbarTop
	taskbarLeft
	taskbarRight
)

// edgeFromWorkArea deduce dónde está la taskbar comparando el área de trabajo
// con el monitor. Si coinciden (taskbar auto-oculta o sin taskbar), asume abajo.
func edgeFromWorkArea(work, monitor Rect) taskbarEdge {
	switch {
	case work.Bottom < monitor.Bottom:
		return taskbarBottom
	case work.Top > monitor.Top:
		return taskbarTop
	case work.Left > monitor.Left:
		return taskbarLeft
	case work.Right < monitor.Right:
		return taskbarRight
	default:
		return taskbarBottom
	}
}

func clamp(v, min, max int) int {
	if max < min {
		return min
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// ClampFlyoutHeight limita el alto solicitado al rango permitido y al área de
// trabajo disponible.
func ClampFlyoutHeight(height, workHeight int) int {
	if height < FlyoutMinHeight {
		height = FlyoutMinHeight
	}
	limit := FlyoutMaxHeight
	if available := workHeight - 2*flyoutMargin; available < limit {
		limit = available
	}
	if limit < FlyoutMinHeight {
		limit = FlyoutMinHeight
	}
	if height > limit {
		height = limit
	}
	return height
}

// CornerPosition calcula la esquina superior izquierda del panel para la
// esquina inferior derecha del área de trabajo. Se usa en los SO donde no hay
// una API para leer la geometría del icono del tray (Linux/macOS): el panel
// aparece como un flyout flotante en esa esquina.
func CornerPosition(work Rect, width, height int) Point {
	if width <= 0 {
		width = FlyoutWidth
	}
	if height <= 0 {
		height = FlyoutMinHeight
	}
	return Point{
		X: clamp(work.Right-width-flyoutMargin, work.Left, work.Right),
		Y: clamp(work.Bottom-height-flyoutMargin, work.Top, work.Bottom),
	}
}

// AnchorPosition calcula la esquina superior izquierda del panel para que quede
// pegado a la taskbar y alineado con el icono. width/height están en píxeles
// físicos.
func AnchorPosition(cursor Point, work, monitor Rect, width, height int) Point {
	if width <= 0 {
		width = FlyoutWidth
	}
	if height <= 0 {
		height = FlyoutMinHeight
	}

	minX := work.Left + flyoutMargin
	maxX := work.Right - width - flyoutMargin
	minY := work.Top + flyoutMargin
	maxY := work.Bottom - height - flyoutMargin

	switch edgeFromWorkArea(work, monitor) {
	case taskbarTop:
		return Point{
			X: clamp(cursor.X-width+flyoutIconOffset, minX, maxX),
			Y: minY,
		}
	case taskbarLeft:
		return Point{
			X: minX,
			Y: clamp(cursor.Y-height+flyoutIconOffset, minY, maxY),
		}
	case taskbarRight:
		return Point{
			X: clamp(work.Right-width-flyoutMargin, minX, maxX),
			Y: clamp(cursor.Y-height+flyoutIconOffset, minY, maxY),
		}
	default: // taskbarBottom
		return Point{
			X: clamp(cursor.X-width+flyoutIconOffset, minX, maxX),
			Y: clamp(work.Bottom-height-flyoutMargin, minY, maxY),
		}
	}
}
