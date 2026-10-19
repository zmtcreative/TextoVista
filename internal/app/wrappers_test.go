package app

import (
	"context"
	"errors"
	"testing"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type capturedWrapperEvent struct {
	ctx  context.Context
	name string
	data []interface{}
}

type capturedWrapperDialog struct {
	ctx     context.Context
	options runtime.MessageDialogOptions
}

func captureWrapperRuntime(t *testing.T) (*[]capturedWrapperEvent, *[]capturedWrapperDialog) {
	t.Helper()

	events := []capturedWrapperEvent{}
	dialogs := []capturedWrapperDialog{}
	originalThemeEventsEmit := themeEventsEmit
	originalPrintEventsEmit := printEventsEmit
	originalPrintSaveFileDialog := printSaveFileDialog
	originalErrorEventsEmit := errorEventsEmit
	originalErrorMessageDialog := errorMessageDialog

	eventSink := func(ctx context.Context, eventName string, optionalData ...interface{}) {
		dataCopy := append([]interface{}{}, optionalData...)
		events = append(events, capturedWrapperEvent{ctx: ctx, name: eventName, data: dataCopy})
	}

	themeEventsEmit = eventSink
	printEventsEmit = eventSink
	printSaveFileDialog = func(ctx context.Context, options runtime.SaveDialogOptions) (string, error) {
		return "", nil
	}
	errorEventsEmit = eventSink
	errorMessageDialog = func(ctx context.Context, options runtime.MessageDialogOptions) (string, error) {
		dialogs = append(dialogs, capturedWrapperDialog{ctx: ctx, options: options})
		return "", nil
	}

	t.Cleanup(func() {
		themeEventsEmit = originalThemeEventsEmit
		printEventsEmit = originalPrintEventsEmit
		printSaveFileDialog = originalPrintSaveFileDialog
		errorEventsEmit = originalErrorEventsEmit
		errorMessageDialog = originalErrorMessageDialog
	})

	return &events, &dialogs
}

func TestThemeManagerSetTheme(t *testing.T) {
	ctx := context.Background()
	tm := NewThemeManager(ctx)
	events, dialogs := captureWrapperRuntime(t)

	if tm.GetTheme() != "light" {
		t.Fatalf("GetTheme() = %q, want %q", tm.GetTheme(), "light")
	}

	tm.SetTheme("dark")

	if tm.GetTheme() != "dark" {
		t.Fatalf("GetTheme() after SetTheme() = %q, want %q", tm.GetTheme(), "dark")
	}
	if len(*events) != 1 {
		t.Fatalf("event count = %d, want 1", len(*events))
	}
	if (*events)[0].name != "theme:changed" || (*events)[0].data[0] != "dark" {
		t.Fatalf("event = %#v, want theme:changed dark", (*events)[0])
	}
	if len(*dialogs) != 0 {
		t.Fatalf("dialog count = %d, want 0", len(*dialogs))
	}
}

func TestPrintManagerPrintContent(t *testing.T) {
	pm := NewPrintManager(context.Background())
	events, dialogs := captureWrapperRuntime(t)

	if err := pm.PrintContent(); err != nil {
		t.Fatalf("PrintContent() error = %v", err)
	}
	if len(*events) != 1 {
		t.Fatalf("event count = %d, want 1", len(*events))
	}
	if (*events)[0].name != "print-content" {
		t.Fatalf("event name = %q, want %q", (*events)[0].name, "print-content")
	}
	if len(*dialogs) != 0 {
		t.Fatalf("dialog count = %d, want 0", len(*dialogs))
	}
}

func TestPrintManagerPrintContentToPDF(t *testing.T) {
	tests := []struct {
		name          string
		dialogPath    string
		dialogErr     error
		wantErr       string
		wantEventName string
		wantEventData string
	}{
		{
			name:      "dialog error is returned",
			dialogErr: errors.New("dialog failed"),
			wantErr:   "dialog failed",
		},
		{
			name:       "empty path is treated as cancel",
			dialogPath: "",
		},
		{
			name:          "selected path emits save event",
			dialogPath:    `C:\docs\output.pdf`,
			wantEventName: "save-as-pdf",
			wantEventData: `C:\docs\output.pdf`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPrintManager(context.Background())
			events, _ := captureWrapperRuntime(t)
			printSaveFileDialog = func(ctx context.Context, options runtime.SaveDialogOptions) (string, error) {
				if options.Title != "Save as PDF" || options.DefaultFilename != "document.pdf" {
					t.Fatalf("save dialog options = %#v", options)
				}
				if len(options.Filters) != 1 || options.Filters[0].Pattern != "*.pdf" {
					t.Fatalf("save dialog filters = %#v", options.Filters)
				}
				return tt.dialogPath, tt.dialogErr
			}

			err := pm.PrintContentToPDF()
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("PrintContentToPDF() error = %v, want %q", err, tt.wantErr)
				}
			} else if err != nil {
				t.Fatalf("PrintContentToPDF() error = %v", err)
			}

			if tt.wantEventName == "" {
				if len(*events) != 0 {
					t.Fatalf("event count = %d, want 0", len(*events))
				}
				return
			}

			if len(*events) != 1 {
				t.Fatalf("event count = %d, want 1", len(*events))
			}
			if (*events)[0].name != tt.wantEventName || (*events)[0].data[0] != tt.wantEventData {
				t.Fatalf("event = %#v, want %s %s", (*events)[0], tt.wantEventName, tt.wantEventData)
			}
		})
	}
}

func TestErrorHandlerHandleError(t *testing.T) {
	tests := []struct {
		name           string
		showDialog     bool
		wantEvent      bool
		wantDialog     bool
		wantDialogText string
	}{
		{
			name:       "event path emits error event",
			showDialog: false,
			wantEvent:  true,
		},
		{
			name:           "dialog path shows message dialog",
			showDialog:     true,
			wantDialog:     true,
			wantDialogText: "Visible to user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eh := NewErrorHandler(context.Background())
			events, dialogs := captureWrapperRuntime(t)

			eh.HandleError(errors.New("boom"), "Visible to user", tt.showDialog)

			if tt.wantEvent {
				if len(*events) != 1 {
					t.Fatalf("event count = %d, want 1", len(*events))
				}
				if (*events)[0].name != "error" || (*events)[0].data[0] != "Visible to user" {
					t.Fatalf("event = %#v, want error event with user message", (*events)[0])
				}
			} else if len(*events) != 0 {
				t.Fatalf("event count = %d, want 0", len(*events))
			}

			if tt.wantDialog {
				if len(*dialogs) != 1 {
					t.Fatalf("dialog count = %d, want 1", len(*dialogs))
				}
				if (*dialogs)[0].options.Title != "Error" || (*dialogs)[0].options.Message != tt.wantDialogText {
					t.Fatalf("dialog = %#v, want Error dialog with user message", (*dialogs)[0].options)
				}
			} else if len(*dialogs) != 0 {
				t.Fatalf("dialog count = %d, want 0", len(*dialogs))
			}
		})
	}
}
