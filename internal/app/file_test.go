package app

import (
	"context"
	"errors"
	"testing"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type capturedFileEvent struct {
	ctx  context.Context
	name string
	data []interface{}
}

type capturedMessageDialog struct {
	ctx     context.Context
	options runtime.MessageDialogOptions
}

func captureFileRuntime(t *testing.T) (*[]capturedFileEvent, *[]capturedMessageDialog) {
	t.Helper()

	events := []capturedFileEvent{}
	dialogs := []capturedMessageDialog{}
	originalOpenDialog := fileOpenDialog
	originalEventsEmit := fileEventsEmit
	originalMessageDialog := fileMessageDialog
	originalBinaryCheck := fileBinaryCheck
	originalLoadAndDisplay := fileLoadAndDisplayMD

	fileOpenDialog = func(ctx context.Context, options runtime.OpenDialogOptions) (string, error) {
		return "", nil
	}
	fileEventsEmit = func(ctx context.Context, eventName string, optionalData ...interface{}) {
		dataCopy := append([]interface{}{}, optionalData...)
		events = append(events, capturedFileEvent{ctx: ctx, name: eventName, data: dataCopy})
	}
	fileMessageDialog = func(ctx context.Context, options runtime.MessageDialogOptions) (string, error) {
		dialogs = append(dialogs, capturedMessageDialog{ctx: ctx, options: options})
		return "", nil
	}
	fileBinaryCheck = func(detector *BinaryDetector, filePath string) (bool, error) {
		return detector.IsBinaryFile(filePath)
	}
	fileLoadAndDisplayMD = func(processor *DocumentProcessor, filePath string) error {
		return processor.LoadAndDisplayMarkdown(filePath)
	}

	t.Cleanup(func() {
		fileOpenDialog = originalOpenDialog
		fileEventsEmit = originalEventsEmit
		fileMessageDialog = originalMessageDialog
		fileBinaryCheck = originalBinaryCheck
		fileLoadAndDisplayMD = originalLoadAndDisplay
	})

	return &events, &dialogs
}

func TestFileManagerLoadFileBinaryCheckFailure(t *testing.T) {
	fm := NewFileManager(context.Background(), nil, nil)
	_, _ = captureFileRuntime(t)
	fileBinaryCheck = func(detector *BinaryDetector, filePath string) (bool, error) {
		return false, errors.New("probe failed")
	}

	err := fm.LoadFile("broken.md")
	if err == nil {
		t.Fatal("LoadFile() error = nil, want binary check failure")
	}
	if err.Error() != "binary file check failed: probe failed" {
		t.Fatalf("LoadFile() error = %q, want binary check failure", err)
	}
}

func TestFileManagerLoadFileRejectsBinaryFile(t *testing.T) {
	fm := NewFileManager(context.Background(), nil, nil)
	_, _ = captureFileRuntime(t)
	fileBinaryCheck = func(detector *BinaryDetector, filePath string) (bool, error) {
		return true, nil
	}

	err := fm.LoadFile("binary.exe")
	if err == nil {
		t.Fatal("LoadFile() error = nil, want binary rejection")
	}
	if err.Error() != "binary file cannot be opened: binary.exe" {
		t.Fatalf("LoadFile() error = %q, want binary rejection", err)
	}
}

func TestFileManagerLoadFilePropagatesDocumentLoadFailure(t *testing.T) {
	fm := NewFileManager(context.Background(), nil, nil)
	_, _ = captureFileRuntime(t)
	fileBinaryCheck = func(detector *BinaryDetector, filePath string) (bool, error) {
		return false, nil
	}
	fileLoadAndDisplayMD = func(processor *DocumentProcessor, filePath string) error {
		return errors.New("render failed")
	}

	err := fm.LoadFile("sample.md")
	if err == nil {
		t.Fatal("LoadFile() error = nil, want downstream load failure")
	}
	if err.Error() != "failed to load file sample.md: render failed" {
		t.Fatalf("LoadFile() error = %q, want wrapped downstream failure", err)
	}
}

func TestFileManagerLoadFileSuccess(t *testing.T) {
	fm := NewFileManager(context.Background(), nil, nil)
	_, _ = captureFileRuntime(t)
	called := false
	fileBinaryCheck = func(detector *BinaryDetector, filePath string) (bool, error) {
		return false, nil
	}
	fileLoadAndDisplayMD = func(processor *DocumentProcessor, filePath string) error {
		called = true
		if filePath != "sample.md" {
			t.Fatalf("LoadAndDisplayMarkdown filePath = %q, want %q", filePath, "sample.md")
		}
		return nil
	}

	if err := fm.LoadFile("sample.md"); err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	if !called {
		t.Fatal("LoadFile() did not call document loader")
	}
}

func TestFileManagerOpenFileMenuHandlerCancelledDialog(t *testing.T) {
	fm := NewFileManager(context.Background(), nil, nil)
	events, dialogs := captureFileRuntime(t)
	currentFile := "unchanged.md"
	fileOpenDialog = func(ctx context.Context, options runtime.OpenDialogOptions) (string, error) {
		return "", errors.New("The user cancelled the dialog")
	}

	fm.OpenFileMenuHandler(&menu.CallbackData{}, &currentFile)

	if currentFile != "unchanged.md" {
		t.Fatalf("currentFile = %q, want unchanged value", currentFile)
	}
	if len(*events) != 0 {
		t.Fatalf("captured %d events, want 0", len(*events))
	}
	if len(*dialogs) != 0 {
		t.Fatalf("captured %d dialogs, want 0", len(*dialogs))
	}
}

func TestFileManagerOpenFileMenuHandlerDialogErrorEmitsEvent(t *testing.T) {
	fm := NewFileManager(context.Background(), nil, nil)
	events, dialogs := captureFileRuntime(t)
	currentFile := "unchanged.md"
	fileOpenDialog = func(ctx context.Context, options runtime.OpenDialogOptions) (string, error) {
		if options.Title != "Open Markdown File" {
			t.Fatalf("dialog title = %q, want %q", options.Title, "Open Markdown File")
		}
		if len(options.Filters) != 2 {
			t.Fatalf("dialog filter count = %d, want 2", len(options.Filters))
		}
		return "", errors.New("system dialog failed")
	}

	fm.OpenFileMenuHandler(&menu.CallbackData{}, &currentFile)

	if len(*events) != 1 {
		t.Fatalf("captured %d events, want 1", len(*events))
	}
	if (*events)[0].name != "error" {
		t.Fatalf("event name = %q, want %q", (*events)[0].name, "error")
	}
	if (*events)[0].data[0] != "Failed to open file dialog: system dialog failed" {
		t.Fatalf("event payload = %v, want dialog error", (*events)[0].data[0])
	}
	if len(*dialogs) != 0 {
		t.Fatalf("captured %d dialogs, want 0", len(*dialogs))
	}
}

func TestFileManagerOpenFileMenuHandlerBinaryCheckFailureShowsDialog(t *testing.T) {
	fm := NewFileManager(context.Background(), nil, nil)
	events, dialogs := captureFileRuntime(t)
	currentFile := "unchanged.md"
	fileOpenDialog = func(ctx context.Context, options runtime.OpenDialogOptions) (string, error) {
		return `C:\docs\maybe-binary.md`, nil
	}
	fileBinaryCheck = func(detector *BinaryDetector, filePath string) (bool, error) {
		return false, errors.New("scanner offline")
	}

	fm.OpenFileMenuHandler(&menu.CallbackData{}, &currentFile)

	if len(*dialogs) != 1 {
		t.Fatalf("captured %d dialogs, want 1", len(*dialogs))
	}
	if (*dialogs)[0].options.Title != "Binary File Check Failed" {
		t.Fatalf("dialog title = %q, want %q", (*dialogs)[0].options.Title, "Binary File Check Failed")
	}
	if (*dialogs)[0].options.Message != binaryCheckFailedMessage(`C:\docs\maybe-binary.md`) {
		t.Fatalf("dialog message = %q, want binary check failed message", (*dialogs)[0].options.Message)
	}
	if len(*events) != 0 {
		t.Fatalf("captured %d events, want 0", len(*events))
	}
	if currentFile != "unchanged.md" {
		t.Fatalf("currentFile = %q, want unchanged value", currentFile)
	}
}

func TestFileManagerOpenFileMenuHandlerBinaryFileShowsDialog(t *testing.T) {
	fm := NewFileManager(context.Background(), nil, nil)
	events, dialogs := captureFileRuntime(t)
	currentFile := "unchanged.md"
	fileOpenDialog = func(ctx context.Context, options runtime.OpenDialogOptions) (string, error) {
		return `C:\docs\binary.dll`, nil
	}
	fileBinaryCheck = func(detector *BinaryDetector, filePath string) (bool, error) {
		return true, nil
	}

	fm.OpenFileMenuHandler(&menu.CallbackData{}, &currentFile)

	if len(*dialogs) != 1 {
		t.Fatalf("captured %d dialogs, want 1", len(*dialogs))
	}
	if (*dialogs)[0].options.Title != "Cannot Open Binary File" {
		t.Fatalf("dialog title = %q, want %q", (*dialogs)[0].options.Title, "Cannot Open Binary File")
	}
	if (*dialogs)[0].options.Message != cannotOpenBinaryFileMessage(`C:\docs\binary.dll`) {
		t.Fatalf("dialog message = %q, want binary file message", (*dialogs)[0].options.Message)
	}
	if len(*events) != 0 {
		t.Fatalf("captured %d events, want 0", len(*events))
	}
	if currentFile != "unchanged.md" {
		t.Fatalf("currentFile = %q, want unchanged value", currentFile)
	}
}

func TestFileManagerOpenFileMenuHandlerGenericLoadErrorEmitsEvent(t *testing.T) {
	fm := NewFileManager(context.Background(), nil, nil)
	events, dialogs := captureFileRuntime(t)
	currentFile := "unchanged.md"
	fileOpenDialog = func(ctx context.Context, options runtime.OpenDialogOptions) (string, error) {
		return `C:\docs\sample.md`, nil
	}
	fileBinaryCheck = func(detector *BinaryDetector, filePath string) (bool, error) {
		return false, nil
	}
	fileLoadAndDisplayMD = func(processor *DocumentProcessor, filePath string) error {
		return errors.New("render failed")
	}

	fm.OpenFileMenuHandler(&menu.CallbackData{}, &currentFile)

	if len(*events) != 1 {
		t.Fatalf("captured %d events, want 1", len(*events))
	}
	if (*events)[0].name != "error" {
		t.Fatalf("event name = %q, want %q", (*events)[0].name, "error")
	}
	if (*events)[0].data[0] != `Failed to load selected file: failed to load file C:\docs\sample.md: render failed` {
		t.Fatalf("event payload = %v, want wrapped generic load error", (*events)[0].data[0])
	}
	if len(*dialogs) != 0 {
		t.Fatalf("captured %d dialogs, want 0", len(*dialogs))
	}
	if currentFile != "unchanged.md" {
		t.Fatalf("currentFile = %q, want unchanged value", currentFile)
	}
}

func TestFileManagerOpenFileMenuHandlerSuccessUpdatesCurrentFile(t *testing.T) {
	fm := NewFileManager(context.Background(), nil, nil)
	events, dialogs := captureFileRuntime(t)
	currentFile := "old.md"
	called := false
	fileOpenDialog = func(ctx context.Context, options runtime.OpenDialogOptions) (string, error) {
		return `C:\docs\sample.md`, nil
	}
	fileBinaryCheck = func(detector *BinaryDetector, filePath string) (bool, error) {
		return false, nil
	}
	fileLoadAndDisplayMD = func(processor *DocumentProcessor, filePath string) error {
		called = true
		return nil
	}

	fm.OpenFileMenuHandler(&menu.CallbackData{}, &currentFile)

	if !called {
		t.Fatal("OpenFileMenuHandler() did not call document loader")
	}
	if currentFile != `C:\docs\sample.md` {
		t.Fatalf("currentFile = %q, want selected path", currentFile)
	}
	if len(*events) != 0 {
		t.Fatalf("captured %d events, want 0", len(*events))
	}
	if len(*dialogs) != 0 {
		t.Fatalf("captured %d dialogs, want 0", len(*dialogs))
	}
}
