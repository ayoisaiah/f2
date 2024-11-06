//go:build windows
// +build windows

package validate_test

import (
	"testing"

	"github.com/ayoisaiah/f2/v2/internal/file"
	"github.com/ayoisaiah/f2/v2/internal/status"
	"github.com/ayoisaiah/f2/v2/internal/testutil"
)

func TestValidateWindows(t *testing.T) {
	t.Helper()

	testCases := []testutil.TestCase{
		{
			Name: "detect trailing period conflict in file names",
			Changes: file.Changes{
				{
					Source:  "index.js",
					Target:  "main.js..",
					BaseDir: "dev",
					Status:  status.TrailingPeriod,
				},
			},
			ConflictDetected: true,
		},
		{
			Name: "detect trailing period conflict in directories",
			Changes: file.Changes{
				{
					Source:  "No Pressure (2021) S1.E1.1080p.mkv",
					Target:  "2021.../No Pressure (2021) S1.E1.1080p.mkv",
					BaseDir: "movies",
					Status:  status.TrailingPeriod,
				},
				{
					Source:  "No Pressure (2021) S1.E2.1080p.mkv",
					Target:  "2021.../No Pressure (2021) S1.E2.1080p.mkv",
					BaseDir: "movies",
					Status:  status.TrailingPeriod,
				},
				{
					Source:  "No Pressure (2021) S1.E3.1080p.mkv",
					Target:  "2021.../No Pressure (2021) S1.E3.1080p.mkv",
					BaseDir: "movies",
					Status:  status.TrailingPeriod,
				},
			},
			ConflictDetected: true,
		},
		{
			Name: "detect forbidden characters in filename",
			Changes: file.Changes{
				{
					Source:  "atomic-habits.pdf",
					Target:  "<>:?etc.pdf",
					BaseDir: "ebooks",
					Status:  status.ForbiddenCharacters,
				},
			},
			ConflictDetected: true,
		},
		{
			Name: "detect filename longer than 255 characters",
			Changes: file.Changes{
				{
					Source:  "1984.pdf",
					Target:  "It was a bright cold day in April, and the clocks were striking thirteen. Winston Smith, his chin nuzzled into his breast in an effort to escape the vile wind, slipped quickly through the glass doors of Victory Mansions, though not quickly enough to prevent a swirl of gritty dust from entering along with him.pdf",
					BaseDir: "ebooks",
					Status:  status.FilenameLengthExceeded,
				},
			},
			ConflictDetected: true,
		},
		{
			Name: "up to 255 emoji characters should not cause a conflict",
			Changes: file.Changes{
				{
					Source:  "1984.pdf",
					Target:  "😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀",
					BaseDir: "ebooks",
				},
			},
			Want: []string{
				"ebooks/😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀",
			},
		},
		{
			Name: "auto fix forbidden characters in filename",
			Changes: file.Changes{
				{
					Source:  "atomic-habits.pdf",
					Target:  "<>:?etc.pdf",
					BaseDir: "ebooks",
				},
			},
			Want: []string{"ebooks/etc.pdf"},
			Args: []string{"-r", "", "-F"},
		},
		{
			Name: "auto fix filename longer than maximum length conflict",
			Changes: file.Changes{
				{
					Source:  "1984.pdf",
					Target:  "It was a bright cold day in April, and the clocks were striking thirteen. Winston Smith, his chin nuzzled into his breast in an effort to escape the vile wind, slipped quickly through the glass doors of Victory Mansions, though not quickly enough to prevent a swirl of gritty dust from entering along with him.pdf",
					BaseDir: "ebooks",
				},
			},
			Want: []string{
				"ebooks/It was a bright cold day in April, and the clocks were striking thirteen. Winston Smith, his chin nuzzled into his breast in an effort to escape the vile wind, slipped quickly through the glass doors of Victory Mansions, though not quickly enough to p.pdf",
			},
			Args: []string{"-r", "", "-F"},
		},
	}

	validateTest(t, testCases)
}
