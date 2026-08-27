// Command muster is a local-first desktop task manager over Backlog.md projects.
package main

import (
	"context"
	"embed"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/FMakareev/muster-backlog/internal/app"
	"github.com/FMakareev/muster-backlog/internal/mcpserver"
	"github.com/FMakareev/muster-backlog/internal/registry"
)

// The built frontend is embedded into the binary, so a release is a single file.
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// One binary, two ways in. `muster mcp` speaks the Model Context Protocol
	// over stdio and never constructs the application, which is what lets an
	// agent use Muster whether or not the window is open - and what an MCP
	// client expects: a process it spawns and talks to over a pipe.
	if len(os.Args) > 1 && os.Args[1] == "mcp" {
		if err := serveMCP(); err != nil && !errors.Is(err, context.Canceled) {
			log.Fatalf("muster mcp: %v", err)
		}
		return
	}

	a := application.New(application.Options{
		Name:        "Muster",
		Description: "A local-first desktop task manager over all your Backlog.md projects at once",
		Services: []application.Service{
			application.NewService(app.NewService()),
			application.NewService(app.NewBoardService()),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	window := a.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:   app.MainWindowName,
		Title:  "Muster",
		Width:  1280,
		Height: 800,
		// A board over several projects needs room; below this it stops being readable.
		MinWidth:         960,
		MinHeight:        600,
		BackgroundColour: application.NewRGB(15, 17, 21),
		URL:              "/",
	})

	// Closing the window leaves the application resident when the tray
	// preference is on. The handler reads the preference each time, so
	// changing it in the interface takes effect without a restart.
	app.InstallCloseBehaviour(window)

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

// serveMCP runs the Model Context Protocol server until the client goes away.
//
// Nothing is drawn and no window exists; the registry is read from the same
// place the application reads it, so an agent sees exactly the projects the
// person registered and no others.
func serveMCP() error {
	server, err := mcpserver.New(registry.DefaultPath())
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return server.Run(ctx)
}
