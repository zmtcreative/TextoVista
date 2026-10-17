package main

import (
	"io/fs"
	"os"
	"testing"

	"texto-vista/internal/cli"
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
	if appOptions.Width != 1040 || appOptions.Height != 984 {
		t.Fatalf("window size = %dx%d, want 1040x984", appOptions.Width, appOptions.Height)
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
	if appOptions.OnStartup == nil || appOptions.OnDomReady == nil || appOptions.OnShutdown == nil {
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