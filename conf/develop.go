//go:build !release

package conf

import (
	"path/filepath"
	"runtime"

	"github.com/spf13/afero"
)

var (
	FS afero.Fs

	NodeId = "node-id"
)

func initDevelop() {
	FS = func() afero.Fs {
		_, filename, _, _ := runtime.Caller(0)
		projectRoot := filepath.Dir(filepath.Dir(filename))
		basePath = filepath.Join(projectRoot, "tmp")
		return afero.NewBasePathFs(afero.NewOsFs(), basePath)
	}()
	initFS()
}

func initFS() {
	FS.MkdirAll("/csi", 0755)
	FS.MkdirAll("/var/lib/csi-loop", 0755)
}

var basePath string

func RealPath(path string) string {
	return filepath.Join(basePath, path)

}
