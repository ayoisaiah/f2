//go:build windows
// +build windows

package find_test

import (
	"testing"

	"github.com/ayoisaiah/f2/internal/testutil"
)

var windowsTestCases = []testutil.TestCase{
	{
		Name: "dot files are not hidden in Windows",
		Want: []string{
			".hidden_file",
			"backup/documents/.hidden_old_resume.txt",
			"documents/.hidden_file.txt",
			"photos/vacation/mountains/.hidden_old_photo.jpg",
		},
		Args: []string{"-f", "hidden", "-R"},
	},
}

// TestFindWindows only tests search behaviors perculiar to Windows
func TestFindWindows(t *testing.T) {
	findTest(t, windowsTestCases)
}
