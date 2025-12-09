//go:build !release

// Package conf provides environment-specific configuration for the CSI loop driver.
// This file contains the testing configuration using an in-memory filesystem.
package conf

import (
	"fmt"
	"testing"

	"github.com/spf13/afero"
)

var isTestRun = testing.Testing()

func init() {
	if !isTestRun {
		initDevelop()
	} else {
		initTesting()
	}
}

// initTesting initializes the testing environment.
// It sets up an in-memory filesystem and mocks RunCommand to fail by default.
// Tests should override RunCommand with their own mock implementations.
func initTesting() {
	FS = afero.NewMemMapFs()
	initFS()

	RealPath = func(path string) string {
		return path
	}

	RunCommand = func(name string, args ...string) error {
		return fmt.Errorf("RunCommand not mocked in test: %s %v", name, args)
	}
}
