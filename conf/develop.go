//go:build !release

// Package conf provides environment-specific configuration for the CSI loop driver.
// This file contains the development configuration using a sandboxed filesystem.
package conf

import (
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/afero"
)

// FS is the filesystem abstraction used by the driver.
// In development mode, this uses BasePathFs rooted at project/tmp.
var FS afero.Fs

// NodeId is the unique identifier for the node running this driver.
// In development mode, this defaults to "node-id".
const NodeId = "node-id"

// RealPath converts virtual paths to real filesystem paths.
// In development mode, this prepends the basePath to create paths in project/tmp.
var RealPath func(path string) string

// RunCommand executes system commands.
// In development mode, this runs actual system commands via exec.Command.
var RunCommand = func(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// initDevelop initializes the development environment.
// It sets up a sandboxed filesystem under project/tmp and creates required directories.
func initDevelop() {
	FS = func() afero.Fs {
		_, filename, _, _ := runtime.Caller(0)
		projectRoot := filepath.Dir(filepath.Dir(filename))
		basePath := filepath.Join(projectRoot, "tmp")
		RealPath = func(path string) string { return filepath.Join(basePath, path) }
		return afero.NewBasePathFs(afero.NewOsFs(), basePath)
	}()
	initFS()
}

// initFS creates required directories in the filesystem.
func initFS() {
	FS.MkdirAll("/csi", 0755)
	FS.MkdirAll("/var/lib/csi-loop", 0755)
}
