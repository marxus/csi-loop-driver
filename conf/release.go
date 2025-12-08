//go:build release

package conf

import (
	"os"

	"github.com/spf13/afero"
)

var FS = afero.NewOsFs()

var NodeId = os.Getenv("NODE_ID")

func RealPath(path string) string {
	return path
}
