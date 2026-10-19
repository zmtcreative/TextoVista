package app

import (
	"context"
	"log"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	errorEventsEmit    = runtime.EventsEmit
	errorMessageDialog = runtime.MessageDialog
)

// ErrorHandler centralizes error handling and user notifications
type ErrorHandler struct {
	ctx context.Context
}

// NewErrorHandler creates a new ErrorHandler
func NewErrorHandler(ctx context.Context) *ErrorHandler {
	return &ErrorHandler{ctx: ctx}
}

// HandleError logs the error and optionally shows it to the user
func (eh *ErrorHandler) HandleError(err error, userMessage string, showDialog bool) {
	log.Printf("##> LOG: Error: %v", err)

	if showDialog {
		errorMessageDialog(eh.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Error",
			Message: userMessage,
		})
	} else {
		errorEventsEmit(eh.ctx, "error", userMessage)
	}
}
