package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	fileOpenDialog       = runtime.OpenFileDialog
	fileEventsEmit       = runtime.EventsEmit
	fileMessageDialog    = runtime.MessageDialog
	fileBinaryCheck      = func(detector *BinaryDetector, filePath string) (bool, error) { return detector.IsBinaryFile(filePath) }
	fileLoadAndDisplayMD = func(processor *DocumentProcessor, filePath string) error {
		return processor.LoadAndDisplayMarkdown(filePath)
	}
)

// Custom error types for better error handling
var (
	ErrBinaryFileCheckFailed = errors.New("binary file check failed")
	ErrBinaryFileCannotOpen  = errors.New("binary file cannot be opened")
)

// FileManager handles file operations and dialog interactions
type FileManager struct {
	ctx            context.Context
	binaryDetector *BinaryDetector
	docProcessor   *DocumentProcessor
}

func binaryCheckFailedMessage(filePath string) string {
	return fmt.Sprintf("File MAY be binary: %s\n\nPlease select a text-based Markdown file (.md or .markdown).", filepath.Base(filePath))
}

func cannotOpenBinaryFileMessage(filePath string) string {
	return fmt.Sprintf("Cannot open binary file: %s\n\nPlease select a text-based Markdown file (.md or .markdown).", filepath.Base(filePath))
}

// NewFileManager creates a new FileManager
func NewFileManager(ctx context.Context, binaryDetector *BinaryDetector, docProcessor *DocumentProcessor) *FileManager {
	return &FileManager{
		ctx:            ctx,
		binaryDetector: binaryDetector,
		docProcessor:   docProcessor,
	}
}

// LoadFile validates and renders a selected markdown file.
func (fm *FileManager) LoadFile(filePath string) error {
	isBinary, err := fileBinaryCheck(fm.binaryDetector, filePath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrBinaryFileCheckFailed, err)
	}

	if isBinary {
		return fmt.Errorf("%w: %s", ErrBinaryFileCannotOpen, filePath)
	}

	if err := fileLoadAndDisplayMD(fm.docProcessor, filePath); err != nil {
		return fmt.Errorf("failed to load file %s: %w", filePath, err)
	}

	return nil
}

// OpenFileMenuHandler handles the File > Open menu action
func (fm *FileManager) OpenFileMenuHandler(_ *menu.CallbackData, currentFile *string) {
	log.Println("##> LOG: File -> Open menu item clicked. Opening file dialog...")

	// Open a file dialog to allow the user to select a Markdown file.
	selection, err := fileOpenDialog(fm.ctx, runtime.OpenDialogOptions{
		Title: "Open Markdown File",
		Filters: []runtime.FileFilter{
			{DisplayName: "Markdown Files (*.md;*.markdown)", Pattern: "*.md;*.markdown"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		// Wails runtime doesn't provide a specific error type for cancellation.
		// Check error message for cancellation indicators.
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "cancelled") || strings.Contains(errMsg, "canceled") {
			log.Println("##> LOG: File dialog cancelled by user.")
			return
		}
		log.Printf("##> LOG: Error opening file dialog: %v", err)
		fileEventsEmit(fm.ctx, "error", "Failed to open file dialog: "+err.Error())
		return
	}

	if selection != "" {
		log.Printf("##> LOG: User selected file: %s", selection)
		err = fm.LoadFile(selection)
		if err != nil {
			log.Printf("##> LOG: Error loading selected Markdown file %q: %v", selection, err)
			if errors.Is(err, ErrBinaryFileCheckFailed) {
				fileMessageDialog(fm.ctx, runtime.MessageDialogOptions{
					Type:    runtime.ErrorDialog,
					Title:   "Binary File Check Failed",
					Message: binaryCheckFailedMessage(selection),
				})
				return
			}
			if errors.Is(err, ErrBinaryFileCannotOpen) {
				fileMessageDialog(fm.ctx, runtime.MessageDialogOptions{
					Type:    runtime.ErrorDialog,
					Title:   "Cannot Open Binary File",
					Message: cannotOpenBinaryFileMessage(selection),
				})
				return
			}
			fileEventsEmit(fm.ctx, "error", "Failed to load selected file: "+err.Error())
		} else {
			log.Printf("##> LOG: Successfully loaded Markdown file: %s", selection)
			*currentFile = selection // Update currentFile to the newly opened file
		}
	} else {
		log.Println("##> LOG: No file selected. User cancelled the dialog.")
	}
}
