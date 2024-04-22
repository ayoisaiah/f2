//go:build windows
// +build windows

package find_test

import (
	"testing"

	"github.com/ayoisaiah/f2/internal/testutil"
)

var windowsTestCases = []testutil.TestCase{
	// {
	// 	// FIXME: Dot files should not be hidden in Windows
	// 	Name: "dot files are hidden in Windows",
	// 	Want: []string{
	// 		// ".hidden_file",
	// 		// "backup/documents/.hidden_old_resume.txt",
	// 		// "documents/.hidden_file.txt",
	// 		// "photos/vacation/mountains/.hidden_old_photo.jpg",
	// 	},
	// 	Args: []string{"-f", "hidden", "-R"},
	// },

	// TODO: Add a test for hidden files
}

// TestFindWindows only tests search behaviors perculiar to Windows
func TestFindWindows(t *testing.T) {
	findTest(t, windowsTestCases)
}
