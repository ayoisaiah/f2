//go:build windows
// +build windows

package replace_test

import (
	"testing"

	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/testutil"
)

func TestWindowsTransforms(t *testing.T) {
	testCases := []testutil.TestCase{
		{
			Name: "remove windows disallowed characters",
			Changes: file.Changes{
				{
					Source: "report:::project*details|on<2024/01/11>.txt",
				},
			},
			Want: []string{"reportprojectdetailson20240111.txt"},
			Args: []string{"-f", ".*", "-r", "{.win}"},
		},
	}

	replaceTest(t, testCases)
}
