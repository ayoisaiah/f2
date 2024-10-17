//go:build windows
// +build windows

package rename_test

import (
	"testing"

	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/testutil"
)

func TestRenameWindows(t *testing.T) {
	testCases := []testutil.TestCase{
		{
			Name: "rename with new directory (backslash)",
			Changes: file.Changes{
				{
					Source: "File.txt",
					Target: `new_folder\myFile.txt`,
				},
			},
		},
	}

	renameTest(t, testCases)
}
