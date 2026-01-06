package main

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed bin/* lists/*
var binaryAssets embed.FS

func main() {
	appInstance := NewApp()

	err := wails.Run(&options.App{
		Title:             "Liberty Engine",
		Width:             420,
		Height:            680,
		DisableResize:     true,
		Fullscreen:        false,
		Frameless:         true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 9, G: 9, B: 11, A: 255}, // #09090b
		OnStartup:        appInstance.startup,
		Bind: []interface{}{
			appInstance,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func unpack(dest string) error {
	return fs.WalkDir(binaryAssets, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil { return err }
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dest, path), 0755)
		}
		data, _ := binaryAssets.ReadFile(path)
		return os.WriteFile(filepath.Join(dest, path), data, 0755)
	})
}