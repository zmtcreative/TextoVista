package app

const (
	MinWidthDefault  = 616
	MinHeightDefault = 539
	MaxWidthDefault  = 1040
	MaxHeightDefault = 984
	WindowMarginX    = 40
	WindowMarginY    = 60
)

// CalculateWindowDimensions computes the startup width, height, and maximized state.
// - If savedState.Maximized is true, maximized is returned as true.
// - If valid saved dimensions exist, they are restored and clamped to the current work area.
// - If no valid saved dimensions exist, default dimensions are calculated:
//   - Width is never wider than MaxWidthDefault (1040), scaling down if workAreaW - 40 is smaller.
//   - Height scales down if workAreaH - 60 is smaller than MaxHeightDefault (984), clamped at MinHeightDefault (539).
func CalculateWindowDimensions(savedState WindowStateConfig, workAreaW, workAreaH int) (int, int, bool) {
	if savedState.Maximized {
		w := savedState.Width
		h := savedState.Height
		if w <= 0 || h <= 0 {
			w, h, _ = CalculateWindowDimensions(WindowStateConfig{}, workAreaW, workAreaH)
		} else {
			w = clamp(w, MinWidthDefault, max(MinWidthDefault, workAreaW-WindowMarginX))
			h = clamp(h, MinHeightDefault, max(MinHeightDefault, workAreaH-WindowMarginY))
		}
		return w, h, true
	}

	if savedState.Width > 0 && savedState.Height > 0 {
		w := clamp(savedState.Width, MinWidthDefault, max(MinWidthDefault, workAreaW-WindowMarginX))
		h := clamp(savedState.Height, MinHeightDefault, max(MinHeightDefault, workAreaH-WindowMarginY))
		return w, h, false
	}

	maxUsableW := max(MinWidthDefault, workAreaW-WindowMarginX)
	w := min(MaxWidthDefault, maxUsableW)
	w = max(MinWidthDefault, w)

	maxUsableH := max(MinHeightDefault, workAreaH-WindowMarginY)
	h := min(MaxHeightDefault, maxUsableH)
	h = max(MinHeightDefault, h)

	return w, h, false
}

func clamp(val, minVal, maxVal int) int {
	if val < minVal {
		return minVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}

