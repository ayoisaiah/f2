//go:build darwin
// +build darwin

package replace_test

import (
	"testing"

	"github.com/ayoisaiah/f2/v2/internal/file"
	"github.com/ayoisaiah/f2/v2/internal/testutil"
)

func TestMacOSTransforms(t *testing.T) {
	testCases := []testutil.TestCase{
		{
			Name: "remove macOS disallowed characters",
			Changes: file.Changes{
				{
					Source: "report:::project*details|on<2024/01/11>.txt",
				},
			},
			Want: []string{"reportproject*details|on<2024/01/11>.txt"},
			Args: []string{"-f", ".*", "-r", "{.mac}"},
		},
	}

	replaceTest(t, testCases)
}
