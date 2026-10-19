package app

import "github.com/wailsapp/wails/v2/pkg/menu"

// FileManagerInterface defines the contract for file management
type FileManagerInterface interface {
	OpenFileMenuHandler(data *menu.CallbackData, currentFile *string)
}

// DocumentProcessorInterface defines the contract for document processing
type DocumentProcessorInterface interface {
	LoadAndDisplayMarkdown(filePath string) error
	AddDocClass(thisClass ...string)
	RemoveDocClass(thisClass ...string)
	ToggleDocClass(thisClass ...string)
}

// ThemeManagerInterface defines the contract for theme management
type ThemeManagerInterface interface {
	GetTheme() string
	SetTheme(theme string)
}

// PrintManagerInterface defines the contract for printing
type PrintManagerInterface interface {
	PrintContent() error
	PrintContentToPDF() error
}
