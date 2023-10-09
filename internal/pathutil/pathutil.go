package pathutil

import (
	"path/filepath"
)

// StripExtension returns the input file name without its extension.
func StripExtension(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}
