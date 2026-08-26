// Command muster is a local-first desktop task manager over Backlog.md projects.
package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/FMakareev/muster-backlog/internal/app"
)

// The built frontend is embedded into the binary, so a release is a single file.
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
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

	a.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Muster",
		Width:  1280,
		Height: 800,
		// A board over several projects needs room; below this it stops being readable.
		MinWidth:         960,
		MinHeight:        600,
		BackgroundColour: application.NewRGB(15, 17, 21),
		URL:              "/",
	})

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
