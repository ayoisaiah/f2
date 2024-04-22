package validate_test

import (
	"path/filepath"
	"testing"

	"github.com/ayoisaiah/f2/internal/conflict"
	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/testutil"
	"github.com/ayoisaiah/f2/validate"
)

var autoFixArgs = []string{"-r", "", "-F"}

func validateTest(t *testing.T, cases []testutil.TestCase) {
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
			cases[i].Changes[j].RelTargetPath = filepath.Join(
				ch.BaseDir,
				ch.Target,
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

			conflicts := validate.Validate(
				tc.Changes,
				config.AutoFixConflicts,
				config.AllowOverwrites,
			)

			testutil.CompareConflicts(t, tc.Conflicts, conflicts)

			if len(tc.Want) != 0 {
				// tc.Changes is modified in place when auto fixing conflicts
				testutil.CompareTargetPath(t, tc.Want, tc.Changes)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	testCases := []testutil.TestCase{
		{
			Name: "detect empty filename conflict",
			Changes: []*file.Change{
				{
					Source:  "1984.pdf",
					Target:  "",
					BaseDir: "ebooks",
				},
			},
			Conflicts: conflict.Collection{
				conflict.EmptyFilename: []conflict.Conflict{
					{
						Sources: []string{"ebooks/1984.pdf"},
						Target:  "ebooks",
					},
				},
			},
		},
		{
			Name: "detect overwriting newly renamed path conflict",
			Changes: []*file.Change{
				{
					Source:  "index.js",
					Target:  "index.svelte",
					BaseDir: "dev",
				},
				{
					Source:  "index.ts",
					Target:  "index.svelte",
					BaseDir: "dev",
				},
			},
			Conflicts: conflict.Collection{
				conflict.OverwritingNewPath: []conflict.Conflict{
					{
						Sources: []string{"dev/index.js", "dev/index.ts"},
						Target:  "dev/index.svelte",
					},
				},
			},
		},
		{
			Name: "report conflict when target path exists but changes AFTER the overwriting file is renamed",
			Changes: []*file.Change{
				{
					Source:  "dsc-001.arw",
					Target:  "dsc-002.arw",
					BaseDir: "testdata/images",
				},
				{
					Source:  "dsc-002.arw",
					Target:  "dsc-003.arw",
					BaseDir: "testdata/images",
				},
			},
			Conflicts: conflict.Collection{
				conflict.FileExists: []conflict.Conflict{
					{
						Sources: []string{"testdata/images/dsc-001.arw"},
						Target:  "testdata/images/dsc-002.arw",
					},
				},
			},
		},
		// {
		// FIXME: Not sure why this isn't working
		// 	Name: "don't report conflict if target file exists but changes BEFORE the overwriting file is renamed",
		// 	Changes: []*file.Change{
		// 		{
		// 			Source:  "dsc-001.arw",
		// 			Target:  "dsc-000.arw",
		// 			BaseDir: "testdata/images",
		// 		},
		// 		{
		// 			Source:  "dsc-002.arw",
		// 			Target:  "dsc-001.arw",
		// 			BaseDir: "testdata/images",
		// 		},
		// 	},
		// 	Want: []string{
		// 		"testdata/images/dsc-000.arw",
		// 		"testdata/images/dsc-001.arw",
		// 	},
		// 	Conflicts: make(conflict.Collection),
		// },
		{
			Name: "auto fix path exists conflict",
			Changes: []*file.Change{
				{
					Source:  "dsc-001.arw",
					Target:  "dsc-002.arw",
					BaseDir: "testdata/images",
				},
			},
			Want: []string{
				"testdata/images/dsc-002 (2).arw",
			},
			Conflicts: make(conflict.Collection),
			Args:      autoFixArgs,
		},
		{
			Name: "auto fix overwriting several files conflict",
			Changes: []*file.Change{
				{
					Source:  "1984.pdf",
					Target:  "1.pdf",
					BaseDir: "ebooks",
				},
				{
					Source:  "animal-farm.pdf",
					Target:  "1.pdf",
					BaseDir: "ebooks",
				},
				{
					Source:  "fear-of-life.pdf",
					Target:  "1.pdf",
					BaseDir: "ebooks",
				},
				{
					Source:  "lolita.pdf",
					Target:  "1.pdf",
					BaseDir: "ebooks/banned",
				},
				{
					Source:  "my-body-is-growing.pdf",
					Target:  "1.pdf",
					BaseDir: "ebooks/banned",
				},
			},
			Want: []string{
				"ebooks/1.pdf",
				"ebooks/1 (2).pdf",
				"ebooks/1 (3).pdf",
				"ebooks/banned/1.pdf",
				"ebooks/banned/1 (2).pdf",
			},
			Conflicts: make(conflict.Collection),
			Args:      autoFixArgs,
		},
	}

	validateTest(t, testCases)
}
