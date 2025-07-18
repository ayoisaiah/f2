package replace_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ayoisaiah/f2/v2/internal/file"
	"github.com/ayoisaiah/f2/v2/internal/sortfiles"
	"github.com/ayoisaiah/f2/v2/internal/testutil"
	"github.com/ayoisaiah/f2/v2/replace"
)

func replaceTest(t *testing.T, cases []testutil.TestCase) {
	t.Helper()

	testutil.ProcessTestCaseChanges(t, cases)

	for i := range cases {
		tc := cases[i]

		if strings.Contains(tc.Name, "pair") {
			sortfiles.Pairs(tc.Changes, []string{})
		}

		testutil.RunTestCase(
			t,
			&tc,
			func(t *testing.T, tc *testutil.TestCase) {
				t.Helper()

				conf := testutil.GetConfig(t, tc, ".")

				changes, err := replace.Replace(conf, tc.Changes)
				if err == nil {
					testutil.CompareTargetPath(t, tc.Want, changes)
					return
				}

				if !errors.Is(err, tc.Error) {
					t.Fatal(err)
				}
			},
		)
	}
}

func TestReplace(t *testing.T) {
	testCases := []testutil.TestCase{
		{
			Name: "basic replace",
			Changes: file.Changes{
				{
					Source: "macos_update_notes_2023.txt",
				},
				{
					Source: "macos_user_guide_macos_sierra.pdf",
				},
			},
			Want: []string{
				"darwin_update_notes_2023.txt",
				"darwin_user_guide_darwin_sierra.pdf",
			},
			Args: []string{"-f", "macos", "-r", "darwin"},
		},
		{
			Name: "replace only the first match",
			Changes: file.Changes{
				{
					Source: "budget_budget_budget_2023.xlsx",
				},
			},
			Want: []string{
				"forecast_budget_budget_2023.xlsx",
			},
			Args: []string{"-f", "budget", "-r", "forecast", "-l", "1"},
		},
		{
			Name: "replace the first 2 matches in reverse",
			Changes: file.Changes{
				{
					Source: "budget_budget_budget_2023.xlsx",
				},
				{
					Source: "budget_2024.xlsx",
				},
			},
			Want: []string{
				"budget_forecast_forecast_2023.xlsx",
				"forecast_2024.xlsx",
			},
			Args: []string{"-f", "budget", "-r", "forecast", "-l", "-2"},
		},
		{
			Name: "rename with capture variables",
			Changes: file.Changes{
				{
					Source: "dsc-001.arw",
				},
				{
					Source: "dsc-002.arw",
				},
			},
			Want: []string{
				"001-dsc.arw",
				"002-dsc.arw",
			},
			Args: []string{"-f", "(dsc)(-)(\\d+)", "-r", "$3$2$1"},
		},
		{
			Name: "use capture variables in replacement chain",
			Changes: file.Changes{
				{
					BaseDir: "music",
					Source:  "Overgrown (2013)",
					IsDir:   true,
				},
				{
					Source:  "01 Overgrown.flac",
					BaseDir: "music/Overgrown (2013)",
				},
				{
					Source:  "02 I Am Sold.flac",
					BaseDir: "music/Overgrown (2013)",
				},
				{
					Source:  "Cover.jpg",
					BaseDir: "music/Overgrown (2013)",
				},
			},
			Want: []string{
				"music/2013/overgrown",
				"music/Overgrown (2013)/01-overgrown.flac",
				"music/Overgrown (2013)/02-i-am-sold.flac",
				"music/Overgrown (2013)/cover.jpg",
			},
			Args: []string{
				"-f",
				".*",
				"-r",
				"{.lw}",
				"-f",
				"\\s",
				"-r",
				"-",
				"-f",
				"([a-z]+)-\\((2\\d+)\\)",
				"-r",
				"$2/$1",
				"-deR",
			},
		},
		{
			Name: "transform capture variables",
			Changes: file.Changes{
				{
					BaseDir: "ebooks",
					Source:  "atomic-habits.pdf",
				},
				{
					BaseDir: "ebooks",
					Source:  "animal-farm.epub",
				},
			},
			Want: []string{
				"ebooks/ATOMIC-HABITS.Pdf",
				"ebooks/ANIMAL-FARM.Epub",
			},
			Args: []string{"-f", "(.*)\\.(.*)", "-r", "{<$1>.up}.{<$2>.ti}"},
		},
		{
			Name: "rename file pairs",
			Changes: file.Changes{
				{
					Source: "image.dng",
				},
				{
					Source: "image.heif",
				},
				{
					Source: "image.jpg",
				},
				{
					Source: "image.xmp",
				},
				{
					Source: "some_image.jpg",
				},
			},
			Want: []string{
				"picture-001.dng",
				"picture-001.heif",
				"picture-001.jpg",
				"picture-001.xmp",
				"picture-002.jpg",
			},
			Args: []string{"-f", ".*", "-r", "picture-{%03d}", "--pair"},
		},
		{
			Name: "multiple file pairs",
			Changes: file.Changes{
				{
					Source: "image.dng",
				},
				{
					Source: "image.heif",
				},
				{
					Source: "image.jpg",
				},
				{
					Source: "some_image.jpg",
				},
				{
					Source: "some_image.xmp",
				},
			},
			Want: []string{
				"picture-001.dng",
				"picture-001.heif",
				"picture-001.jpg",
				"picture-002.jpg",
				"picture-002.xmp",
			},
			Args: []string{"-f", ".*", "-r", "picture-{%03d}", "--pair"},
		},
		{
			Name: "rename file pairs with a different target directory",
			Changes: file.Changes{
				{
					Source:    "image.dng",
					TargetDir: "pictures",
				},
				{
					Source:    "image.heif",
					TargetDir: "pictures",
				},
				{
					Source:    "image.jpg",
					TargetDir: "pictures",
				},
			},
			Want: []string{
				"pictures/picture-001.dng",
				"pictures/picture-001.heif",
				"pictures/picture-001.jpg",
			},
			Args: []string{
				"-f",
				".*",
				"-r",
				"picture-{%03d}",
				"--pair",
				"-t",
				"pictures",
			},
		},
		{
			Name: "move each file two levels up",
			Changes: file.Changes{
				{
					BaseDir: "dev/scripts/js",
					Source:  "main.js",
				},
				{
					BaseDir: "dev/scripts/js",
					Source:  "server.js",
				},
			},
			Want: []string{
				"dev/main.js",
				"dev/server.js",
			},
			Args: []string{
				"-f",
				".*",
				"-r",
				"../../{f}{ext}",
			},
		},
	}

	replaceTest(t, testCases)
}
