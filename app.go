package main

import (
	"context"
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

const iniDir = "/Users/arinokazuma/work/GitHub/MdImgr/tests"
const template = `![images/WeightMedianFilter/$1]({{"/assets/images/WeightMedianFilter/$1" | absolute_url}})`

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

func (t *TargetDir) DropFile(srcPath string) error {
	now := time.Now()

	fname := now.Format("2006_0102_150405") + ".png"

	destPath := filepath.Join(t.targetDir, fname)

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

func (t *TargetDir) RegisterFileDropListener(ctx context.Context) {
	println("register!")
	runtime.OnFileDrop(ctx, func(x, y int, paths []string) {
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
		runtime.EventsEmit(ctx, "image-list-update")
	})
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

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.targetDir.RegisterFileDropListener(ctx)
}

func (a *App) ListFiles() []string {
	return a.targetDir.ListFiles()
}

func (a *App) CopyUrl(fname string) {
	txt := strings.ReplaceAll(template, "$1", fname)
	runtime.ClipboardSetText(a.ctx, txt)
}
