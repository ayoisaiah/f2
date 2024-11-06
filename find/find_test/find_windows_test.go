//go:build windows
// +build windows

package find_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/ayoisaiah/f2/v2/internal/testutil"
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

func setupWindowsHidden(t *testing.T, testDir string) (teardown func()) {
	err := filepath.WalkDir(
		testDir,
		func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err // Handle errors gracefully
			}

			if !d.IsDir() && filepath.Base(path)[0] == 46 {
				setHidden((path))
			}

			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	return func() {}
}

var windowsTestCases = []testutil.TestCase{
	{
		Name: "dot files shouldn't be regarded as hidden in Windows",
		Want: []string{
			".hidden_file",
			"backup/documents/.hidden_resume.txt",
			"documents/.hidden_file.txt",
			"photos/vacation/mountains/.hidden_photo.jpg",
		},
		Args: []string{"-f", "hidden", "-R"},
	},

	{
		Name:      "exclude files with hidden attribute",
		Want:      []string{},
		Args:      []string{"-f", "hidden", "-R"},
		SetupFunc: setupWindowsHidden,
	},

	{
		Name: "include files with hidden attribute",
		Want: []string{
			".hidden_file",
			"backup/documents/.hidden_resume.txt",
			"documents/.hidden_file.txt",
			"photos/vacation/mountains/.hidden_photo.jpg",
		},
		Args:      []string{"-f", "hidden", "-RH"},
		SetupFunc: setupWindowsHidden,
	},
}

// TestFindWindows only tests search behaviors perculiar to Windows
func TestFindWindows(t *testing.T) {
	testDir := testutil.SetupFileSystem(t, "find", findFileSystem)

	findTest(t, windowsTestCases, testDir)
}
