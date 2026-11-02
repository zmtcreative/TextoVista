package main

import (
	"embed"

	"texto-vista/internal/app"
	"texto-vista/internal/cli"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

var (
	Version = "dev-build"
	Commit  = "none"
	Date    = "unknown"
)

func buildWailsAppOptions(mdrApp *App) *options.App {
	var savedState app.WindowStateConfig
	if mdrApp != nil && mdrApp.configManager != nil {
		savedState = mdrApp.configManager.GetWindowState()
	}

	workAreaW, workAreaH, err := app.GetAvailableWorkArea()
	if err != nil || workAreaW <= 0 || workAreaH <= 0 {
		workAreaW = app.MaxWidthDefault + app.WindowMarginX
		workAreaH = app.MaxHeightDefault + app.WindowMarginY
	}

	width, height, isMaximized := app.CalculateWindowDimensions(savedState, workAreaW, workAreaH)

	startState := options.Normal
	if isMaximized {
		startState = options.Maximised
	}

	appOpts := &options.App{
		Title:            "TextoVista",
		MinWidth:         app.MinWidthDefault,
		MinHeight:        app.MinHeightDefault,
		Width:            width,
		Height:           height,
		WindowStartState: startState,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		Bind: []interface{}{
			mdrApp,
		},
	}

	if mdrApp != nil {
		appOpts.OnStartup = mdrApp.startup
		appOpts.OnDomReady = mdrApp.domReady
		appOpts.OnShutdown = mdrApp.shutdown
		appOpts.OnBeforeClose = mdrApp.beforeClose
		appOpts.Menu = mdrApp.menu()
	}

	return appOpts
}

func main() {
	// Handle command-line arguments FIRST ---
	cliArgs, _ := cli.GetArgs()

	// Unknown CLI flags are intentionally ignored so the GUI app can tolerate
	// wrapper- or launcher-added arguments while still honoring --file/-f.

	// Create an instance of the mdrApp structure, passing the parsed CLI arguments
	// This will initialize the application with the command-line arguments.
	mdrApp := NewApp(cliArgs)

	// Create application with options
	werr := wails.Run(buildWailsAppOptions(mdrApp))

	if werr != nil {
		println("Error:", werr.Error())
	}
}
