package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// const iniDir = "/Users/arinokazuma/work/GitHub/MdImgr/tests"
const iniDir = "/Users/arinokazuma/work/GitHub/karino2.github.io/assets/images/MLAA"
const template = `![images/MLAA/$1]({{"/assets/images/MLAA/$1" | absolute_url}})`

type TargetDir struct {
	targetDir string
}

func NewTargetDir() *TargetDir {
	return &TargetDir{targetDir: iniDir}
}

func (t *TargetDir) ReadFile(fname string) ([]byte, error) {
	requestedPath := path.Join(t.targetDir, fname)
	return os.ReadFile(requestedPath)
}

func hasImageExt(fname string) bool {
	// only support png for a while.
	return filepath.Ext(fname) == ".png"
}

/*
List image files.
*/
func (t *TargetDir) ListFiles() []string {
	println("list file:", t.targetDir)
	var result []string
	result = make([]string, 0)
	files, err := os.ReadDir(t.targetDir)
	if err != nil {
		println("list file fails:", err)
		return result
	}
	for _, file := range files {
		if !file.IsDir() {
			if hasImageExt(file.Name()) {
				result = append(result, file.Name())
			}
		}
	}
	return result
}

func (t *TargetDir) NewFilePath() string {
	now := time.Now()

	fname := now.Format("2006_0102_150405") + ".png"

	return filepath.Join(t.targetDir, fname)
}

func (t *TargetDir) DropFile(srcPath string) error {
	destPath := t.NewFilePath()

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to create src file: %s, %w", srcPath, err)
	}

	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create dest file: %s, %w", destPath, err)
	}

	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	return nil
}

func (t *TargetDir) SaveImageData(data string) error {
	destPath := t.NewFilePath()

	parts := strings.SplitN(data, ",", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid base64 data url")
	}
	base64Data := parts[1]

	decodedImage, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("fail to decode base64: %w", err)
	}

	err = os.WriteFile(destPath, decodedImage, 0644)
	if err != nil {
		return fmt.Errorf("fail to save: %w", err)
	}

	fmt.Printf("Saved: %s, bytesize: %d.", destPath, len(decodedImage))

	return nil
}

func (t *TargetDir) OnDropFiles(paths []string) {
	println("ondrop!")
	for _, p := range paths {
		if hasImageExt(p) {
			err := t.DropFile(p)
			if err != nil {
				println("DropFile Error:", err)
				return
			}
			// notify update.
		}
	}
}

type TargetDirLoader struct {
	http.Handler
	targetDir *TargetDir
}

func NewTargetDirLoader(td *TargetDir) *TargetDirLoader {
	return &TargetDirLoader{targetDir: td}
}

func (h *TargetDirLoader) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var err error
	requestedFilename := strings.TrimPrefix(req.URL.Path, "/")
	println("Requesting file:", requestedFilename)

	fileData, err := h.targetDir.ReadFile(requestedFilename)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		res.Write(fmt.Appendf(nil, "Could not load file %s", requestedFilename))
	}

	res.Write(fileData)
}

type App struct {
	ctx       context.Context
	targetDir *TargetDir
}

func NewApp(td *TargetDir) *App {
	return &App{targetDir: td}
}

func (a *App) NotifyUpdateImageList() {
	runtime.EventsEmit(a.ctx, "image-list-update")
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	runtime.OnFileDrop(ctx, func(x, y int, paths []string) {
		a.targetDir.OnDropFiles(paths)
		a.NotifyUpdateImageList()
	})
}

func (a *App) ListFiles() []string {
	return a.targetDir.ListFiles()
}

func (a *App) CopyUrl(fname string) {
	txt := strings.ReplaceAll(template, "$1", fname)
	runtime.ClipboardSetText(a.ctx, txt)
}

func (a *App) SaveImage(data string) {
	a.targetDir.SaveImageData(data)

	a.NotifyUpdateImageList()
}
