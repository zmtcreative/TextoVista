package main

// This file was created to expand test coverage without adding more tests to app_test.go.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	internalapp "texto-vista/internal/app"
	"texto-vista/internal/cli"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/menu"
)

type capturedAppEvent struct {
	ctx  context.Context
	name string
	data []interface{}
}

type fakeDocumentWatcher struct {
	addCalls    []string
	removeCalls []string
	closed      bool
	events      chan fsnotify.Event
	errors      chan error
}

func newFakeDocumentWatcher() *fakeDocumentWatcher {
	return &fakeDocumentWatcher{
		events: make(chan fsnotify.Event),
		errors: make(chan error),
	}
}

func (w *fakeDocumentWatcher) Add(name string) error {
	w.addCalls = append(w.addCalls, name)
	return nil
}

func (w *fakeDocumentWatcher) Remove(name string) error {
	w.removeCalls = append(w.removeCalls, name)
	return nil
}

func (w *fakeDocumentWatcher) Close() error {
	w.closed = true
	close(w.events)
	close(w.errors)
	return nil
}

func (w *fakeDocumentWatcher) Events() <-chan fsnotify.Event {
	return w.events
}

func (w *fakeDocumentWatcher) Errors() <-chan error {
	return w.errors
}

type immediateTimer struct{}

func (immediateTimer) Stop() bool {
	return true
}

type countingTimer struct {
	stopCalls int
}

func (t *countingTimer) Stop() bool {
	t.stopCalls++
	return true
}

func newSection1TestApp(t *testing.T) *App {
	t.Helper()

	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })
	os.Args = []string{"test-texto-vista.exe"}
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cliArgs := &cli.CliArgs{
		AppProgName:        stringPtr("test-texto-vista"),
		AppProgNameWithExt: stringPtr("test-texto-vista.exe"),
		CmdlineOptions:     stringPtr("--file <path>"),
	}

	app := NewApp(cliArgs)
	if app == nil {
		t.Fatal("NewApp() returned nil")
	}

	return app
}

func captureAppRuntime(t *testing.T) (*[]capturedAppEvent, *bool) {
	t.Helper()

	events := []capturedAppEvent{}
	quitCalled := false
	originalEventsEmit := appEventsEmit
	originalQuit := appQuit

	appEventsEmit = func(ctx context.Context, eventName string, optionalData ...interface{}) {
		dataCopy := append([]interface{}{}, optionalData...)
		events = append(events, capturedAppEvent{ctx: ctx, name: eventName, data: dataCopy})
	}
	appQuit = func(context.Context) {
		quitCalled = true
	}

	t.Cleanup(func() {
		appEventsEmit = originalEventsEmit
		appQuit = originalQuit
	})

	return &events, &quitCalled
}

func section1ConfigPath() string {
	return filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "test-texto-vista", "test-texto-vista.json")
}

func TestAppStartupInitializesManagers(t *testing.T) {
	app := newSection1TestApp(t)
	ctx := context.Background()

	app.startup(ctx)

	if app.ctx != ctx {
		t.Fatal("startup() did not store context")
	}
	if app.themeManager == nil {
		t.Fatal("startup() did not initialize themeManager")
	}
	if app.printManager == nil {
		t.Fatal("startup() did not initialize printManager")
	}
	if app.binaryDetector == nil {
		t.Fatal("startup() did not initialize binaryDetector")
	}
	if app.documentProcessor == nil {
		t.Fatal("startup() did not initialize documentProcessor")
	}
	if app.fileManager == nil {
		t.Fatal("startup() did not initialize fileManager")
	}
	if got := app.themeManager.GetTheme(); got != "light" {
		t.Fatalf("themeManager.GetTheme() = %q, want %q", got, "light")
	}
}

func TestAppDomReadyWithoutInitialFileEmitsWelcomePayload(t *testing.T) {
	app := newSection1TestApp(t)
	ctx := context.Background()
	app.ctx = ctx
	events, _ := captureAppRuntime(t)

	app.domReady(ctx)

	if len(*events) != 1 {
		t.Fatalf("domReady() emitted %d events, want 1", len(*events))
	}
	if (*events)[0].name != "markdown-rendered" {
		t.Fatalf("first event = %q, want %q", (*events)[0].name, "markdown-rendered")
	}

	renderData, ok := (*events)[0].data[0].(internalapp.MarkdownRenderData)
	if !ok {
		t.Fatalf("markdown-rendered payload type = %T, want MarkdownRenderData", (*events)[0].data[0])
	}
	if renderData.Title != "Welcome to TextoVista!" {
		t.Fatalf("renderData.Title = %q, want welcome title", renderData.Title)
	}
	if !strings.Contains(renderData.HTML, "No File Loaded") {
		t.Fatalf("renderData.HTML missing no-file message: %q", renderData.HTML)
	}
	if !strings.Contains(renderData.HTML, "test-texto-vista --file path/to/your/file.md") {
		t.Fatalf("renderData.HTML missing CLI usage hint: %q", renderData.HTML)
	}
	if !strings.Contains(renderData.FrontmatterHTML, "No frontmatter") {
		t.Fatalf("renderData.FrontmatterHTML = %q, want no-frontmatter placeholder", renderData.FrontmatterHTML)
	}
}

func TestAppDomReadyWithHelpEmitsHelpEvent(t *testing.T) {
	app := newSection1TestApp(t)
	ctx := context.Background()
	app.ctx = ctx
	app.showHelp = true
	events, _ := captureAppRuntime(t)

	app.domReady(ctx)

	if len(*events) != 2 {
		t.Fatalf("domReady() emitted %d events, want 2", len(*events))
	}
	if (*events)[1].name != "show-help" {
		t.Fatalf("help event = %q, want %q", (*events)[1].name, "show-help")
	}
	if got := (*events)[1].data[0]; got != "Command-Line Options" {
		t.Fatalf("show-help title = %v, want %q", got, "Command-Line Options")
	}
	if got := (*events)[1].data[1]; got != "--file <path>" {
		t.Fatalf("show-help body = %v, want %q", got, "--file <path>")
	}
}

func TestAppDomReadyWithInitialFileLoadFailureEmitsError(t *testing.T) {
	app := newSection1TestApp(t)
	ctx := context.Background()
	app.startup(ctx)
	app.currentFile = filepath.Join(t.TempDir(), "missing.md")
	events, _ := captureAppRuntime(t)

	app.domReady(ctx)

	if len(*events) != 1 {
		t.Fatalf("domReady() emitted %d events, want 1", len(*events))
	}
	if (*events)[0].name != "error" {
		t.Fatalf("event name = %q, want %q", (*events)[0].name, "error")
	}
	message, ok := (*events)[0].data[0].(string)
	if !ok {
		t.Fatalf("error payload type = %T, want string", (*events)[0].data[0])
	}
	if !strings.Contains(message, "Failed to load initial file:") {
		t.Fatalf("error message = %q, want initial file prefix", message)
	}
	if !strings.Contains(message, "binary file check failed") {
		t.Fatalf("error message = %q, want binary check failure details", message)
	}
}

func TestAppMenuStructureAndCallbacks(t *testing.T) {
	app := newSection1TestApp(t)
	ctx := context.Background()
	app.ctx = ctx
	app.printManager = internalapp.NewPrintManager(ctx)
	app.fileManager = internalapp.NewFileManager(ctx, nil, nil)
	events, quitCalled := captureAppRuntime(t)

	appMenu := app.menu()
	if appMenu == nil {
		t.Fatal("menu() returned nil")
	}
	if len(appMenu.Items) != 2 {
		t.Fatalf("top-level menu count = %d, want 2", len(appMenu.Items))
	}

	fileMenu := appMenu.Items[0]
	helpMenu := appMenu.Items[1]
	if fileMenu.Label != "File" {
		t.Fatalf("file menu label = %q, want %q", fileMenu.Label, "File")
	}
	if helpMenu.Label != "Help" {
		t.Fatalf("help menu label = %q, want %q", helpMenu.Label, "Help")
	}

	fileLabels := menuLabels(fileMenu.SubMenu)
	wantFileLabels := []string{"Open", "", "Print", "", "Settings", "", "Exit"}
	assertMenuLabels(t, fileLabels, wantFileLabels)

	helpLabels := menuLabels(helpMenu.SubMenu)
	wantHelpLabels := []string{"Command-Line Options", "", "About"}
	assertMenuLabels(t, helpLabels, wantHelpLabels)

	fileMenu.SubMenu.Items[4].Click(&menu.CallbackData{})
	helpMenu.SubMenu.Items[0].Click(&menu.CallbackData{})
	helpMenu.SubMenu.Items[2].Click(&menu.CallbackData{})
	fileMenu.SubMenu.Items[6].Click(&menu.CallbackData{})

	if len(*events) != 3 {
		t.Fatalf("menu callbacks emitted %d events, want 3", len(*events))
	}
	if (*events)[0].name != "show-settings" {
		t.Fatalf("settings callback emitted %q, want %q", (*events)[0].name, "show-settings")
	}
	if (*events)[1].name != "show-help" || (*events)[1].data[0] != "Command-Line Options" {
		t.Fatalf("command-line callback = %#v, want show-help Command-Line Options", (*events)[1])
	}
	if (*events)[2].name != "show-help" || (*events)[2].data[0] != "About" {
		t.Fatalf("about callback = %#v, want show-help About", (*events)[2])
	}
	if !*quitCalled {
		t.Fatal("exit callback did not invoke quit")
	}
}

func TestAppSaveSettingsRecreatesManagersAndSwallowsReloadError(t *testing.T) {
	app := newSection1TestApp(t)
	ctx := context.Background()
	app.ctx = ctx
	app.binaryDetector = internalapp.NewBinaryDetector()
	app.documentProcessor = internalapp.NewDocumentProcessorWithStyle(ctx, app.configManager)
	app.fileManager = internalapp.NewFileManager(ctx, app.binaryDetector, app.documentProcessor)
	app.currentFile = filepath.Join(t.TempDir(), "missing.md")
	oldDocumentProcessor := app.documentProcessor
	oldFileManager := app.fileManager

	settings := *app.GetSettings()
	settings.AlertCallouts.AlertCalloutStyle = "InvalidStyle"

	if err := app.SaveSettings(&settings); err != nil {
		t.Fatalf("SaveSettings() error = %v", err)
	}
	if app.configManager.AlertCalloutStyle() != "GFMPlus" {
		t.Fatalf("AlertCalloutStyle() = %q, want %q", app.configManager.AlertCalloutStyle(), "GFMPlus")
	}
	if app.documentProcessor == oldDocumentProcessor {
		t.Fatal("SaveSettings() did not recreate documentProcessor")
	}
	if app.fileManager == oldFileManager {
		t.Fatal("SaveSettings() did not recreate fileManager")
	}
	if _, err := os.Stat(section1ConfigPath()); err != nil {
		t.Fatalf("SaveSettings() did not persist config file: %v", err)
	}
}

func TestAppSaveSettingsSessionOnlyDoesNotPersistConfig(t *testing.T) {
	app := newSection1TestApp(t)
	ctx := context.Background()
	app.ctx = ctx
	app.binaryDetector = internalapp.NewBinaryDetector()
	app.documentProcessor = internalapp.NewDocumentProcessorWithStyle(ctx, app.configManager)
	app.fileManager = internalapp.NewFileManager(ctx, app.binaryDetector, app.documentProcessor)
	app.currentFile = filepath.Join(t.TempDir(), "missing.md")
	oldDocumentProcessor := app.documentProcessor
	oldFileManager := app.fileManager

	settings := *app.GetSettings()
	settings.AlertCallouts.AlertCalloutStyle = "InvalidStyle"

	if err := app.SaveSettingsSessionOnly(&settings); err != nil {
		t.Fatalf("SaveSettingsSessionOnly() error = %v", err)
	}
	if app.configManager.AlertCalloutStyle() != "GFMPlus" {
		t.Fatalf("AlertCalloutStyle() = %q, want %q", app.configManager.AlertCalloutStyle(), "GFMPlus")
	}
	if app.documentProcessor == oldDocumentProcessor {
		t.Fatal("SaveSettingsSessionOnly() did not recreate documentProcessor")
	}
	if app.fileManager == oldFileManager {
		t.Fatal("SaveSettingsSessionOnly() did not recreate fileManager")
	}
	if _, err := os.Stat(section1ConfigPath()); !os.IsNotExist(err) {
		t.Fatalf("SaveSettingsSessionOnly() unexpectedly persisted config: stat err = %v", err)
	}
}

func TestAppFontSettingsValidationAndPersistence(t *testing.T) {
	app := newSection1TestApp(t)

	if err := app.SetApplicationFont("Atkinson Hyperlegible", 7); err == nil {
		t.Fatal("SetApplicationFont() accepted too-small font size")
	}
	if err := app.SetApplicationMonospaceFont("Cascadia Code", 73); err == nil {
		t.Fatal("SetApplicationMonospaceFont() accepted too-large font size")
	}
	if err := app.SetApplicationFont("Atkinson Hyperlegible", 16); err != nil {
		t.Fatalf("SetApplicationFont() error = %v", err)
	}
	if err := app.SetApplicationMonospaceFont("Cascadia Code", 14); err != nil {
		t.Fatalf("SetApplicationMonospaceFont() error = %v", err)
	}

	currentFont := app.GetCurrentFont()
	if currentFont["fontFamily"] != "Atkinson Hyperlegible" {
		t.Fatalf("GetCurrentFont().fontFamily = %v, want %q", currentFont["fontFamily"], "Atkinson Hyperlegible")
	}
	if currentFont["fontSize"] != 16.0 {
		t.Fatalf("GetCurrentFont().fontSize = %v, want %v", currentFont["fontSize"], 16.0)
	}

	currentMonospaceFont := app.GetCurrentMonospaceFont()
	if currentMonospaceFont["fontFamily"] != "Cascadia Code" {
		t.Fatalf("GetCurrentMonospaceFont().fontFamily = %v, want %q", currentMonospaceFont["fontFamily"], "Cascadia Code")
	}
	if currentMonospaceFont["fontSize"] != 14.0 {
		t.Fatalf("GetCurrentMonospaceFont().fontSize = %v, want %v", currentMonospaceFont["fontSize"], 14.0)
	}

	if _, err := os.Stat(section1ConfigPath()); err != nil {
		t.Fatalf("font settings were not persisted: %v", err)
	}
}

func TestAppAdvancedFontDetectionPersistsState(t *testing.T) {
	app := newSection1TestApp(t)

	if err := app.SetAdvancedFontDetection(true); err != nil {
		t.Fatalf("SetAdvancedFontDetection(true) error = %v", err)
	}
	if !app.GetAdvancedFontDetectionStatus() {
		t.Fatal("GetAdvancedFontDetectionStatus() = false, want true")
	}
	if _, err := os.Stat(section1ConfigPath()); err != nil {
		t.Fatalf("advanced font detection was not persisted: %v", err)
	}
}

func TestAppCurrentFileStateHelpers(t *testing.T) {
	app := newSection1TestApp(t)

	if app.HasCurrentFile() {
		t.Fatal("HasCurrentFile() = true, want false")
	}
	if app.GetCurrentFile() != "" {
		t.Fatalf("GetCurrentFile() = %q, want empty string", app.GetCurrentFile())
	}

	app.currentFile = `C:\docs\guide.md`

	if !app.HasCurrentFile() {
		t.Fatal("HasCurrentFile() = false, want true")
	}
	if app.GetCurrentFile() != `C:\docs\guide.md` {
		t.Fatalf("GetCurrentFile() = %q, want %q", app.GetCurrentFile(), `C:\docs\guide.md`)
	}
}

func TestAppSetCurrentFileStartsWatchingDirectory(t *testing.T) {
	app := newSection1TestApp(t)
	fakeWatcher := newFakeDocumentWatcher()
	originalNewWatcher := newDocumentWatcher
	newDocumentWatcher = func() (documentWatcher, error) {
		return fakeWatcher, nil
	}
	t.Cleanup(func() {
		newDocumentWatcher = originalNewWatcher
	})

	filePath := filepath.Join(t.TempDir(), "guide.md")
	app.setCurrentFile(filePath)

	if app.GetCurrentFile() != filePath {
		t.Fatalf("GetCurrentFile() = %q, want %q", app.GetCurrentFile(), filePath)
	}
	if len(fakeWatcher.addCalls) != 1 || fakeWatcher.addCalls[0] != filepath.Dir(filePath) {
		t.Fatalf("watcher add calls = %#v, want %q", fakeWatcher.addCalls, filepath.Dir(filePath))
	}
	if app.watchedFile != filepath.Clean(filePath) {
		t.Fatalf("watchedFile = %q, want %q", app.watchedFile, filepath.Clean(filePath))
	}
}

func TestAppHandleWatchedFileEventReloadsWhenAutoRefreshEnabled(t *testing.T) {
	app := newSection1TestApp(t)
	filePath := filepath.Join(t.TempDir(), "watched.md")
	app.currentFile = filePath
	app.watchedFile = filepath.Clean(filePath)

	originalAfterFunc := appAfterFunc
	originalLoadMarkdown := appLoadMarkdown
	appAfterFunc = func(_ time.Duration, fn func()) appTimer {
		fn()
		return immediateTimer{}
	}
	reloadCalls := 0
	appLoadMarkdown = func(_ *internalapp.DocumentProcessor, path string) error {
		reloadCalls++
		if path != filePath {
			t.Fatalf("reload path = %q, want %q", path, filePath)
		}
		return nil
	}
	t.Cleanup(func() {
		appAfterFunc = originalAfterFunc
		appLoadMarkdown = originalLoadMarkdown
	})

	app.handleWatchedFileEvent(fsnotify.Event{Name: filePath, Op: fsnotify.Write})

	if reloadCalls != 1 {
		t.Fatalf("reloadCalls = %d, want 1", reloadCalls)
	}
}

func TestAppHandleWatchedFileEventSkipsReloadWhenAutoRefreshDisabled(t *testing.T) {
	app := newSection1TestApp(t)
	filePath := filepath.Join(t.TempDir(), "watched.md")
	settings := *app.GetSettings()
	settings.Application.UseAutoRefresh = false
	app.configManager.SetConfig(&settings)
	app.currentFile = filePath
	app.watchedFile = filepath.Clean(filePath)

	originalAfterFunc := appAfterFunc
	originalLoadMarkdown := appLoadMarkdown
	appAfterFunc = func(_ time.Duration, fn func()) appTimer {
		fn()
		return immediateTimer{}
	}
	reloadCalls := 0
	appLoadMarkdown = func(_ *internalapp.DocumentProcessor, _ string) error {
		reloadCalls++
		return nil
	}
	t.Cleanup(func() {
		appAfterFunc = originalAfterFunc
		appLoadMarkdown = originalLoadMarkdown
	})

	app.handleWatchedFileEvent(fsnotify.Event{Name: filePath, Op: fsnotify.Create})

	if reloadCalls != 0 {
		t.Fatalf("reloadCalls = %d, want 0", reloadCalls)
	}
}

func TestAppScheduleAutoRefreshStopsPreviousTimer(t *testing.T) {
	app := newSection1TestApp(t)

	previousTimer := &countingTimer{}
	app.autoRefreshTimer = previousTimer

	originalAfterFunc := appAfterFunc
	originalLoadMarkdown := appLoadMarkdown
	appAfterFunc = func(_ time.Duration, _ func()) appTimer {
		return immediateTimer{}
	}
	appLoadMarkdown = func(_ *internalapp.DocumentProcessor, _ string) error {
		return nil
	}
	t.Cleanup(func() {
		appAfterFunc = originalAfterFunc
		appLoadMarkdown = originalLoadMarkdown
	})

	app.scheduleAutoRefresh()

	if previousTimer.stopCalls != 1 {
		t.Fatalf("previousTimer.Stop() call count = %d, want 1", previousTimer.stopCalls)
	}
	if app.autoRefreshTimer == nil {
		t.Fatal("autoRefreshTimer = nil, want non-nil")
	}
}

func TestAppScheduleAutoRefreshWithImmediateCallbackDoesNotDeadlock(t *testing.T) {
	app := newSection1TestApp(t)
	filePath := filepath.Join(t.TempDir(), "immediate.md")
	app.currentFile = filePath

	originalAfterFunc := appAfterFunc
	originalLoadMarkdown := appLoadMarkdown
	appAfterFunc = func(_ time.Duration, fn func()) appTimer {
		fn()
		return immediateTimer{}
	}
	appLoadMarkdown = func(_ *internalapp.DocumentProcessor, path string) error {
		if path != filePath {
			t.Fatalf("reload path = %q, want %q", path, filePath)
		}
		return nil
	}
	t.Cleanup(func() {
		appAfterFunc = originalAfterFunc
		appLoadMarkdown = originalLoadMarkdown
	})

	done := make(chan struct{})
	go func() {
		app.scheduleAutoRefresh()
		close(done)
	}()

	select {
	case <-done:
		// expected
	case <-time.After(2 * time.Second):
		t.Fatal("scheduleAutoRefresh() timed out; possible deadlock")
	}
}

func menuLabels(menu *menu.Menu) []string {
	labels := make([]string, 0, len(menu.Items))
	for _, item := range menu.Items {
		if item.IsSeparator() {
			labels = append(labels, "")
			continue
		}
		labels = append(labels, item.Label)
	}
	return labels
}

func assertMenuLabels(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("menu label count = %d, want %d (%v)", len(got), len(want), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("menu label[%d] = %q, want %q", index, got[index], want[index])
		}
	}
}
