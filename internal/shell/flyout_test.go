package shell

import "testing"

func TestAnchorPositionTaskbarEdges(t *testing.T) {
	monitor := Rect{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	cursor := Point{X: 1900, Y: 1030}
	width, height := 380, 400

	tests := []struct {
		name string
		work Rect
		want Point
	}{
		{
			name: "taskbar abajo",
			work: Rect{Left: 0, Top: 0, Right: 1920, Bottom: 1040},
			want: Point{X: 1532, Y: 632},
		},
		{
			name: "taskbar arriba",
			work: Rect{Left: 0, Top: 40, Right: 1920, Bottom: 1080},
			want: Point{X: 1532, Y: 48},
		},
		{
			name: "taskbar izquierda",
			work: Rect{Left: 48, Top: 0, Right: 1920, Bottom: 1080},
			want: Point{X: 56, Y: 654},
		},
		{
			name: "taskbar derecha",
			work: Rect{Left: 0, Top: 0, Right: 1872, Bottom: 1080},
			want: Point{X: 1484, Y: 654},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AnchorPosition(cursor, tt.work, monitor, width, height)
			if got != tt.want {
				t.Errorf("AnchorPosition() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAnchorPositionMultiMonitorNegative(t *testing.T) {
	monitor := Rect{Left: -1920, Top: 0, Right: 0, Bottom: 1080}
	work := Rect{Left: -1920, Top: 0, Right: 0, Bottom: 1040}
	cursor := Point{X: -10, Y: 1030}

	got := AnchorPosition(cursor, work, monitor, 380, 400)
	want := Point{X: -388, Y: 632}
	if got != want {
		t.Errorf("AnchorPosition() = %+v, want %+v", got, want)
	}
}

func TestAnchorPositionSmallWorkArea(t *testing.T) {
	monitor := Rect{Left: 0, Top: 0, Right: 300, Bottom: 200}
	work := Rect{Left: 0, Top: 0, Right: 300, Bottom: 160}
	cursor := Point{X: 290, Y: 150}

	got := AnchorPosition(cursor, work, monitor, 380, 400)
	if got.X < 8 || got.Y < 8 {
		t.Errorf("debería respetar el margen mínimo: %+v", got)
	}
}

func TestClampFlyoutHeight(t *testing.T) {
	tests := []struct {
		name       string
		height     int
		workHeight int
		want       int
	}{
		{name: "mínimo", height: 100, workHeight: 1080, want: FlyoutMinHeight},
		{name: "máximo", height: 999, workHeight: 1080, want: FlyoutMaxHeight},
		{name: "limitado por work area", height: 590, workHeight: 550, want: 534},
		{name: "work area diminuta", height: 350, workHeight: 320, want: FlyoutMinHeight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClampFlyoutHeight(tt.height, tt.workHeight); got != tt.want {
				t.Errorf("ClampFlyoutHeight(%d, %d) = %d, want %d", tt.height, tt.workHeight, got, tt.want)
			}
		})
	}
}

func TestCornerPosition(t *testing.T) {
	work := Rect{Left: 0, Top: 0, Right: 1920, Bottom: 1080}

	tests := []struct {
		name          string
		width, height int
		want          Point
	}{
		{name: "tamaño del flyout", width: 380, height: 600, want: Point{X: 1532, Y: 472}},
		{name: "tamaño inválido usa mínimos", width: 0, height: 0, want: Point{X: 1532, Y: 752}},
		{name: "más grande que el área", width: 3000, height: 2000, want: Point{X: 0, Y: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CornerPosition(work, tt.width, tt.height); got != tt.want {
				t.Errorf("CornerPosition(%+v, %d, %d) = %+v, want %+v", work, tt.width, tt.height, got, tt.want)
			}
		})
	}
}
