package shell

import "testing"

type mockBackend struct {
	shown  int
	hidden int
}

func (m *mockBackend) Show() { m.shown++ }
func (m *mockBackend) Hide() { m.hidden++ }

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

	c.Hide()
	if c.IsVisible() || b.hidden != 1 {
		t.Errorf("Hide: visible=%v hidden=%d, want false/1", c.IsVisible(), b.hidden)
	}

	// oculto → Toggle → visible
	c.Toggle()
	if !c.IsVisible() || b.shown != 2 {
		t.Errorf("Toggle (mostrar): visible=%v shown=%d, want true/2", c.IsVisible(), b.shown)
	}

	// visible → Toggle → oculto
	c.Toggle()
	if c.IsVisible() || b.hidden != 2 {
		t.Errorf("Toggle (ocultar): visible=%v hidden=%d, want false/2", c.IsVisible(), b.hidden)
	}
}
