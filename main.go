package main

import (
	"embed"

	"synx/internal/app"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed desktop/frontend/*
var assets embed.FS

func main() {
	synxApp, err := app.New(assets)
	if err != nil {
		panic(err)
	}

	err = wails.Run(&options.App{
		Title:            "SynX",
		Width:            1180,
		Height:           760,
		MinWidth:         980,
		MinHeight:        620,
		BackgroundColour: &options.RGBA{R: 10, G: 11, B: 16, A: 1},
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        synxApp.Startup,
		OnShutdown:       synxApp.Shutdown,
		Bind:             []interface{}{synxApp},
	})
	if err != nil {
		panic(err)
	}
}
