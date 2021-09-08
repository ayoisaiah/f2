//go:build !windows
// +build !windows

package f2

const pathSeperator = "/"

// isHidden checks if a file is hidden on Unix operating systems
// the error is returned to match the signature of the Windows
// version of the function.
func isHidden(filename, baseDir string) (bool, error) {
	return filename[0] == dotCharacter, nil
}
