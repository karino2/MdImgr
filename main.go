package main

import (
	"embed"
	"os"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

/*
MdImgr.app [targetDir] [template]
*/
func main() {
	targetDirName := ""
	template := ""

	if len(os.Args) > 1 {
		targetDirName = os.Args[1]
	}
	if len(os.Args) > 2 {
		template = os.Args[2]
	}

	println("targetDir:", targetDirName)
	println("template:", template)

	targetDir := &TargetDir{targetDir: targetDirName}
	app := NewApp(targetDir, template)

	AppMenu := menu.NewMenu()
	if runtime.GOOS == "darwin" {
		AppMenu.Append(menu.AppMenu()) // On macOS platform, this must be done right after `NewMenu()`
	}
	FileMenu := AppMenu.AddSubmenu("File")
	FileMenu.AddText("Open Dir", keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
		app.SelectDirAndNotify()
	})

	if runtime.GOOS == "darwin" {
		AppMenu.Append(menu.EditMenu()) // On macOS platform, EditMenu should be appended to enable Cmd+C, Cmd+V, Cmd+Z... shortcuts
	}

	ViewMenu := AppMenu.AddSubmenu("View")
	ViewMenu.AddText("Reload", keys.CmdOrCtrl("r"), func(_ *menu.CallbackData) {
		app.NotifyUpdateImageList()
	})

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "MdImgr",
		Width:  500,
		Height: 768,
		Menu:   AppMenu,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: NewTargetDirLoader(targetDir),
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
