package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"text/template"
	"time"

	"texto-vista/internal/app"

	"texto-vista/internal/cli"

	"github.com/fsnotify/fsnotify"
	"github.com/tidwall/gjson"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed wails.json
var wailsConfig string

//go:embed frontend/src/assets/html/about.gohtml
var aboutTemplate string

//go:embed frontend/src/assets/html/license-short.html
var licenseShort string

var (
	appEventsEmit   = wailsruntime.EventsEmit
	appQuit         = wailsruntime.Quit
	appLoadMarkdown = func(processor *app.DocumentProcessor, filePath string) error {
		return processor.LoadAndDisplayMarkdown(filePath)
	}
	newDocumentWatcher = func() (documentWatcher, error) {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return nil, err
		}
		return &fsnotifyDocumentWatcher{watcher: watcher}, nil
	}
	appAfterFunc = func(delay time.Duration, fn func()) appTimer {
		return time.AfterFunc(delay, fn)
	}
)

const autoRefreshDebounce = 250 * time.Millisecond

type appTimer interface {
	Stop() bool
}

type documentWatcher interface {
	Add(name string) error
	Remove(name string) error
	Close() error
	Events() <-chan fsnotify.Event
	Errors() <-chan error
}

type fsnotifyDocumentWatcher struct {
	watcher *fsnotify.Watcher
}

func (w *fsnotifyDocumentWatcher) Add(name string) error {
	return w.watcher.Add(name)
}

func (w *fsnotifyDocumentWatcher) Remove(name string) error {
	return w.watcher.Remove(name)
}

func (w *fsnotifyDocumentWatcher) Close() error {
	return w.watcher.Close()
}

func (w *fsnotifyDocumentWatcher) Events() <-chan fsnotify.Event {
	return w.watcher.Events
}

func (w *fsnotifyDocumentWatcher) Errors() <-chan error {
	return w.watcher.Errors
}

// App struct
type App struct {
	ctx                context.Context
	currentFile        string
	appProgName        string            // Store the application name without extension
	appProgNameWithExt string            // Store the application name with extension
	showHelp           bool              // Flag to indicate if help should be shown
	cmdlineOptions     string            // Store command line options here
	versionInfo        string            // Store version information

	// Managers
	themeManager      *app.ThemeManager
	printManager      *app.PrintManager
	fileManager       *app.FileManager
	documentProcessor *app.DocumentProcessor
	binaryDetector    *app.BinaryDetector
	configManager     *app.ConfigManager
	fontManager       *app.FontManager

	watchMu          sync.Mutex
	fileWatcher      documentWatcher
	watchedDir       string
	watchedFile      string
	autoRefreshTimer appTimer
}

// NewApp creates a new App application struct
func NewApp(cliArgs *cli.CliArgs) *App {
	// Initialize configuration manager first
	configManager := app.NewConfigManager()

	// Apply CLI overrides to configuration
	//   (Removed these from CliArgs [2025-08-20])
	//   No need to deal with overrides for now -- leave as is in case we need it in the future
	// configManager.ApplyCliOverrides(cliArgs.AllowInlineHTML, cliArgs.SanitizeHTML, cliArgs.StripH1)

	// Get final configuration after overrides
	// finalConfig := configManager.GetConfig()

	// Handle app name
	appProgNameWithExt := stringFromPtr(cliArgs.AppProgNameWithExt, "texto-vista")
	setAboutString, err := setAbout(appProgNameWithExt)
	if err != nil {
		// Log the error and use a fallback version string
		log.Printf("##> LOG: Error generating about string: %v", err)
		setAboutString = fmt.Sprintf("<h1>%s</h1><p>Version: %s</p>", appProgNameWithExt, Version)
	}

	return &App{
		currentFile:        stringFromPtr(cliArgs.InitialFile, ""),             // Default to empty, can be set via CLI flag
		appProgName:        stringFromPtr(cliArgs.AppProgName, "texto-vista"), // Store the application name without extension
		appProgNameWithExt: appProgNameWithExt,                              // Store the application name with extension
		showHelp:           boolFromPtr(cliArgs.ShowHelp, false),            // Default to false, can be set via CLI flag
		versionInfo:        setAboutString,                                  // Set version info using the application name with extension
		cmdlineOptions:     stringFromPtr(cliArgs.CmdlineOptions, ""),       // Store the command line options for help display
		configManager:      configManager,                                   // Store the config manager
		fontManager:        app.NewFontManager(configManager),               // Initialize font manager with config manager
	}
}

// boolFromPtr safely dereferences a *bool, returning a default value if the pointer is nil.
func boolFromPtr(p *bool, defaultValue bool) bool {
	if p != nil {
		return *p
	}
	return defaultValue
}

func stringFromPtr(p *string, defaultValue string) string {
	if p != nil {
		return *p
	}
	return defaultValue
}

func setAbout(appProgNameWithExt string) (string, error) {
	var versionText bytes.Buffer

	authorName := gjson.Get(wailsConfig, "author.name").String()
	// authorEmail := gjson.Get(wailsConfig, "author.email").String()
	productName := gjson.Get(wailsConfig, "info.productName").String()

	tplData := struct {
		ProductName string
		AppName     string
		Version     string
		BuildDate   string
		Copyright   string
		License     string
	}{
		ProductName: productName,
		AppName:     appProgNameWithExt,
		Version:     Version,
		BuildDate:   Date,
		Copyright:   fmt.Sprintf("Copyright 2026 %s", authorName),
		License:     licenseShort,
	}
	tpl, err := template.New("about").Parse(aboutTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing about template: %w", err)
	}
	err = tpl.Execute(&versionText, tplData)
	if err != nil {
		return "", fmt.Errorf("executing about template: %w", err)
	}

	return versionText.String(), nil
}

// startup is called when the app starts. The context is created
// after the app is created, but before the event loop starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize managers
	a.themeManager = app.NewThemeManager(ctx)
	a.printManager = app.NewPrintManager(ctx)
	a.binaryDetector = app.NewBinaryDetector()

	// Get the current configuration for document processor
	a.documentProcessor = app.NewDocumentProcessorWithStyle(ctx, a.configManager)
	a.fileManager = app.NewFileManager(ctx, a.binaryDetector, a.documentProcessor)
}

func initialFileLoadErrorMessage(err error) string {
	return "Failed to load initial file: " + err.Error()
}

// domReady is called after the frontend loads the DOM.
// This is where we load and display the initial Markdown file if provided via CLI.
func (a *App) domReady(ctx context.Context) {
	// Create structured data for the frontend
	renderData := app.MarkdownRenderData{
		HTML:            "",
		Title:           "",
		Date:            "",
		FrontmatterHTML: "",
	}

	if a.currentFile != "" {
		log.Printf("##> LOG: Loading initial file from command line: %s", a.currentFile)
		err := a.fileManager.LoadFile(a.currentFile)
		if err != nil {
			log.Printf("##> LOG: Error loading initial Markdown file %q: %v", a.currentFile, err)
			appEventsEmit(a.ctx, "error", initialFileLoadErrorMessage(err))
		} else if syncErr := a.syncFileWatcher(); syncErr != nil {
			log.Printf("##> LOG: Warning: failed to start file watcher for %q: %v", a.currentFile, syncErr)
		}
	} else {
		// Emit a welcome message if no initial file is provided
		renderData.Title = "Welcome to TextoVista!"
		welcomeHTML := "<h1>" + renderData.Title + "</h1>" +
			"<h2>No File Loaded</h2>" +
			"<p>Open a Markdown file using the <code>File &gt; Open</code> menu option or provide a path via the command line (e.g., <code>" +
			a.appProgName +
			" --file path/to/your/file.md</code>).</p>"
		renderData.HTML = welcomeHTML
		renderData.FrontmatterHTML = `<div class="frontmatter-container"><div class="frontmatter-header">No frontmatter</div></div>`
		appEventsEmit(a.ctx, "markdown-rendered", renderData)
	}
	if a.showHelp {
		appEventsEmit(a.ctx, "show-help", "Command-Line Options", a.cmdlineOptions)
	}
}

// shutdown is called when the app is about to exit.
// Perform any cleanup here if necessary.
func (a *App) shutdown(ctx context.Context) {
	a.closeFileWatcher()
	log.Println("##> LOG: Application is shutting down.")
}

func (a *App) menu() *menu.Menu {
	// Create the application menu
	appMenu := menu.NewMenu()

	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("Open", keys.CmdOrCtrl("o"), func(data *menu.CallbackData) {
		previousFile := a.GetCurrentFile()
		a.fileManager.OpenFileMenuHandler(data, &a.currentFile)
		if a.currentFile != "" && a.currentFile != previousFile {
			a.setCurrentFile(a.currentFile)
		}
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Print", keys.CmdOrCtrl("p"), func(_ *menu.CallbackData) {
		a.printManager.PrintContent()
	})
	// fileMenu.AddText("Save as PDF", keys.CmdOrCtrl("e"), func(_ *menu.CallbackData) {
	//     a.printManager.PrintContentToPDF()
	// })
	fileMenu.AddSeparator()
	fileMenu.AddText("Settings", keys.CmdOrCtrl("comma"), func(_ *menu.CallbackData) {
		appEventsEmit(a.ctx, "show-settings")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Exit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		appQuit(a.ctx)
	})

	// --- Add a new Help menu ---
	helpMenu := appMenu.AddSubmenu("Help")
	helpMenu.AddText("Command-Line Options", keys.CmdOrCtrl("h"), func(_ *menu.CallbackData) {
		// Emit an event to the frontend, sending the help text as data.
		appEventsEmit(a.ctx, "show-help", "Command-Line Options", a.cmdlineOptions)
	})
	helpMenu.AddSeparator()
	helpMenu.AddText("About", keys.CmdOrCtrl("a"), func(_ *menu.CallbackData) {
		// Emit an event to the frontend, sending the version information as data.
		appEventsEmit(a.ctx, "show-help", "About", a.versionInfo)
	})

	return appMenu
}

// Delegate methods to managers

// PrintContent prints the current HTML content
func (a *App) PrintContent() error {
	return a.printManager.PrintContent()
}

// PrintContentToPDF exports the current content to PDF (Windows-specific)
func (a *App) PrintContentToPDF() error {
	return a.printManager.PrintContentToPDF()
}

// GetTheme returns the current theme
func (a *App) GetTheme() string {
	return a.themeManager.GetTheme()
}

// SetTheme sets the theme and emits an event to the frontend
func (a *App) SetTheme(theme string) {
	a.themeManager.SetTheme(theme)
}

// LoadAndDisplayMarkdown loads and displays a markdown file
func (a *App) LoadAndDisplayMarkdown(filePath string) error {
	err := appLoadMarkdown(a.documentProcessor, filePath)
	if err == nil {
		a.setCurrentFile(filePath)
	}
	return err
}

// ReloadCurrentDocument reloads and regenerates the currently opened markdown document
func (a *App) ReloadCurrentDocument() error {
	return a.reloadCurrentDocument()
}

// Settings-related methods

// GetSettings returns the current application settings
func (a *App) GetSettings() *app.Config {
	return a.configManager.GetConfig()
}

// GetAlertCalloutStyles returns the available alert callout styles
func (a *App) GetAlertCalloutStyles() map[string]string {
	return app.AlertCalloutStyles
}

// SaveSettings saves the provided settings configuration
func (a *App) SaveSettings(settings *app.Config) error {
	// Validate the alert callout style
	settings.AlertCallouts.AlertCalloutStyle = a.configManager.ValidateAlertCalloutStyle(settings.AlertCallouts.AlertCalloutStyle)

	// Update the configuration
	a.configManager.SetConfig(settings)

	// Save to file
	if err := a.configManager.SaveConfig(); err != nil {
		return err
	}

	// Update the current app settings

	// Recreate the document processor and file manager under the watcher lock to
	// prevent a data race with the file-watcher goroutine that reads these fields.
	a.watchMu.Lock()
	a.documentProcessor = app.NewDocumentProcessorWithStyle(a.ctx, a.configManager)
	a.fileManager = app.NewFileManager(a.ctx, a.binaryDetector, a.documentProcessor)
	a.watchMu.Unlock()

	// Automatically reload the current document to apply new settings
	if a.currentFile != "" {
		if reloadErr := a.ReloadCurrentDocument(); reloadErr != nil {
			log.Printf("##> LOG: Warning: Failed to reload document after settings change: %v", reloadErr)
			// Don't return the reload error, as settings were successfully saved
		}
	}

	return nil
}

// SaveSettingsSessionOnly applies the provided settings to the current session without saving to disk
func (a *App) SaveSettingsSessionOnly(settings *app.Config) error {
	// Validate the alert callout style
	settings.AlertCallouts.AlertCalloutStyle = a.configManager.ValidateAlertCalloutStyle(settings.AlertCallouts.AlertCalloutStyle)

	// Update the configuration in memory only (do NOT call SaveConfig)
	a.configManager.SetConfig(settings)

	// Update the current app settings

	// Recreate the document processor and file manager under the watcher lock to
	// prevent a data race with the file-watcher goroutine that reads these fields.
	a.watchMu.Lock()
	a.documentProcessor = app.NewDocumentProcessorWithStyle(a.ctx, a.configManager)
	a.fileManager = app.NewFileManager(a.ctx, a.binaryDetector, a.documentProcessor)
	a.watchMu.Unlock()

	// Automatically reload the current document to apply new settings
	if a.currentFile != "" {
		if reloadErr := a.ReloadCurrentDocument(); reloadErr != nil {
			log.Printf("##> LOG: Warning: Failed to reload document after settings change: %v", reloadErr)
			// Don't return the reload error, as settings were successfully applied
		}
	}

	return nil
}

// GetAvailableFonts returns a list of available system fonts
func (a *App) GetAvailableFonts() []string {
	return a.fontManager.GetSystemFonts()
}

// SetApplicationFont updates the font family and size settings
func (a *App) SetApplicationFont(fontFamily string, fontSize float64) error {
	// Validate font size (reasonable range: 8-72 pixels)
	if fontSize < 8 || fontSize > 72 {
		return fmt.Errorf("font size must be between 8 and 72 pixels, got: %.1f", fontSize)
	}

	// Update configuration
	a.configManager.SetFontFamily(fontFamily)
	a.configManager.SetFontSize(fontSize)

	// Save the configuration
	if err := a.configManager.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save font configuration: %w", err)
	}

	return nil
}

// GetCurrentFont returns the current font family and size
func (a *App) GetCurrentFont() map[string]interface{} {
	return map[string]interface{}{
		"fontFamily": a.configManager.GetFontFamily(),
		"fontSize":   a.configManager.GetFontSize(),
	}
}

// GetAvailableMonospaceFonts returns a list of available monospace fonts
func (a *App) GetAvailableMonospaceFonts() []string {
	return a.fontManager.GetMonospaceFonts()
}

// SetApplicationMonospaceFont updates the monospace font family and size settings
func (a *App) SetApplicationMonospaceFont(fontFamily string, fontSize float64) error {
	// Validate font size (reasonable range: 8-72 pixels)
	if fontSize < 8 || fontSize > 72 {
		return fmt.Errorf("monospace font size must be between 8 and 72 pixels, got: %.1f", fontSize)
	}

	// Update configuration
	a.configManager.SetFontFamilyMono(fontFamily)
	a.configManager.SetFontSizeMono(fontSize)

	// Save the configuration
	if err := a.configManager.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save monospace font configuration: %w", err)
	}

	return nil
}

// GetCurrentMonospaceFont returns the current monospace font family and size
func (a *App) GetCurrentMonospaceFont() map[string]interface{} {
	return map[string]interface{}{
		"fontFamily": a.configManager.GetFontFamilyMono(),
		"fontSize":   a.configManager.GetFontSizeMono(),
	}
}

// GetAdvancedFontDetectionStatus returns the current advanced font detection setting
func (a *App) GetAdvancedFontDetectionStatus() bool {
	return a.configManager.GetUseAdvancedFontDetection()
}

// GetCurrentFile returns the current file path or empty string if none
func (a *App) GetCurrentFile() string {
	a.watchMu.Lock()
	defer a.watchMu.Unlock()
	return a.currentFile
}

// HasCurrentFile returns true if a file is currently loaded
func (a *App) HasCurrentFile() bool {
	return a.GetCurrentFile() != ""
}

// SetAdvancedFontDetection enables or disables advanced font detection
func (a *App) SetAdvancedFontDetection(enabled bool) error {
	a.configManager.SetUseAdvancedFontDetection(enabled)

	// Save the configuration
	if err := a.configManager.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save advanced font detection setting: %w", err)
	}

	return nil
}

// GetMonospaceFontsWithDetectionInfo returns monospace fonts with detection method information
func (a *App) GetMonospaceFontsWithDetectionInfo() map[string]interface{} {
	isAdvancedEnabled := a.configManager.GetUseAdvancedFontDetection()

	basicFonts := a.fontManager.BasicMonospaceDetection()
	advancedFonts := a.fontManager.AdvancedMonospaceDetection()

	return map[string]interface{}{
		"fonts": a.fontManager.GetMonospaceFonts(),
		"detectionMode": map[string]interface{}{
			"isAdvancedEnabled": isAdvancedEnabled,
			"basicCount":        len(basicFonts),
			"advancedCount":     len(advancedFonts),
			"currentMode": func() string {
				if isAdvancedEnabled {
					return "advanced"
				}
				return "basic"
			}(),
		},
	}
}

func (a *App) reloadCurrentDocument() error {
	a.watchMu.Lock()
	currentFile := a.currentFile
	processor := a.documentProcessor
	a.watchMu.Unlock()

	if currentFile == "" {
		return fmt.Errorf("no document currently loaded")
	}

	if err := appLoadMarkdown(processor, currentFile); err != nil {
		return fmt.Errorf("failed to reload document %s: %w", currentFile, err)
	}

	return nil
}

func (a *App) setCurrentFile(filePath string) {
	a.watchMu.Lock()
	a.currentFile = filePath
	a.watchMu.Unlock()

	if err := a.syncFileWatcher(); err != nil {
		log.Printf("##> LOG: Warning: failed to sync file watcher for %q: %v", filePath, err)
	}
}

func (a *App) syncFileWatcher() error {
	a.watchMu.Lock()
	defer a.watchMu.Unlock()

	if a.currentFile == "" {
		if a.fileWatcher != nil && a.watchedDir != "" {
			_ = a.fileWatcher.Remove(a.watchedDir)
		}
		a.watchedDir = ""
		a.watchedFile = ""
		return nil
	}

	if a.fileWatcher == nil {
		watcher, err := newDocumentWatcher()
		if err != nil {
			return fmt.Errorf("failed to create file watcher: %w", err)
		}
		a.fileWatcher = watcher
		go a.runFileWatcher(watcher)
	}

	targetFile := filepath.Clean(a.currentFile)
	targetDir := filepath.Dir(targetFile)

	if a.watchedDir != "" && a.watchedDir != targetDir {
		if err := a.fileWatcher.Remove(a.watchedDir); err != nil && !os.IsNotExist(err) {
			log.Printf("##> LOG: Warning: failed to remove watch for %q: %v", a.watchedDir, err)
		}
		a.watchedDir = ""
	}

	if a.watchedDir != targetDir {
		if err := a.fileWatcher.Add(targetDir); err != nil {
			return fmt.Errorf("failed to watch directory %s: %w", targetDir, err)
		}
		a.watchedDir = targetDir
	}

	a.watchedFile = targetFile
	return nil
}

func (a *App) closeFileWatcher() {
	a.watchMu.Lock()
	defer a.watchMu.Unlock()

	if a.autoRefreshTimer != nil {
		a.autoRefreshTimer.Stop()
		a.autoRefreshTimer = nil
	}

	if a.fileWatcher != nil {
		if err := a.fileWatcher.Close(); err != nil {
			log.Printf("##> LOG: Warning: failed to close file watcher: %v", err)
		}
		a.fileWatcher = nil
	}

	a.watchedDir = ""
	a.watchedFile = ""
}

func (a *App) runFileWatcher(watcher documentWatcher) {
	for {
		select {
		case event, ok := <-watcher.Events():
			if !ok {
				return
			}
			a.handleWatchedFileEvent(event)
		case err, ok := <-watcher.Errors():
			if !ok {
				return
			}
			log.Printf("##> LOG: Warning: file watcher error: %v", err)
		}
	}
}

func (a *App) handleWatchedFileEvent(event fsnotify.Event) {
	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Remove) == 0 {
		return
	}

	a.watchMu.Lock()
	currentFile := a.watchedFile
	a.watchMu.Unlock()

	if !sameDocumentPath(currentFile, event.Name) {
		return
	}

	a.scheduleAutoRefresh()
}

func (a *App) scheduleAutoRefresh() {
	newTimer := appAfterFunc(autoRefreshDebounce, func() {
		if !a.configManager.UseAutoRefresh() {
			return
		}

		if err := a.reloadCurrentDocument(); err != nil {
			log.Printf("##> LOG: Warning: auto refresh failed for %q: %v", a.GetCurrentFile(), err)
		}
	})

	a.watchMu.Lock()
	previousTimer := a.autoRefreshTimer
	a.autoRefreshTimer = newTimer
	a.watchMu.Unlock()

	if previousTimer != nil {
		previousTimer.Stop()
	}
}

func sameDocumentPath(left, right string) bool {
	cleanLeft := filepath.Clean(left)
	cleanRight := filepath.Clean(right)

	if goruntime.GOOS == "windows" {
		return strings.EqualFold(cleanLeft, cleanRight)
	}

	return cleanLeft == cleanRight
}
