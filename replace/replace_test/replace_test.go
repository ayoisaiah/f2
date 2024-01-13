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
			Name: "replace with auto incrementing integers",
			Changes: []*file.Change{
				{
					Source: "a.txt",
				},
				{
					Source: "b.txt",
				},
				{
					Source: "c.txt",
				},
			},
			Want: []string{"1.txt", "2.txt", "3.txt"},
			Args: []string{"-f", "a|b|c", "-r", "{%d}"},
		},
		{
			Name: "replace with multiple incrementing integers",
			Changes: []*file.Change{
				{
					Source: "a.txt",
				},
				{
					Source: "b.txt",
				},
				{
					Source: "c.txt",
				},
			},
			Want: []string{"1_10_0100.txt", "2_20_0200.txt", "3_30_0300.txt"},
			Args: []string{"-f", "a|b|c", "-r", "{%d}_{10%02d10}_{100%04d100}"},
		},
		{
			Name: "skip numbers",
			Changes: []*file.Change{
				{
					Source: "a.txt",
				},
				{
					Source: "b.txt",
				},
				{
					Source: "c.txt",
				},
			},
			Want: []string{"1_10_0100.txt", "2_20_0200.txt", "3_30_0300.txt"},
			Args: []string{"-f", "a|b|c", "-r", "{%d}_{10%02d10}_{100%04d100}"},
		},
	}

	replaceTest(t, testCases)
}
