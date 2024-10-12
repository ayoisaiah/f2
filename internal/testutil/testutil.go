package testutil

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/pterm/pterm"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/ayoisaiah/f2/app"
	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/osutil"
	"github.com/ayoisaiah/f2/internal/status"
)

// TestCase represents a unique test case.
type TestCase struct {
	Error            error                                                `json:"error"`
	SetEnv           map[string]string                                    `json:"env"`
	SetupFunc        func(t *testing.T, testDir string) (teardown func()) `json:"-"`
	StdoutGoldenFile string                                               `json:"stdout_golden_file"`
	DefaultOpts      string                                               `json:"default_opts"`
	Name             string                                               `json:"name"`
	StderrGoldenFile string                                               `json:"stderr_golden_file"`
	SnapShot         struct {
		Stdout []byte
		Stderr []byte
	} `json:"-"`
	Args             []string     `json:"args"`
	PathArgs         []string     `json:"path_args"`
	Changes          file.Changes `json:"changes"`
	Want             []string     `json:"want"`
	ConflictDetected bool         `json:"conflict_detected"`
	PipeOutput       bool         `json:"pipe_output"`
}

// SetupFileSystem creates all required files and folders for
// the tests and returns the absolute path to the root directory.
func SetupFileSystem(
	tb testing.TB,
	testName string,
	fileSystem []string,
) string {
	tb.Helper()

	testDir, err := os.MkdirTemp(os.TempDir(), testName)
	if err != nil {
		tb.Fatal(err)
	}

	tb.Cleanup(func() {
		err = os.RemoveAll(testDir)
		if err != nil {
			tb.Log(err)
		}
	})

	for _, v := range fileSystem {
		dir := filepath.Dir(v)

		filePath := filepath.Join(testDir, dir)

		err := os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			tb.Fatalf(
				"Unable to create directories in path: '%s', due to err: %v",
				filePath,
				err,
			)
		}
	}

	for _, f := range fileSystem {
		pathToFile := filepath.Join(testDir, f)

		testFile, err := os.Create(pathToFile)
		if err != nil {
			tb.Fatalf(
				"Unable to write to file: '%s', due to err: %v",
				pathToFile,
				err,
			)
		}

		testFile.Close()
	}

	return testDir
}

// CompareChanges compares the expected file changes to the ones received.
func CompareChanges(t *testing.T, want, got file.Changes) {
	t.Helper()

	assert.Equal(t, want, got)
}

// CompareSourcePath compares the expected source paths to the actual source
// paths.
func CompareSourcePath(t *testing.T, want []string, changes file.Changes) {
	t.Helper()

	got := make([]string, len(changes))

	for i := range changes {
		got[i] = filepath.FromSlash(changes[i].SourcePath)
	}

	for i := range want {
		want[i] = filepath.FromSlash(want[i])
	}

	assert.Equal(t, want, got)
}

// CompareTargetPath verifies that the renaming target matches expectations.
func CompareTargetPath(t *testing.T, want []string, changes file.Changes) {
	t.Helper()

	got := make([]string, len(changes))

	for i := range changes {
		got[i] = filepath.FromSlash(changes[i].TargetPath)
	}

	for i := range want {
		want[i] = filepath.FromSlash(want[i])
	}

	assert.Equal(t, want, got)
}

// CompareGoldenFile verifies that the output of an operation matches
// the expected output.
func CompareGoldenFile(t *testing.T, tc *TestCase) {
	t.Helper()

	if runtime.GOOS == osutil.Windows {
		// TODO: need to sort out line endings
		t.Skip("skipping golden file test in Windows")
	}

	g := goldie.New(
		t,
		goldie.WithFixtureDir("testdata"),
	)

	compareOutput := func(output []byte, fileSuffix, goldenFileName string) {
		if goldenFileName == "" {
			goldenFileName = strings.ReplaceAll(tc.Name, " ", "_") + fileSuffix
		}

		if output != nil {
			g.Assert(t, goldenFileName, output)
		} else {
			f := filepath.Join("testdata", goldenFileName+".golden")
			if _, err := os.Stat(f); err == nil || errors.Is(err, os.ErrExist) {
				t.Fatalf("expected no output, but golden file exists: %s", f)
			}
		}
	}

	compareOutput(tc.SnapShot.Stdout, "_stdout", tc.StdoutGoldenFile)
	compareOutput(tc.SnapShot.Stderr, "_stderr", tc.StderrGoldenFile)
}

// UpdateBaseDir adds the testDir to each expected path for easy comparison.
func UpdateBaseDir(expected []string, testDir string) {
	for i := range expected {
		expected[i] = filepath.Join(testDir, expected[i])
	}
}

func UpdateFileChanges(files file.Changes) {
	for i := range files {
		ch := files[i]

		files[i].OriginalName = ch.Source
		files[i].Position = i
		files[i].SourcePath = filepath.Join(
			ch.BaseDir,
			ch.Source,
		)
		files[i].TargetPath = filepath.Join(
			ch.BaseDir,
			ch.Target,
		)
	}
}

func RunTestCase(
	t *testing.T,
	tc *TestCase,
	runFunc func(t *testing.T, tc *TestCase),
) {
	t.Helper()

	t.Run(tc.Name, func(t *testing.T) {
		if tc.SetupFunc != nil {
			t.Cleanup(tc.SetupFunc(t, ""))
		}

		runFunc(t, tc)
	})
}

func ProcessTestCaseChanges(t *testing.T, cases []TestCase) {
	t.Helper()

	for i := range cases {
		tc := cases[i]
		for j := range tc.Changes {
			ch := tc.Changes[j]

			if ch.Status == "" {
				cases[i].Changes[j].Status = status.OK
			}

			cases[i].Changes[j].OriginalName = ch.Source

			if cases[i].Changes[j].TargetPath == "" {
				cases[i].Changes[j].SourcePath = filepath.Join(
					ch.BaseDir,
					ch.Source,
				)
			}

			if cases[i].Changes[j].TargetPath == "" {
				cases[i].Changes[j].TargetPath = filepath.Join(
					ch.BaseDir,
					ch.Target,
				)
			}
		}
	}
}

// GetConfig constructs the app configuration from command-line arguments.
func GetConfig(t *testing.T, tc *TestCase, testDir string) *config.Config {
	t.Helper()

	for k, v := range tc.SetEnv {
		t.Setenv(k, v)
	}

	// add fake binary name as first argument
	args := append([]string{"f2_test"}, tc.Args...)

	if len(tc.PathArgs) > 0 {
		for i, v := range tc.PathArgs {
			tc.PathArgs[i] = filepath.Join(testDir, v)
		}
	} else {
		tc.PathArgs = []string{testDir}
	}

	// add test directory as last argument
	args = append(args, tc.PathArgs...)

	f2App, err := app.Get(os.Stdin, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	f2App.Action = func(ctx *cli.Context) error {
		// Reset pterm to default state
		pterm.EnableStyling()
		// Re-initialize config with pipe output value set per test
		_, _ = config.Init(ctx, tc.PipeOutput)

		return nil
	}

	// Initialize the config
	err = f2App.Run(args)
	if err != nil {
		t.Fatal(err)
	}

	return config.Get()
}
