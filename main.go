package main

import (
	"embed"

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
	return &options.App{
		Title:     "TextoVista",
		MinWidth:  616,
		MinHeight: 539,
		// MaxWidth: 1296,
		// MaxHeight: 1020,
		Width:  1040,
		Height: 984,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        mdrApp.startup,
		OnDomReady:       mdrApp.domReady,
		OnShutdown:       mdrApp.shutdown,
		Menu:             mdrApp.menu(),
		Bind: []interface{}{
			mdrApp,
		},
	}
}

func main() {
    // Handle command-line arguments FIRST ---
    cliArgs, _ := cli.GetArgs()

	// Unknown CLI flags are intentionally ignored so the GUI app can tolerate
	// wrapper- or launcher-added arguments while still honoring --file/-f.

    // Create an instance of the mdrApp structure, passing the parsed CLI arguments
	// This will initialize the application with the command-line arguments.
    mdrApp := NewApp(cliArgs)

	// TODO: Add configurable window width/height with screen dimension validation

	// Create application with options
	werr := wails.Run(buildWailsAppOptions(mdrApp))

	if werr != nil {
		println("Error:", werr.Error())
	}
}

