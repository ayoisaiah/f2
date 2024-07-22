package replace_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/testutil"
	"github.com/ayoisaiah/f2/replace"
)

func replaceTest(t *testing.T, cases []testutil.TestCase) {
	t.Helper()

	for i := range cases {
		tc := cases[i]

		for j := range tc.Changes {
			ch := tc.Changes[j]

			cases[i].Changes[j].OriginalSource = ch.Source
			cases[i].Changes[j].RelSourcePath = filepath.Join(
				ch.BaseDir,
				ch.Source,
			)
		}
	}

	for i := range cases {
		tc := cases[i]

		t.Run(tc.Name, func(t *testing.T) {
			if tc.SetupFunc != nil {
				t.Cleanup(tc.SetupFunc(t))
			}

			config := testutil.GetConfig(t, &tc, ".")

			changes, err := replace.Replace(config, tc.Changes)
			if err == nil {
				testutil.CompareTargetPath(t, tc.Want, changes)
				return
			}

			if !errors.Is(err, tc.Error) {
				t.Fatal(err)
			}
		})
	}
}

func TestReplace(t *testing.T) {
	testCases := []testutil.TestCase{
		{
			Name: "basic replace",
			Changes: []*file.Change{
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
			Name: "basic replace with positional arguments",
			Changes: []*file.Change{
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
			Args: []string{"macos", "darwin"},
		},
		{
			Name: "replace only the first match",
			Changes: []*file.Change{
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
			Changes: []*file.Change{
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
			Name: "replace the first 2 matches in reverse",
			Changes: []*file.Change{
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
			Changes: []*file.Change{
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
			Changes: []*file.Change{
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
				"music/Overgrown (2013)/01-overgrown.flac",
				"music/Overgrown (2013)/02-i-am-sold.flac",
				"music/Overgrown (2013)/cover.jpg",
				"music/2013/overgrown",
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
			Changes: []*file.Change{
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
	}

	replaceTest(t, testCases)
}
