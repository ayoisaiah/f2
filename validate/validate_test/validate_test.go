package validate_test

import (
	"path/filepath"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/copier"

	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/status"
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

			if ch.Status == "" {
				cases[i].Changes[j].Status = status.OK
			}

			cases[i].Changes[j].OriginalName = ch.Source
			cases[i].Changes[j].SourcePath = filepath.Join(
				ch.BaseDir,
				ch.Source,
			)
			cases[i].Changes[j].TargetPath = filepath.Join(
				ch.BaseDir,
				ch.Target,
			)
		}
	}

	for i := range cases {
		tc := cases[i]

		t.Run(tc.Name, func(t *testing.T) {
			if tc.SetupFunc != nil {
				t.Cleanup(tc.SetupFunc(t, ""))
			}

			if len(tc.Args) == 0 {
				tc.Args = []string{"-r", ""}
			}

			config := testutil.GetConfig(t, &tc, ".")

			var expectedChanges file.Changes

			err := copier.Copy(&expectedChanges, &tc.Changes)
			if err != nil {
				t.Fatal(err)
			}

			for j := range tc.Changes {
				tc.Changes[j].Status = status.OK
			}

			conflictDetected := validate.Validate(
				tc.Changes,
				config.AutoFixConflicts,
				config.AllowOverwrites,
			)

			if tc.ConflictDetected && !conflictDetected {
				spew.Dump(tc.Changes, conflictDetected)
				t.Fatal("expected a conflict, but got none")
			}

			if tc.ConflictDetected {
				testutil.CompareChanges(t, expectedChanges, tc.Changes)
			} else {
				testutil.CompareTargetPath(t, tc.Want, tc.Changes)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	testCases := []testutil.TestCase{
		{
			Name: "detect empty filename conflict",
			Changes: file.Changes{
				{
					Source:  "1984.pdf",
					Target:  "",
					BaseDir: "ebooks",
					Status:  status.EmptyFilename,
				},
			},
			ConflictDetected: true,
		},
		{
			Name: "detect overwriting newly renamed path conflict",
			Changes: file.Changes{
				{
					Source:  "index.js",
					Target:  "index.svelte",
					BaseDir: "dev",
				},
				{
					Source:  "index.ts",
					Target:  "index.svelte",
					Status:  status.OverwritingNewPath,
					BaseDir: "dev",
				},
			},
			ConflictDetected: true,
		},
		{
			Name: "report conflict when target path exists but changes AFTER the overwriting file is renamed",
			Changes: file.Changes{
				{
					Source:  "dsc-001.arw",
					Target:  "dsc-002.arw",
					Status:  status.PathExists,
					BaseDir: "testdata/images",
				},
				{
					Source:  "dsc-002.arw",
					Target:  "dsc-003.arw",
					Status:  status.TargetFileChanging,
					BaseDir: "testdata/images",
				},
			},
			ConflictDetected: true,
		},
		{
			Name: "don't report conflict if target file exists but changes BEFORE the overwriting file is renamed",
			Changes: file.Changes{
				{
					Source:  "dsc-001.arw",
					Target:  "dsc-000.arw",
					BaseDir: "testdata/images",
				},
				{
					Source:  "dsc-002.arw",
					Target:  "dsc-001.arw",
					BaseDir: "testdata/images",
				},
			},
			Want: []string{
				"testdata/images/dsc-000.arw",
				"testdata/images/dsc-001.arw",
			},
		},
		{
			Name: "auto fix path exists conflict",
			Changes: file.Changes{
				{
					Source:  "dsc-001.arw",
					Target:  "dsc-002.arw",
					BaseDir: "testdata/images",
				},
			},
			Want: []string{
				"testdata/images/dsc-002(1).arw",
			},
			Args: autoFixArgs,
		},
		{
			Name: "auto fix overwriting several files conflict",
			Changes: file.Changes{
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
				"ebooks/1(1).pdf",
				"ebooks/1(2).pdf",
				"ebooks/banned/1.pdf",
				"ebooks/banned/1(1).pdf",
			},
			Args: autoFixArgs,
		},
		{
			Name: "auto fix overwriting files conflict with custom pattern",
			Changes: file.Changes{
				{
					Source:  "myFile.pdf",
					Target:  "myFile.pdf",
					BaseDir: "ebooks",
				},
				{
					Source:  "myOwnFile.pdf",
					Target:  "myFile.pdf",
					BaseDir: "ebooks",
				},
				{
					Source:  "aFile",
					Target:  "hisFile.pdf",
					BaseDir: "ebooks",
				},
				{
					Source:  "theirFile.pdf",
					Target:  "hisFile.pdf",
					BaseDir: "ebooks",
				},
				{
					Source:  "myFile_01.pdf",
					Target:  "myFile_01.pdf",
					BaseDir: "ebooks",
				},
			},
			Want: []string{
				"ebooks/myFile.pdf",
				"ebooks/myFile_02.pdf",
				"ebooks/hisFile.pdf",
				"ebooks/hisFile_01.pdf",
				"ebooks/myFile_01.pdf",
			},
			Args: append(autoFixArgs, "--fix-conflicts-pattern", "_%02d"),
		},
		{
			Name: "detect if target file is changing later",
			Changes: file.Changes{
				{
					Source: "03.txt",
					Target: "02.txt",
				},
				{
					Source: "02.txt",
					Target: "01.txt",
					Status: status.TargetFileChanging,
				},
				{
					Source: "01.txt",
					Target: "00.txt",
					Status: status.TargetFileChanging,
				},
			},
			ConflictDetected: true,
		},
		{
			Name: "auto fix target file changing later",
			Changes: file.Changes{
				{
					Source: "03.txt",
					Target: "02.txt",
				},
				{
					Source: "02.txt",
					Target: "01.txt",
				},
				{
					Source: "01.txt",
					Target: "00.txt",
				},
			},
			Want: []string{"00.txt", "01.txt", "02.txt"},
			Args: autoFixArgs,
		},
	}

	validateTest(t, testCases)
}
