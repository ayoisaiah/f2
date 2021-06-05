// +build windows

package f2

import (
	"path/filepath"
	"syscall"
)

const pathSeperator = `\`

// isHidden checks if a file is hidden on Windows.
func isHidden(filename, baseDir string) (bool, error) {
	// dotfiles also count as hidden
	if filename[0] == dotCharacter {
		return true, nil
	}

	pointer, err := syscall.UTF16PtrFromString(filepath.Join(baseDir, filename))
	if err != nil {
		return false, err
	}

	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		return false, err
	}

	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
