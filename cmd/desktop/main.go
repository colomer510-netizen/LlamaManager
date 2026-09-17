package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"llamamanager/internal/web"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Start LlamaManager's web server in background
	go web.StartWebServer()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "LlamaManager V4",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 30, G: 30, B: 46, A: 1},
		Bind: []interface{}{},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
