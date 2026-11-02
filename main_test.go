package main

import (
	"context"
	"io/fs"
	"os"
	"testing"

	internalapp "texto-vista/internal/app"
	"texto-vista/internal/cli"

	"github.com/wailsapp/wails/v2/pkg/options"
)

func newMainSection9App(t *testing.T) *App {
	t.Helper()

	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })
	os.Args = []string{"texto-vista.exe"}
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cliArgs := &cli.CliArgs{
		AppProgName:        stringPtr("texto-vista"),
		AppProgNameWithExt: stringPtr("texto-vista.exe"),
		CmdlineOptions:     stringPtr("--file <path>"),
	}

	app := NewApp(cliArgs)
	if app == nil {
		t.Fatal("NewApp() returned nil")
	}

	return app
}

func TestBuildWailsAppOptions(t *testing.T) {
	mdrApp := newMainSection9App(t)
	appOptions := buildWailsAppOptions(mdrApp)

	if appOptions == nil {
		t.Fatal("buildWailsAppOptions() returned nil")
	}
	if appOptions.Title != "TextoVista" {
		t.Fatalf("Title = %q, want %q", appOptions.Title, "TextoVista")
	}
	if appOptions.MinWidth != 616 || appOptions.MinHeight != 539 {
		t.Fatalf("min size = %dx%d, want 616x539", appOptions.MinWidth, appOptions.MinHeight)
	}
	if appOptions.Width > 1040 || appOptions.Width < 616 {
		t.Fatalf("window width = %d, want between 616 and 1040", appOptions.Width)
	}
	if appOptions.Height > 984 || appOptions.Height < 539 {
		t.Fatalf("window height = %d, want between 539 and 984", appOptions.Height)
	}
	if appOptions.BackgroundColour == nil {
		t.Fatal("BackgroundColour is nil")
	}
	if appOptions.BackgroundColour.R != 27 || appOptions.BackgroundColour.G != 38 || appOptions.BackgroundColour.B != 54 || appOptions.BackgroundColour.A != 1 {
		t.Fatalf("BackgroundColour = %#v, want RGBA{27,38,54,1}", appOptions.BackgroundColour)
	}
	if appOptions.AssetServer == nil || appOptions.AssetServer.Assets == nil {
		t.Fatal("AssetServer.Assets is nil")
	}
	if _, err := fs.ReadDir(appOptions.AssetServer.Assets, "."); err != nil {
		t.Fatalf("AssetServer.Assets is not readable: %v", err)
	}
	if appOptions.OnStartup == nil || appOptions.OnDomReady == nil || appOptions.OnShutdown == nil || appOptions.OnBeforeClose == nil {
		t.Fatal("lifecycle callbacks were not wired")
	}
	if appOptions.Menu == nil {
		t.Fatal("Menu is nil")
	}
	if len(appOptions.Menu.Items) != 2 {
		t.Fatalf("top-level menu count = %d, want 2", len(appOptions.Menu.Items))
	}
	if len(appOptions.Bind) != 1 {
		t.Fatalf("Bind length = %d, want 1", len(appOptions.Bind))
	}
	if boundApp, ok := appOptions.Bind[0].(*App); !ok || boundApp != mdrApp {
		t.Fatalf("Bind[0] = %#v, want original *App", appOptions.Bind[0])
	}
}

func TestBuildWailsAppOptionsMenuLabels(t *testing.T) {
	mdrApp := newMainSection9App(t)
	appOptions := buildWailsAppOptions(mdrApp)

	if got := appOptions.Menu.Items[0].Label; got != "File" {
		t.Fatalf("first menu label = %q, want %q", got, "File")
	}
	if got := appOptions.Menu.Items[1].Label; got != "Help" {
		t.Fatalf("second menu label = %q, want %q", got, "Help")
	}

	fileMenu := appOptions.Menu.Items[0].SubMenu
	helpMenu := appOptions.Menu.Items[1].SubMenu
	if fileMenu == nil || helpMenu == nil {
		t.Fatal("submenu wiring is incomplete")
	}
	if len(fileMenu.Items) != 7 {
		t.Fatalf("file menu item count = %d, want 7", len(fileMenu.Items))
	}
	if len(helpMenu.Items) != 3 {
		t.Fatalf("help menu item count = %d, want 3", len(helpMenu.Items))
	}
	if fileMenu.Items[0].Label != "Open" || fileMenu.Items[2].Label != "Print" || fileMenu.Items[4].Label != "Settings" || fileMenu.Items[6].Label != "Exit" {
		t.Fatalf("unexpected File menu labels: %#v", []string{fileMenu.Items[0].Label, fileMenu.Items[2].Label, fileMenu.Items[4].Label, fileMenu.Items[6].Label})
	}
	if helpMenu.Items[0].Label != "Command-Line Options" || helpMenu.Items[2].Label != "About" {
		t.Fatalf("unexpected Help menu labels: %#v", []string{helpMenu.Items[0].Label, helpMenu.Items[2].Label})
	}
	if fileMenu.Items[0].Click == nil || fileMenu.Items[2].Click == nil || fileMenu.Items[4].Click == nil || fileMenu.Items[6].Click == nil {
		t.Fatal("one or more File menu callbacks are nil")
	}
	if helpMenu.Items[0].Click == nil || helpMenu.Items[2].Click == nil {
		t.Fatal("one or more Help menu callbacks are nil")
	}
}

func TestBuildWailsAppOptionsRestoresSavedWindowState(t *testing.T) {
	mdrApp := newMainSection9App(t)
	mdrApp.configManager.SetWindowState(internalapp.WindowStateConfig{
		Width:     920,
		Height:    740,
		Maximized: false,
	})

	appOptions := buildWailsAppOptions(mdrApp)
	if appOptions.Width != 920 || appOptions.Height != 740 {
		t.Fatalf("restored window dimensions = %dx%d, want 920x740", appOptions.Width, appOptions.Height)
	}
	if appOptions.WindowStartState != options.Normal {
		t.Fatalf("restored window start state = %v, want Normal", appOptions.WindowStartState)
	}
}

func TestBuildWailsAppOptionsRestoresMaximizedState(t *testing.T) {
	mdrApp := newMainSection9App(t)
	mdrApp.configManager.SetWindowState(internalapp.WindowStateConfig{
		Width:     1000,
		Height:    800,
		Maximized: true,
	})

	appOptions := buildWailsAppOptions(mdrApp)
	if appOptions.WindowStartState != options.Maximised {
		t.Fatalf("restored window start state = %v, want Maximised", appOptions.WindowStartState)
	}
}

func TestAppBeforeCloseSavesWindowState(t *testing.T) {
	mdrApp := newMainSection9App(t)
	ctx := context.WithValue(context.Background(), "frontend", "mock")

	origIsMax := appWindowIsMaximised
	origGetSize := appWindowGetSize
	defer func() {
		appWindowIsMaximised = origIsMax
		appWindowGetSize = origGetSize
	}()

	appWindowIsMaximised = func(ctx context.Context) bool { return false }
	appWindowGetSize = func(ctx context.Context) (int, int) { return 880, 680 }

	mdrApp.beforeClose(ctx)

	state := mdrApp.configManager.GetWindowState()
	if state.Width != 880 || state.Height != 680 || state.Maximized {
		t.Fatalf("saved window state = %#v, want 880x680 unmaximized", state)
	}

	// Close while maximized: maximized flag is saved while unmaximized dimensions are preserved
	appWindowIsMaximised = func(ctx context.Context) bool { return true }
	mdrApp.beforeClose(ctx)

	stateMax := mdrApp.configManager.GetWindowState()
	if !stateMax.Maximized || stateMax.Width != 880 || stateMax.Height != 680 {
		t.Fatalf("saved maximized window state = %#v, want maximized with preserved normal size 880x680", stateMax)
	}
}

func TestAppStartupClampsNegativeYPosition(t *testing.T) {
	mdrApp := newMainSection9App(t)
	ctx := context.WithValue(context.Background(), "frontend", "mock")

	origIsMax := appWindowIsMaximised
	origGetPos := appWindowGetPosition
	origSetPos := appWindowSetPosition
	defer func() {
		appWindowIsMaximised = origIsMax
		appWindowGetPosition = origGetPos
		appWindowSetPosition = origSetPos
	}()

	appWindowIsMaximised = func(ctx context.Context) bool { return false }
	appWindowGetPosition = func(ctx context.Context) (int, int) { return 100, -50 }
	var setX, setY int
	calledSetPos := false
	appWindowSetPosition = func(ctx context.Context, x, y int) {
		calledSetPos = true
		setX = x
		setY = y
	}

	mdrApp.startup(ctx)

	if !calledSetPos {
		t.Fatal("expected WindowSetPosition to be called when y < 0")
	}
	if setX != 100 || setY != 0 {
		t.Fatalf("WindowSetPosition called with (%d, %d), want (100, 0)", setX, setY)
	}
}
