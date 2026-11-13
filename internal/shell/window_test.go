package shell

import (
	"testing"
)

type mockBackend struct {
	shown   int
	hidden  int
	resizes int
	places  int
	focus   int
	width   int
	height  int
	anchor  Point
}

func (m *mockBackend) Show() { m.shown++ }
func (m *mockBackend) Hide() { m.hidden++ }
func (m *mockBackend) Resize(width, height int) {
	m.resizes++
	m.width, m.height = width, height
}
func (m *mockBackend) Place(anchor Point, width, height int) {
	m.places++
	m.anchor = anchor
	m.width, m.height = width, height
}
func (m *mockBackend) Focus() { m.focus++ }

func TestWindowControllerStartsHidden(t *testing.T) {
	b := &mockBackend{}
	c := NewWindowController(b)
	if c.IsVisible() {
		t.Error("debería iniciar oculto")
	}
	if b.shown != 0 || b.hidden != 0 {
		t.Errorf("no debería tocar el backend al crear: shown=%d hidden=%d", b.shown, b.hidden)
	}
}

func TestWindowControllerShowHideToggle(t *testing.T) {
	b := &mockBackend{}
	c := NewWindowController(b)

	c.Show()
	if !c.IsVisible() || b.shown != 1 {
		t.Errorf("Show: visible=%v shown=%d, want true/1", c.IsVisible(), b.shown)
	}
	if b.resizes != 1 || b.places != 2 || b.focus != 1 {
		t.Errorf("Show: resizes=%d places=%d focus=%d, want 1/2/1", b.resizes, b.places, b.focus)
	}

	c.Hide()
	if c.IsVisible() || b.hidden != 1 {
		t.Errorf("Hide: visible=%v hidden=%d, want false/1", c.IsVisible(), b.hidden)
	}

	// oculto -> Toggle -> visible
	c.Toggle()
	if !c.IsVisible() || b.shown != 2 {
		t.Errorf("Toggle (mostrar): visible=%v shown=%d, want true/2", c.IsVisible(), b.shown)
	}

	// visible -> Toggle -> oculto
	c.Toggle()
	if c.IsVisible() || b.hidden != 2 {
		t.Errorf("Toggle (ocultar): visible=%v hidden=%d, want false/2", c.IsVisible(), b.hidden)
	}
}

func TestWindowControllerShowAtAnchors(t *testing.T) {
	b := &mockBackend{}
	c := NewWindowController(b)
	anchor := Point{X: 1200, Y: 900}

	c.ShowAt(anchor)
	if !c.IsVisible() {
		t.Fatal("debería estar visible")
	}
	if b.anchor != anchor {
		t.Errorf("ancla = %+v, want %+v", b.anchor, anchor)
	}
	if b.width != FlyoutWidth {
		t.Errorf("ancho = %d, want %d", b.width, FlyoutWidth)
	}
}

func TestWindowControllerToggleAtAfterHideShowsImmediately(t *testing.T) {
	b := &mockBackend{}
	c := NewWindowController(b)
	anchor := Point{X: 100, Y: 200}

	c.ShowAt(anchor)
	c.Hide()
	c.ToggleAt(anchor)
	if !c.IsVisible() || b.shown != 2 {
		t.Fatalf("ToggleAt tras cerrar debe mostrar de inmediato: visible=%v shown=%d", c.IsVisible(), b.shown)
	}

	c.ToggleAt(anchor)
	if c.IsVisible() || b.hidden != 2 {
		t.Fatalf("ToggleAt con el panel visible debe ocultar: visible=%v hidden=%d", c.IsVisible(), b.hidden)
	}
}

func TestWindowControllerStartsWithFixedHeight(t *testing.T) {
	b := &mockBackend{}
	c := NewWindowController(b)

	c.ShowAt(Point{X: 10, Y: 20})
	if b.width != FlyoutWidth || b.height != FlyoutMaxHeight {
		t.Errorf("tamaño = %dx%d, want %dx%d", b.width, b.height, FlyoutWidth, FlyoutMaxHeight)
	}
}

func TestWindowControllerSetHeightClamps(t *testing.T) {
	b := &mockBackend{}
	c := NewWindowController(b)

	c.SetHeight(100)
	if b.height != FlyoutMinHeight {
		t.Errorf("alto bajo = %d, want %d", b.height, FlyoutMinHeight)
	}
	if b.resizes != 1 || b.places != 0 {
		t.Errorf("oculto: resizes=%d places=%d, want 1/0", b.resizes, b.places)
	}

	c.SetHeight(999)
	if b.height != FlyoutMaxHeight {
		t.Errorf("alto alto = %d, want %d", b.height, FlyoutMaxHeight)
	}

	resizes := b.resizes
	c.SetHeight(FlyoutMaxHeight)
	if b.resizes != resizes {
		t.Error("no debería redimensionar si el alto no cambia")
	}

	c.ShowAt(Point{X: 10, Y: 20})
	places := b.places
	c.SetHeight(400)
	if b.places != places+1 {
		t.Errorf("visible: places=%d, want %d", b.places, places+1)
	}
}
