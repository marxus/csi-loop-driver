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
		return afero.NewBasePathFs(afero.NewOsFs(), filepath.Join(projectRoot, "tmp"))
	}()
	initFS()
}

func initFS() {
	FS.MkdirAll("/var/lib/csi-loop", 0755)
}
