package path

import (
	"os"
	"runtime"

	"github.com/ayoisaiah/f2/internal/utils"
)

// Collection represents a collection of paths and their respective
// contents.
type Collection map[string][]os.DirEntry

// Separator represents the filepath separator.
var Separator = "/"

func init() {
	if runtime.GOOS == utils.Windows {
		Separator = `\`
	}
}
