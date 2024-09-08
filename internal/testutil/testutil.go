package testutil

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/ayoisaiah/f2/app"
	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/osutil"
	cp "github.com/otiai10/copy"
)

// TestCase represents a unique test case.
type TestCase struct {
	Error            error                                                `json:"error"`
	ConflictDetected bool                                                 `json:"conflict_detected"`
	SetupFunc        func(t *testing.T, testDir string) (teardown func()) `json:"-"`
	DefaultOpts      string                                               `json:"default_opts"`
	Name             string                                               `json:"name"`
	GoldenFile       string                                               `json:"golden_file"`
	Args             []string                                             `json:"args"`
	PathArgs         []string                                             `json:"path_args"`
	Changes          file.Changes                                         `json:"changes"`
	Want             []string                                             `json:"want"`
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

	_ = cp.Copy("testdata", testDir)

	tb.Cleanup(func() {
		err = os.RemoveAll(testDir)
		if err != nil {
			tb.Log(err)
		}
	})

	// change to testDir directory
	err = os.Chdir(testDir)
	if err != nil {
		tb.Fatal(err)
	}

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
func CompareChanges(t *testing.T, want file.Changes, got file.Changes) {
	assert.Equal(t, want, got)
}

// CompareSourcePath compares the expected source paths to the actual source
// paths.
func CompareSourcePath(t *testing.T, want []string, changes file.Changes) {
	t.Helper()

	got := make([]string, len(changes))

	for i := range changes {
		got[i] = changes[i].SourcePath
	}

	assert.Equal(t, want, got)
}

// CompareTargetPath verifies that the renaming target matches expectations.
func CompareTargetPath(t *testing.T, want []string, changes file.Changes) {
	t.Helper()

	got := make([]string, len(changes))

	for i := range changes {
		got[i] = changes[i].TargetPath
	}

	assert.Equal(t, want, got)
}

// CompareGoldenFile verifies that the output of an operation matches
// the expected output.
func CompareGoldenFile(
	t *testing.T,
	tc *TestCase,
	result []byte,
	fileName ...string,
) {
	t.Helper()

	goldenFile := strings.ReplaceAll(tc.Name, " ", "_")

	if len(fileName) > 0 {
		goldenFile = fileName[0]
	}

	if runtime.GOOS == osutil.Windows {
		goldenFile = goldenFile + "_windows"
	}

	g := goldie.New(
		t,
		goldie.WithFixtureDir("testdata"),
	)

	g.Assert(t, goldenFile, result)
}

// UpdateBaseDir adds the testDir to each expected path for easy comparison.
func UpdateBaseDir(expected []string, testDir string) {
	for i := range expected {
		expected[i] = filepath.Join(testDir, expected[i])
	}
}

// GetConfig constructs the app configuration from command-line arguments.
func GetConfig(t *testing.T, tc *TestCase, testDir string) *config.Config {
	t.Helper()

	var buf bytes.Buffer

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

	f2App, err := app.Get(os.Stdin, &buf)
	if err != nil {
		t.Fatal(err)
	}

	f2App.Action = func(ctx *cli.Context) error {
		return nil
	}

	// Initialize the config
	err = f2App.Run(args)
	if err != nil {
		t.Fatal(err)
	}

	return config.Get()
}
