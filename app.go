package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const iniDir = "/Users/arinokazuma/work/GitHub/MdImgr/tests"

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
			// only support png for a while.
			if filepath.Ext(file.Name()) == ".png" {
				result = append(result, file.Name())
			}
		}
	}
	return result
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
}

func (a *App) ListFiles() []string {
	return a.targetDir.ListFiles()
}
