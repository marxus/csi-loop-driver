//go:build !release

package conf

import (
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

func initTesting() {
	// In testing, use in-memory filesystem
	FS = afero.NewMemMapFs()
	initFS()
}
