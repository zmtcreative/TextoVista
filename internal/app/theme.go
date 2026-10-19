package app

import (
	"context"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var themeEventsEmit = runtime.EventsEmit

// ThemeManager handles theme-related functionality
type ThemeManager struct {
	mu    sync.RWMutex
	ctx   context.Context
	theme string
}

// NewThemeManager creates a new ThemeManager
func NewThemeManager(ctx context.Context) *ThemeManager {
	return &ThemeManager{
		ctx:   ctx,
		theme: "light",
	}
}

// GetTheme returns the current theme
func (tm *ThemeManager) GetTheme() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.theme
}

// SetTheme sets the theme and emits an event to the frontend
func (tm *ThemeManager) SetTheme(theme string) {
	tm.mu.Lock()
	tm.theme = theme
	tm.mu.Unlock()
	themeEventsEmit(tm.ctx, "theme:changed", theme)
}
