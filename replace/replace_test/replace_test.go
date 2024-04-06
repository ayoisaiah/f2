package replace_test

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/testutil"
	"github.com/ayoisaiah/f2/replace"
)

func TestMain(m *testing.M) {
	dateFilePath := filepath.Join("testdata", "date.txt")

	_, err := os.Create(dateFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// Update file access and modification times for testdata/date.txt
	// so its always consistent
	atime := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	mtime := time.Date(2019, time.January, 5, 12, 0, 0, 0, time.UTC)

	err = os.Chtimes(dateFilePath, atime, mtime)
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	err = os.Remove(dateFilePath)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}

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
	}

	replaceTest(t, testCases)
}
