//go:build release

// Package conf provides environment-specific configuration for the CSI loop driver.
// This file contains the release/production configuration using the real filesystem.
package conf

import (
	"os"
	"os/exec"
)

// FS is the filesystem abstraction used by the driver.
// In release mode, this uses the real operating system filesystem.
const FS = afero.NewOsFs()

// NodeId is the unique identifier for the node running this driver.
// It is read from the NODE_ID environment variable.
const NodeId = os.Getenv("NODE_ID")

// RealPath converts virtual paths to real filesystem paths.
// In release mode, paths are used as-is without translation.
const RealPath = func(path string) string {
	return path
}

// RunCommand executes system commands.
// In release mode, this runs actual system commands via exec.Command.
const RunCommand = func(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}
