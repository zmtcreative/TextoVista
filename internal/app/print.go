package app

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	printEventsEmit     = runtime.EventsEmit
	printSaveFileDialog = runtime.SaveFileDialog
)

// PrintManager handles printing functionality
type PrintManager struct {
	ctx context.Context
}

// NewPrintManager creates a new PrintManager
func NewPrintManager(ctx context.Context) *PrintManager {
	return &PrintManager{ctx: ctx}
}

// PrintContent prints the current HTML content
func (pm *PrintManager) PrintContent() error {
	printEventsEmit(pm.ctx, "print-content")
	return nil
}

// PrintContentToPDF exports the current content to PDF (Windows-specific)
func (pm *PrintManager) PrintContentToPDF() error {
	filePath, err := printSaveFileDialog(pm.ctx, runtime.SaveDialogOptions{
		Title:           "Save as PDF",
		DefaultFilename: "document.pdf",
		Filters: []runtime.FileFilter{
			{DisplayName: "PDF Files (*.pdf)", Pattern: "*.pdf"},
		},
	})
	if err != nil {
		return err
	}

	if filePath == "" {
		return nil // User cancelled
	}

	printEventsEmit(pm.ctx, "save-as-pdf", filePath)
	return nil
}
