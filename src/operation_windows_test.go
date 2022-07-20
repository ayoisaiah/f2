//go:build windows
// +build windows

package f2

import (
	"syscall"
	"testing"
)

func setWindowsHidden(path string) error {
	filenameW, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	err = syscall.SetFileAttributes(filenameW, syscall.FILE_ATTRIBUTE_HIDDEN)
	if err != nil {
		return err
	}

	return nil
}

func TestWindows(t *testing.T) {
	cases := h2(t, "windows.json")
	h(t, cases)
}
