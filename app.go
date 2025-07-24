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

const template = `![images/MFG_BasicShape/$1]({{"/assets/images/MFG_BasicShape/$1" | absolute_url}})`

type TargetDir struct {
	targetDir string
}

func NewTargetDir() *TargetDir {
	return &TargetDir{targetDir: ""}
}

func (t *TargetDir) SetTargetDir(path string) {
	t.targetDir = path
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

func (t *TargetDir) toPath(fname string) string { return filepath.Join(t.targetDir, fname) }

func (t *TargetDir) NewFilePath() (string, string) {
	now := time.Now()

	fname := now.Format("2006_0102_150405") + ".png"

	return fname, t.toPath(fname)
}

func (t *TargetDir) DropFile(srcPath string) (string, error) {
	fname, destPath := t.NewFilePath()

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to create src file: %s, %w", srcPath, err)
	}

	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create dest file: %s, %w", destPath, err)
	}

	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}
	return fname, nil
}

func (t *TargetDir) SaveImageData(data string) (string, error) {
	fname, destPath := t.NewFilePath()

	parts := strings.SplitN(data, ",", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid base64 data url")
	}
	base64Data := parts[1]

	decodedImage, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("fail to decode base64: %w", err)
	}

	err = os.WriteFile(destPath, decodedImage, 0644)
	if err != nil {
		return "", fmt.Errorf("fail to save: %w", err)
	}

	fmt.Printf("Saved: %s, bytesize: %d.", destPath, len(decodedImage))

	return fname, nil
}

// return last fname
func (t *TargetDir) OnDropFiles(paths []string) string {
	lastFname := ""
	for _, p := range paths {
		if hasImageExt(p) {
			fname, err := t.DropFile(p)
			if err != nil {
				println("DropFile Error:", err)
				return ""
			}
			lastFname = fname
		}
	}
	return lastFname
}

func (t *TargetDir) DeleteFile(fname string) error {
	path := t.toPath(fname)
	trashDir := t.toPath("Trash")

	err := os.MkdirAll(trashDir, 0755)
	if err != nil {
		return err
	}

	destPath := filepath.Join(trashDir, fname)
	return os.Rename(path, destPath)
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
		last := a.targetDir.OnDropFiles(paths)
		a.NotifyUpdateImageList()
		if last != "" {
			a.CopyUrl(last)
		}
	})
}

func (a *App) ListFiles() []string {
	return a.targetDir.ListFiles()
}

func (a *App) CopyUrl(fname string) {
	txt := strings.ReplaceAll(template, "$1", fname)
	runtime.ClipboardSetText(a.ctx, txt)
	runtime.EventsEmit(a.ctx, "show-toast", "Url copied.")
}

func (a *App) DeleteFile(fname string) {
	err := a.targetDir.DeleteFile(fname)
	if err != nil {
		msg := fmt.Sprintf("Delete fail: %v", err)
		runtime.EventsEmit(a.ctx, "show-toast", msg)
		return
	}
	a.NotifyUpdateImageList()
	runtime.EventsEmit(a.ctx, "show-toast", "Move file to trash dir.")
}

func (a *App) SelectDir() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{})
	if err != nil {
		println("Error selecting directory:", err)
		return ""
	}
	return dir
}

func (a *App) SetTargetDir(path string) {
	a.targetDir.SetTargetDir(path)
	a.NotifyUpdateImageList()
}

func (a *App) SaveImage(data string) {
	fname, err := a.targetDir.SaveImageData(data)
	if err == nil {
		a.CopyUrl(fname)
		a.NotifyUpdateImageList()
	}
}
