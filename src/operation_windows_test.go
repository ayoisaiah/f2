//go:build windows
// +build windows

package f2

import (
	"syscall"
	"testing"
)

func setHidden(path string) error {
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
	cases := retrieveTestCases(t, "windows.json")
	runTestCases(t, cases)
}
