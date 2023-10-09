package f2_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	stdpath "path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/adrg/xdg"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/sebdah/goldie/v2"
	"golang.org/x/exp/slices"

	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/jsonutil"
	"github.com/ayoisaiah/f2/internal/osutil"
	"github.com/ayoisaiah/f2/internal/status"

	"github.com/ayoisaiah/f2"
	"github.com/ayoisaiah/f2/internal/conflict"
)

func init() {
	//nolint:dogsled // necessary for testing setup
	_, filename, _, _ := runtime.Caller(0)

	dir := stdpath.Join(stdpath.Dir(filename), "..")

	projectRoot = dir

	workingDir, err := filepath.Abs(".")
	if err != nil {
		log.Fatalf("Unable to retrieve test working directory: %v", err)
	}

	workingDir = strings.ReplaceAll(workingDir, "/", "_")
	if runtime.GOOS == osutil.Windows {
		workingDir = strings.ReplaceAll(workingDir, `\`, "_")
		workingDir = strings.ReplaceAll(workingDir, ":", "_")
	}

	backupFilePath, err = xdg.DataFile(
		filepath.Join("f2", "backups", workingDir+".json"),
	)
	if err != nil {
		log.Fatalf("Unable to retrieve xdg data file directory: %v", err)
	}
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)

var projectRoot string

var testFixtures = "testdata"

var backupFilePath string

var fileSystem = []string{
	"docs/éèêëçñåēčŭ.xlsx",
	"dev/index.js",
	"dev/index.ts",
	"docu.ments/job-contract.docx",
	"special/$-(+)_file.txt",
	"images/dsc-001.arw",
	"images/dsc-002.arw",
	"images/sony/dsc-003.arw",
	"images/canon/startrails1.jpg",
	"images/canon/startrails2.jpg",
	"movies/No Pressure (2021) S1.E1.1080p.mkv",
	"movies/No Pressure (2021) S1.E2.1080p.mkv",
	"movies/No Pressure (2021) S1.E3.1080p.mkv",
	"movies/green-mile_1999.mp4",
	"music/Overgrown (2013)/01 Overgrown.flac",
	"music/Overgrown (2013)/02 I Am Sold.flac",
	"music/Overgrown (2013)/Cover.jpg",
	"ebooks/atomic-habits.pdf",
	"ebooks/1984.pdf",
	"ebooks/animal-farm.epub",
	"ebooks/fear-of-life.EPUB",
	"ebooks/green-mile_1996.mobi",
	"ebooks/.banned/.mein-kampf.pdf",
	"ebooks/.banned/lolita.epub",
	"text/test.TXT",
	"text/test_A.txt",
	"text/test-1.txt",
	"text/test_A-1.txt",
	".golang.pdf",
}

// setupFileSystem creates all required files and folders for
// the tests and returns the absolute path to the root directory.
func setupFileSystem(tb testing.TB, testName string) string {
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

func executeTest(args []string) ([]byte, error) {
	var buf bytes.Buffer

	app := f2.GetApp(os.Stdin, &buf)

	err := app.Run(args)
	if err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

func sortChanges(s []*file.Change) {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Source < s[j].Source
	})
}

func parseArgs(t *testing.T, name, args string) []string {
	t.Helper()

	result := make([]string, len(os.Args))

	copy(result, os.Args)

	if runtime.GOOS == osutil.Windows {
		args = strings.ReplaceAll(args, `\`, `₦`)
	}

	argsSlice, err := shellquote.Split(args)
	if err != nil {
		t.Fatalf(
			"Test (%s) -> shellquote.Split(%s) yielded error: %v",
			name,
			args,
			err,
		)
	}

	if runtime.GOOS == osutil.Windows {
		for i, v := range argsSlice {
			argsSlice[i] = strings.ReplaceAll(v, `₦`, `\`)
		}
	}

	result = append(result[:1], argsSlice...)

	return result
}

type TestCase struct {
	Name        string              `json:"name"`
	Changes     []*file.Change      `json:"changes"`
	Want        []string            `json:"want"`
	Args        string              `json:"args"`
	PathArgs    []string            `json:"path_args"`
	Conflicts   conflict.Collection `json:"conflicts"`
	DefaultOpts string              `json:"default_opts"`
	GoldenFile  string              `json:"golden_file"`
	Setup       []string            `json:"setup"`
}

func retrieveTestCases(t *testing.T, filename string) []TestCase {
	t.Helper()

	var cases []TestCase

	b, err := os.ReadFile(filepath.Join(projectRoot, testFixtures, filename))
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(b, &cases)
	if err != nil {
		t.Fatal(err)
	}

	for i := range cases {
		tc := cases[i]

		for _, v := range tc.Want {
			ch := &file.Change{}
			ch.Status = status.OK
			ch.BaseDir = "." // default to current directory

			tokens := strings.Split(v, "|")

			for j, token := range tokens {
				if j == 0 {
					ch.Source = token
					continue
				}

				if j == 1 {
					ch.Target = filepath.Clean(token)
					continue
				}

				if j == 2 {
					if token != "" {
						ch.BaseDir = token
					}

					continue
				}

				r, err := strconv.ParseBool(token)

				if j == 3 {
					if err != nil {
						t.Fatal(err)
					}

					ch.IsDir = r

					continue
				}

				if j == 4 {
					if err != nil {
						t.Fatal(err)
					}

					ch.WillOverwrite = r

					continue
				}

				if j == 5 {
					ch.Status = status.Status(token)
				}
			}

			tc.Changes = append(tc.Changes, ch)
		}

		cases[i] = tc
	}

	return cases
}

// modifyTestingEnv changes some properties of the test environment
// based on the setup parameter.
func modifyTestingEnv(
	t *testing.T,
	testDir string,
	setup []string,
) (string, error) {
	t.Helper()

	if slices.Contains(setup, "testdata") {
		testDir = testFixtures
		// change test directory
		err := os.Chdir(projectRoot)
		if err != nil {
			t.Fatal(err)
		}
	}

	if slices.Contains(setup, "windows_hidden") {
		err := setHidden(filepath.Join(testDir, "images"))
		if err != nil {
			t.Fatal(err)
		}
	}

	if slices.Contains(setup, "exiftool") {
		_, err := exec.LookPath("exiftool")
		if err != nil {
			t.SkipNow()
		}
	}

	if slices.Contains(setup, "date variables") {
		mtime := time.Date(2022, time.April, 10, 13, 0, 0, 0, time.UTC)
		atime := time.Date(2023, time.July, 11, 13, 0, 0, 0, time.UTC)

		for _, file := range fileSystem {
			path := filepath.Join(testDir, file)

			err := os.Chtimes(path, atime, mtime)
			if err != nil {
				return "", err
			}
		}
	}

	return testDir, nil
}

// preTestSetup ensures that each test case is set up correctly.
func preTestSetup(
	t *testing.T,
	tc *TestCase,
) []string {
	t.Helper()

	testDir := setupFileSystem(t, cleanString(tc.Name))

	if len(tc.Setup) > 0 {
		v, err := modifyTestingEnv(t, testDir, tc.Setup)
		if err != nil {
			t.Fatal(err)
		}

		testDir = v
	}

	t.Setenv(f2.EnvDefaultOpts, tc.DefaultOpts)

	// modify the base directory
	for i := range tc.Changes {
		ch := tc.Changes[i]
		baseDir := ch.BaseDir

		ch.BaseDir = filepath.Join(testDir, baseDir)

		if slices.Contains(tc.Setup, "csv") {
			absPath, err := filepath.Abs(filepath.Join(testDir, baseDir))
			if err != nil {
				t.Fatal(err)
			}

			ch.BaseDir = absPath
		}

		tc.Changes[i] = ch
	}

	// make conflict paths relative to the test directory root
	// to match the expected output from F2
	for k, v := range tc.Conflicts {
		for j, v2 := range v {
			tc.Conflicts[k][j].Target = filepath.Join(
				testDir,
				v2.Target,
			)

			for l, v3 := range v2.Sources {
				v3 = filepath.Join(testDir, v3)
				tc.Conflicts[k][j].Sources[l] = v3
			}
		}
	}

	var pathArgs string
	if len(tc.PathArgs) == 0 {
		pathArgs = testDir
	}

	for _, v := range tc.PathArgs {
		p := fmt.Sprintf("'%s'", filepath.Join(testDir, v))
		pathArgs = strings.TrimSpace(pathArgs + " " + p)
	}

	var args string

	//nolint:gocritic // if-else more appropriate
	if tc.GoldenFile != "" {
		if runtime.GOOS == osutil.Windows {
			t.SkipNow()
		}

		args = tc.Args + " --no-color " + pathArgs
	} else if strings.Contains(tc.Args, "-") {
		args = tc.Args + " --json " + pathArgs
	} else {
		args = tc.Args + " " + pathArgs
	}

	argsSlice := parseArgs(t, tc.Name, args)

	return argsSlice
}

func runTestCases(t *testing.T, cases []TestCase) {
	t.Helper()

	for i := range cases {
		tc := cases[i]

		t.Run(tc.Name, func(t *testing.T) {
			argsSlice := preTestSetup(t, &tc)

			result, err := executeTest(argsSlice)
			if err != nil {
				if len(tc.Conflicts) == 0 &&
					tc.GoldenFile == "" {
					t.Log(string(result))
					t.Fatal(err)
				}
			}

			if tc.GoldenFile != "" {
				g := goldie.New(
					t,
					goldie.WithFixtureDir(testFixtures),
				)

				g.Assert(t, tc.GoldenFile, result)
			} else {
				assertJSON(t, &tc, result)
			}
		})
	}
}

func prettyPrint(i interface{}) string {
	//nolint:errchkjson // no need to check error
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func cleanString(str string) string {
	return nonAlphanumericRegex.ReplaceAllString(str, "_")
}

func assertJSON(t *testing.T, tc *TestCase, result []byte) {
	t.Helper()

	var output jsonutil.Output

	err := json.Unmarshal(result, &output)
	if err != nil {
		t.Fatal(err)
	}

	if len(tc.Conflicts) > 0 {
		if !cmp.Equal(
			tc.Conflicts,
			output.Conflicts,
		) {
			t.Fatalf(
				"Test (%s) — Expected: %+v, got: %+v\n",
				tc.Name,
				tc.Conflicts,
				output.Conflicts,
			)
		}

		return
	}

	sortChanges(tc.Changes)
	sortChanges(output.Changes)

	if !cmp.Equal(
		tc.Changes,
		output.Changes,
		cmpopts.IgnoreUnexported(file.Change{}),
	) &&
		len(tc.Changes) != 0 {
		t.Fatalf(
			"Test (%s) -> Expected results to be: %s, but got: %s\n",
			tc.Name,
			prettyPrint(tc.Changes),
			prettyPrint(output.Changes),
		)
	}
}

func TestAllOSes(t *testing.T) {
	cases := retrieveTestCases(t, "all.json")
	runTestCases(t, cases)
}

func TestShortHelp(t *testing.T) {
	help := f2.ShortHelp(f2.NewApp())

	if runtime.GOOS == osutil.Windows {
		// TODO: due to line endings on Windows
		// FIXME: Needs to be corrected instead of ignored
		t.SkipNow()
	}

	g := goldie.New(
		t,
		goldie.WithFixtureDir(filepath.Join(projectRoot, testFixtures)),
	)
	g.Assert(t, "help", []byte(help))
}
