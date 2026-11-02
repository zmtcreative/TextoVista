package app

import (
	"testing"
)

func TestCalculateWindowDimensionsDefaultUltraWide(t *testing.T) {
	// Ultra-wide display (e.g. 3440x1440 with 1392 work area height)
	w, h, max := CalculateWindowDimensions(WindowStateConfig{}, 3440, 1392)
	if w != 1040 {
		t.Errorf("width on ultra-wide = %d, want 1040 (must never exceed 1040 by default)", w)
	}
	if h != 984 {
		t.Errorf("height on 1440p = %d, want 984", h)
	}
	if max {
		t.Error("maximized = true, want false")
	}
}

func TestCalculateWindowDimensionsConstrainedDisplays(t *testing.T) {
	tests := []struct {
		name      string
		workAreaW int
		workAreaH int
		wantW     int
		wantH     int
	}{
		{
			name:      "1080p at 125% DPI scale",
			workAreaW: 1536,
			workAreaH: 825,
			wantW:     1040,
			wantH:     765, // 825 - 60
		},
		{
			name:      "1080p at 150% DPI scale",
			workAreaW: 1280,
			workAreaH: 688,
			wantW:     1040,
			wantH:     628, // 688 - 60
		},
		{
			name:      "1366x768 display",
			workAreaW: 1366,
			workAreaH: 728,
			wantW:     1040,
			wantH:     668, // 728 - 60
		},
		{
			name:      "Narrow vertical display",
			workAreaW: 800,
			workAreaH: 1200,
			wantW:     760, // 800 - 40
			wantH:     984,
		},
		{
			name:      "Extreme small display clamped to minimum bounds",
			workAreaW: 500,
			workAreaH: 400,
			wantW:     616, // MinWidthDefault
			wantH:     539, // MinHeightDefault
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h, max := CalculateWindowDimensions(WindowStateConfig{}, tt.workAreaW, tt.workAreaH)
			if w != tt.wantW {
				t.Errorf("width = %d, want %d", w, tt.wantW)
			}
			if h != tt.wantH {
				t.Errorf("height = %d, want %d", h, tt.wantH)
			}
			if max {
				t.Error("maximized = true, want false")
			}
		})
	}
}

func TestCalculateWindowDimensionsRestoresSavedState(t *testing.T) {
	saved := WindowStateConfig{
		Width:     1200,
		Height:    800,
		Maximized: false,
	}

	// On large display, custom user width above 1040 is respected when restored from persistence
	w, h, max := CalculateWindowDimensions(saved, 3440, 1392)
	if w != 1200 || h != 800 {
		t.Errorf("restored size = %dx%d, want 1200x800", w, h)
	}
	if max {
		t.Error("maximized = true, want false")
	}
}

func TestCalculateWindowDimensionsClampsSavedStateWhenDisplayShrinks(t *testing.T) {
	// User saved large dimensions on external monitor (1600x1100), now on small laptop display
	saved := WindowStateConfig{
		Width:     1600,
		Height:    1100,
		Maximized: false,
	}

	w, h, max := CalculateWindowDimensions(saved, 1280, 688)
	if w != 1240 { // 1280 - 40
		t.Errorf("clamped width = %d, want 1240", w)
	}
	if h != 628 { // 688 - 60
		t.Errorf("clamped height = %d, want 628", h)
	}
	if max {
		t.Error("maximized = true, want false")
	}
}

func TestCalculateWindowDimensionsRestoresMaximizedState(t *testing.T) {
	saved := WindowStateConfig{
		Width:     1100,
		Height:    750,
		Maximized: true,
	}

	w, h, max := CalculateWindowDimensions(saved, 1920, 1040)
	if !max {
		t.Error("maximized = false, want true")
	}
	if w != 1100 || h != 750 {
		t.Errorf("unmaximized dimensions = %dx%d, want 1100x750", w, h)
	}
}
