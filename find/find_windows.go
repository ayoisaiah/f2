//go:build windows
// +build windows

package find

import (
	"path/filepath"
	"strings"
	"syscall"
)

func isUNCPath(path string) bool {
	// UNC paths start with exactly two backslashes, e.g., \\Server\Share
	return strings.HasPrefix(path, `\\`)
}

// checkIfHidden checks if a file is hidden on Windows.
func checkIfHidden(filename, baseDir string) (bool, error) {
	absPath, err := filepath.Abs(filepath.Join(baseDir, filename))
	if err != nil {
		return false, err
	}

	p := `\\?\` + absPath

	if isUNCPath(absPath) {
		p = absPath
	}

	// Appending `\\?\` to the absolute path helps with
	// preventing 'Path Not Specified Error' when accessing
	// long paths and filenames
	// https://docs.microsoft.com/en-us/windows/win32/fileio/maximum-file-path-limitation?tabs=cmd
	pointer, err := syscall.UTF16PtrFromString(p)
	if err != nil {
		return false, err
	}

	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		return false, err
	}

	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
