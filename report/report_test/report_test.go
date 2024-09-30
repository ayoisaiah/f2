package report_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/ayoisaiah/f2/internal/apperr"
	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/status"
	"github.com/ayoisaiah/f2/internal/testutil"
	"github.com/ayoisaiah/f2/report"
)

var filesWithConflicts = file.Changes{
	{
		Source: "original.txt",
		Target: "",
		Status: status.EmptyFilename,
	},
	{
		Source: "original.txt",
		Target: "new_file.",
		Status: status.TrailingPeriod,
	},
	{
		Source: "file1.txt",
		Target: "existing_file.txt",
		Status: status.PathExists,
	},
	{
		Source: "original.txt",
		Target: "new:file.txt",
		Status: status.ForbiddenCharacters,
	},
	{
		Source: "file2.txt",
		Target: "new_file.txt",
		Status: status.OverwritingNewPath,
	},
	{
		Source: "original.txt",
		Target: "this_is_a_very_long_filename_that_exceeds_the_maximum_allowed_length.txt",
		Status: status.FilenameLengthExceeded,
	},
	{
		Source: "1.txt",
		Target: "2.txt",
		Status: status.TargetFileChanging,
	},
	{
		Source: "nonexistent_file.txt",
		Target: "new_name.txt",
		Status: status.SourceNotFound,
	},
}

var filesNoConflicts = file.Changes{
	{
		Source: "macos_update_notes_2023.txt",
		Target: "macos_update_notes_2023.txt",
		Status: status.Unchanged,
	},
	{
		Source: "file with spaces.txt",
		Target: "file_with_underscores.txt",
		Status: status.OK,
	},
	{
		Source:        "file1.txt",
		Target:        "existing_file.txt",
		Status:        status.Overwriting,
		WillOverwrite: true,
	},
	{
		Source: "nonexistent_file.txt",
		Target: "file_with_underscores.txt",
		Status: status.Ignored,
	},
}

func reportTest(t *testing.T, cases []testutil.TestCase) {
	t.Helper()

	for i := range cases {
		tc := cases[i]

		testutil.UpdateFileChanges(tc.Changes)

		t.Run(tc.Name, func(t *testing.T) {
			if tc.SetupFunc != nil {
				t.Cleanup(tc.SetupFunc(t, ""))
			}

			conf := testutil.GetConfig(t, &tc, ".")

			var stdout bytes.Buffer
			var stderr bytes.Buffer

			config.Stdout = &stdout
			config.Stderr = &stderr

			switch strings.Split(t.Name(), "/")[0] {
			case "TestReport":
				report.Report(conf, tc.Changes, tc.ConflictDetected)
			case "TestPrintResults":
				report.PrintResults(conf, tc.Changes, tc.Error)
			case "TestNoMatches":
				report.NoMatches(conf)
			}

			tc.SnapShot.Stdout = stdout.Bytes()
			tc.SnapShot.Stderr = stderr.Bytes()

			testutil.CompareGoldenFile(t, &tc)
		})
	}
}

func TestPrintResults(t *testing.T) {
	testCases := []testutil.TestCase{
		{
			Name: "print results with errors",
			Changes: file.Changes{
				{
					Source: "a.txt",
					Target: "b.txt",
					Status: status.OK,
					Error: errors.New(
						"rename a.txt b.txt: operation not permitted",
					),
				},
			},
			Error: &apperr.Error{
				Context: []int{0},
			},
			Args: []string{"-r"},
		},
		{
			Name: "print results without errors",
			Changes: file.Changes{
				{
					Source: "a.txt",
					Target: "b.txt",
					Status: status.OK,
				},
			},
			Args: []string{"-f", "-r"},
		},
		{
			Name: "print results without errors (piped output)",
			Changes: file.Changes{
				{
					Source: "a.txt",
					Target: "b.txt",
					Status: status.OK,
				},
			},
			Args:       []string{"-f", "-r"},
			PipeOutput: true,
		},
		{
			Name: "print results without errors (verbose)",
			Changes: file.Changes{
				{
					Source: "a.txt",
					Target: "b.txt",
					Status: status.OK,
				},
			},
			Args: []string{"-f", "-r", "-V"},
		},
	}

	reportTest(t, testCases)
}

func TestReport(t *testing.T) {
	testCases := []testutil.TestCase{
		{
			Name:             "report file conflicts",
			Changes:          filesWithConflicts,
			ConflictDetected: true,
			Args:             []string{"-r"},
		},
		{
			Name:             "report file conflicts with --no-color",
			Changes:          filesWithConflicts,
			StdoutGoldenFile: "report_file_conflicts_no_color_stdout",
			StderrGoldenFile: "report_file_conflicts_no_color_stderr",
			ConflictDetected: true,
			Args:             []string{"-f", "-r", "--no-color"},
		},
		{
			Name:             "report file conflicts with NO_COLOR env",
			StdoutGoldenFile: "report_file_conflicts_no_color_stdout",
			StderrGoldenFile: "report_file_conflicts_no_color_stderr",
			Changes:          filesWithConflicts,
			ConflictDetected: true,
			Args:             []string{"-r"},
			SetEnv: map[string]string{
				"NO_COLOR": "",
			},
		},
		{
			Name:    "report file status",
			Changes: filesNoConflicts,
			Args:    []string{"-r"},
		},
		{
			Name:    "report file status with F2_NO_COLOR env",
			Changes: filesNoConflicts,
			Args:    []string{"-r"},
			SetEnv: map[string]string{
				"F2_NO_COLOR": "",
			},
		},
		{
			Name:             "report file conflicts in JSON",
			Changes:          filesWithConflicts,
			ConflictDetected: true,
			Args:             []string{"-f", "-r", "--json"},
		},
		{
			Name:    "report file status in JSON",
			Changes: filesNoConflicts,
			Args:    []string{"-f", "-r", "--json"},
		},
	}

	reportTest(t, testCases)
}

func TestNoMatches(t *testing.T) {
	testCases := []testutil.TestCase{
		{
			Name: "report no matches (standard)",
			Args: []string{"-f", "-r"},
		},
		{
			Name: "report no matches (csv)",
			Args: []string{"--csv", "input.csv"},
		},
		{
			Name: "report no matches (backup)",
			Args: []string{"-u"},
		},
	}

	reportTest(t, testCases)
}

func TestExitWithErr(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		report.ExitWithErr(errors.New("something went wrong"))
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestExitWithErr")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")

	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestBackupFailed(t *testing.T) {
	tc := testutil.TestCase{
		Name: "report backup failure",
	}

	var stderr bytes.Buffer

	config.Stderr = &stderr

	report.BackupFailed(errors.New("unable to write file"))

	tc.SnapShot.Stderr = stderr.Bytes()

	testutil.CompareGoldenFile(t, &tc)
}

func TestBackupRemovalFailed(t *testing.T) {
	tc := testutil.TestCase{
		Name: "report backup file removal failure",
	}

	var stderr bytes.Buffer

	config.Stderr = &stderr

	report.BackupFileRemovalFailed(errors.New("file not found"))

	tc.SnapShot.Stderr = stderr.Bytes()

	testutil.CompareGoldenFile(t, &tc)
}

func TestNonExistentFile(t *testing.T) {
	tc := testutil.TestCase{
		Name: "report non existent file",
	}

	var stderr bytes.Buffer

	config.Stderr = &stderr

	report.NonExistentFile("test_file.txt", 0)

	tc.SnapShot.Stderr = stderr.Bytes()

	testutil.CompareGoldenFile(t, &tc)
}
