//go:build !windows

package app

// GetAvailableWorkArea provides a fallback for non-Windows platforms.
func GetAvailableWorkArea() (int, int, error) {
	return MaxWidthDefault + WindowMarginX, MaxHeightDefault + WindowMarginY, nil
}

