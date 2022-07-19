package f2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
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
	"github.com/pterm/pterm"
	"github.com/sebdah/goldie/v2"
)

var fixtures = filepath.Join("..", "testdata")

type testCase struct {
	name           string
	want           []Change
	args           string
	undoArgs       []string
	defaultOpts    string
	expectedErrors []int
}

var (
	backupFilePath string
)

var newFileSystem = []string{
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
	".golang.pdf",
}

var fileSystem = []string{
	"No Pressure (2021) S1.E1.1080p.mkv",
	"No Pressure (2021) S1.E2.1080p.mkv",
	"No Pressure (2021) S1.E3.1080p.mkv",
	"docs.03.05.period/word.docx",
	"images/a.jpg",
	"images/b.jPg",
	"images/abc.png",
	"images/456.webp",
	"images/pics/123.JPG",
	"images/pics/free.jpg",
	"images/pics/ios.mp4",
	"morepics/pic-1.avif",
	"morepics/pic-2.avif",
	"morepics/nested/img.jpg",
	"morepics/nested/linux.mp4",
	"scripts/index.js",
	"scripts/main.js",
	"abc.pdf",
	"abc.epub",
	".forbidden.pdf",
	".dir/sample.pdf",
	"conflicts/abc.txt",
	"conflicts/xyz.txt",
	"conflicts/123.txt",
	"conflicts/123 (3).txt",
	"regex/100$-(boring+company).com.ng",
	"weirdo/Data Structures and Algorithms/1. Asymptotic Analysis and Insertion Sort, Merge Sort/2.Sorting & Searching why bother with these simple tasks/this is a long path/1. Sorting & Searching- why bother with these simple tasks- - Data Structure & Algorithms - Part-2.mp4",
}

func init() {
	workingDir, err := filepath.Abs(".")
	if err != nil {
		log.Fatalf("Unable to retrieve working directory: %v", err)
	}

	workingDir = strings.ReplaceAll(workingDir, "/", "_")
	if runtime.GOOS == windows {
		workingDir = strings.ReplaceAll(workingDir, `\`, "_")
		workingDir = strings.ReplaceAll(workingDir, ":", "_")
	}

	backupFilePath, err = xdg.DataFile(
		filepath.Join("f2", "backups", workingDir+".json"),
	)
	if err != nil {
		log.Fatalf("Unable to retrieve xdg data file directory: %v", err)
	}

	rand.Seed(time.Now().UnixNano())
}

// setupNewFileSystem creates all required files and folders for
// the tests and returns a function that is used as
// a teardown function when the tests are done.
func setupNewFileSystem(tb testing.TB) string {
	tb.Helper()

	testDir, err := ioutil.TempDir(".", "")
	if err != nil {
		tb.Fatalf("Unable to create temporary directory for test: %v", err)
	}

	absPath, err := filepath.Abs(testDir)
	if err != nil {
		tb.Fatalf("Unable to get absolute path to test directory: %v", err)
	}

	tb.Cleanup(func() {
		if err = os.RemoveAll(absPath); err != nil {
			tb.Fatalf(
				"Failure occurred while cleaning up the filesystem: %v",
				err,
			)
		}
	})

	for _, v := range newFileSystem {
		dir := filepath.Dir(v)

		filePath := filepath.Join(testDir, dir)

		err = os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			tb.Fatalf(
				"Unable to create directories in path: '%s', due to err: %v",
				filePath,
				err,
			)
		}
	}

	for _, f := range newFileSystem {
		pathToFile := filepath.Join(absPath, f)

		file, err := os.Create(pathToFile)
		if err != nil {
			tb.Fatalf(
				"Unable to write to file: '%s', due to err: %v",
				pathToFile,
				err,
			)
		}

		file.Close()
	}

	return absPath
}

// setupFileSystem creates all required files and folders for
// the tests and returns a function that is used as
// a teardown function when the tests are done.
func setupFileSystem(tb testing.TB) string {
	tb.Helper()

	testDir, err := ioutil.TempDir(".", "")
	if err != nil {
		tb.Fatalf("Unable to create temporary directory for test: %v", err)
	}

	absPath, err := filepath.Abs(testDir)
	if err != nil {
		tb.Fatalf("Unable to get absolute path to test directory: %v", err)
	}

	tb.Cleanup(func() {
		if err = os.RemoveAll(absPath); err != nil {
			tb.Fatalf(
				"Failure occurred while cleaning up the filesystem: %v",
				err,
			)
		}
	})

	for _, v := range fileSystem {
		dir := filepath.Dir(v)

		filePath := filepath.Join(testDir, dir)

		err = os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			tb.Fatalf(
				"Unable to create directories in path: '%s', due to err: %v",
				filePath,
				err,
			)
		}
	}

	for _, f := range fileSystem {
		pathToFile := filepath.Join(absPath, f)

		file, err := os.Create(pathToFile)
		if err != nil {
			tb.Fatalf(
				"Unable to write to file: '%s', due to err: %v",
				pathToFile,
				err,
			)
		}

		file.Close()
	}

	return absPath
}

type testResult struct {
	changes         []Change
	conflicts       map[conflictType][]Conflict
	backupFile      string
	applyError      error
	operationErrors []int
	output          *bytes.Buffer
}

func newTestRun(args []string) ([]byte, error) {
	var buf bytes.Buffer

	app := GetApp(os.Stdin, &buf)

	err := app.Run(args)
	if err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

func testRun(args []string) (testResult, error) {
	var result testResult

	var buf bytes.Buffer

	app := GetApp(os.Stdin, &buf)

	pterm.DisableOutput()

	err := app.Run(args)

	v, ok := app.Metadata["op"]
	if !ok {
		return result, fmt.Errorf("Unable to access test result: %w", err)
	}

	op, ok := v.(*Operation)
	if !ok {
		return result, fmt.Errorf("Unable to assert test operation: %w", err)
	}

	result.changes = op.matches
	result.backupFile = backupFilePath
	result.conflicts = op.conflicts
	result.operationErrors = op.errors
	result.output = &buf

	return result, err
}

func sortChanges(s []Change) {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Source < s[j].Source
	})
}

func parseArgs(t *testing.T, name, args string) []string {
	t.Helper()

	result := make([]string, len(os.Args))

	copy(result, os.Args)

	if runtime.GOOS == windows {
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

	if runtime.GOOS == windows {
		for i, v := range argsSlice {
			argsSlice[i] = strings.ReplaceAll(v, `₦`, `\`)
		}
	}

	result = append(result[:1], argsSlice...)

	return result
}

type TestCase struct {
	Name        string                      `json:"name"`
	Want        []Change                    `json:"want"`
	Args        string                      `json:"args"`
	PathArgs    []string                    `json:"path_args"`
	Conflicts   map[conflictType][]Conflict `json:"conflicts"`
	DefaultOpts string                      `json:"default_opts"`
	GoldenFile  string                      `json:"golden_file"`
}

type TestCase2 struct {
	Name        string                      `json:"name"`
	Want        []string                    `json:"want"`
	Args        string                      `json:"args"`
	PathArgs    []string                    `json:"path_args"`
	Conflicts   map[conflictType][]Conflict `json:"conflicts"`
	DefaultOpts string                      `json:"default_opts"`
	GoldenFile  string                      `json:"golden_file"`
}

func h2(t *testing.T, filename string) []TestCase {
	t.Helper()

	var cases []TestCase2

	b, err := os.ReadFile(filepath.Join(fixtures, filename))
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(b, &cases)
	if err != nil {
		t.Fatal(err)
	}

	c := make([]TestCase, len(cases))

	for i, v := range cases {
		ca := TestCase{
			Name:        v.Name,
			Args:        v.Args,
			PathArgs:    v.PathArgs,
			Conflicts:   v.Conflicts,
			DefaultOpts: v.DefaultOpts,
			GoldenFile:  v.GoldenFile,
		}

		for _, v2 := range v.Want {
			var ch Change

			sl := strings.Split(v2, "|")

			for k, v3 := range sl {
				if k == 0 {
					ch.Source = v3
					continue
				}

				if k == 1 {
					ch.Target = v3
					continue
				}

				if k == 2 {
					if v3 != "" {
						ch.BaseDir = v3
					}

					continue
				}

				r, err := strconv.ParseBool(v3)
				if err != nil {
					t.Fatal(err)
				}

				if k == 3 {
					ch.IsDir = r
					continue
				}

				if k == 4 {
					ch.WillOverwrite = r
					continue
				}
			}

			ca.Want = append(ca.Want, ch)
		}

		c[i] = ca
	}

	return c
}

func preTestSetup(testDir, name string) error {
	if strings.Contains(name, "date variables") {
		mtime := time.Date(2022, time.April, 10, 13, 0, 0, 0, time.UTC)
		atime := time.Date(2023, time.July, 11, 13, 0, 0, 0, time.UTC)

		for _, file := range newFileSystem {
			path := filepath.Join(testDir, file)

			err := os.Chtimes(path, atime, mtime)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func h(t *testing.T, cases []TestCase) {
	t.Helper()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			testDir := setupNewFileSystem(t)

			var prefix string

			colon := strings.Index(tc.Name, ":")
			if colon != -1 {
				prefix = tc.Name[:colon]

				switch prefix {
				case "testdata", "golden":
					testDir = filepath.Join(fixtures)

					if strings.Contains(tc.Name, "exiftool") {
						_, err := exec.LookPath("exiftool")
						if err != nil {
							return
						}
					}
				case "windows":
					if runtime.GOOS != windows {
						return
					}
				case "macos":
					if runtime.GOOS != darwin {
						return
					}
				case "unix":
					if runtime.GOOS == windows {
						return
					}
				case "setup":
					preTestSetup(testDir, tc.Name)
				}
			}

			if tc.DefaultOpts != "" {
				os.Setenv(envDefaultOpts, tc.DefaultOpts)
			}

			for j := range tc.Want {
				ch := tc.Want[j]
				if ch.BaseDir == "" {
					tc.Want[j].BaseDir = testDir
				} else {
					tc.Want[j].BaseDir = filepath.Join(testDir, ch.BaseDir)
				}
			}

			pathArgs := testDir
			if len(tc.PathArgs) != 0 {
				var res []string
				for _, v := range tc.PathArgs {
					res = append(
						res,
						fmt.Sprintf("'%s'", filepath.Join(testDir, v)),
					)
				}

				pathArgs = strings.Join(res, " ")
			}

			csvTestFile := filepath.Join(fixtures, "input.csv")

			if strings.Contains(tc.Args, "<csv>") {
				tc.Args = strings.ReplaceAll(tc.Args, "<csv>", csvTestFile)
			}

			var cargs string
			if prefix == "golden" {
				cargs = tc.Args + " --no-color " + pathArgs
			} else if strings.Contains(tc.Args, "-") {
				cargs = tc.Args + " --json " + pathArgs
			} else {
				cargs = tc.Args + " " + pathArgs
			}

			args := parseArgs(t, tc.Name, cargs)

			result, err := newTestRun(args)
			if err != nil {
				if len(tc.Conflicts) == 0 && prefix != "golden" {
					t.Log(string(result))
					t.Fatal(err)
				}
			}

			if tc.DefaultOpts != "" {
				os.Setenv(envDefaultOpts, "")
			}

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

			if prefix == "golden" {
				g := goldie.New(
					t,
					goldie.WithFixtureDir(filepath.Join(fixtures)),
				)

				g.Assert(t, tc.GoldenFile, result)
			} else {
				jsonTest(t, &tc, result)
			}
		})
	}
}

func jsonTest(t *testing.T, tc *TestCase, result []byte) {
	t.Helper()

	var o jsonOutput

	err := json.Unmarshal(result, &o)
	if err != nil {
		t.Fatal(err)
	}

	if len(tc.Conflicts) > 0 {
		if !cmp.Equal(
			tc.Conflicts,
			o.Conflicts,
		) {
			t.Fatalf(
				"Test (%s) — Expected: %+v, got: %+v\n",
				tc.Name,
				tc.Conflicts,
				o.Conflicts,
			)
		}

		return
	}

	sortChanges(tc.Want)
	sortChanges(o.Changes)

	if !cmp.Equal(
		tc.Want,
		o.Changes,
		cmpopts.IgnoreUnexported(Change{}),
	) &&
		len(tc.Want) != 0 {
		t.Fatalf(
			"Test (%s) -> Expected results to be: %s, but got: %s\n",
			tc.Name,
			prettyPrint(tc.Want),
			prettyPrint(o.Changes),
		)
	}
}

func TestF2(t *testing.T) {
	cases := h2(t, "tests.json")
	h(t, cases)
}

func runFindReplaceHelper(t *testing.T, cases []testCase) {
	t.Helper()

	for _, tc := range cases {
		args := parseArgs(t, tc.name, tc.args)

		if tc.defaultOpts != "" {
			os.Setenv(envDefaultOpts, tc.defaultOpts)
		}

		result, err := testRun(args)
		if err != nil {
			t.Fatalf(
				"Test (%s) -> testRun(%v) yielded error: %v",
				tc.name,
				tc.args,
				err,
			)
		}

		if len(tc.expectedErrors) != len(result.operationErrors) {
			t.Fatalf(
				"Test (%s) -> Expected errors to be: %s, but got: %s",
				tc.name,
				prettyPrint(tc.expectedErrors),
				prettyPrint(result.operationErrors),
			)
		}

		if len(result.conflicts) > 0 {
			t.Fatalf(
				"Test (%s) -> Expected no conflicts but got some: %v",
				tc.name,
				prettyPrint(result.conflicts),
			)
		}

		sortChanges(tc.want)
		sortChanges(result.changes)

		if !cmp.Equal(
			tc.want,
			result.changes,
			cmpopts.IgnoreUnexported(Change{}),
		) &&
			len(tc.want) != 0 {
			t.Fatalf(
				"Test (%s) -> Expected results to be: %s, but got: %s\n",
				tc.name,
				prettyPrint(tc.want),
				prettyPrint(result.changes),
			)
		}

		if tc.defaultOpts != "" {
			os.Setenv(envDefaultOpts, "")
		}
	}
}

func TestApplyUndo(t *testing.T) {
	table := []testCase{
		{
			want: []Change{
				{Source: "No Pressure (2021) S1.E1.1080p.mkv", Target: "1.mkv"},
				{Source: "No Pressure (2021) S1.E2.1080p.mkv", Target: "2.mkv"},
				{Source: "No Pressure (2021) S1.E3.1080p.mkv", Target: "3.mkv"},
			},
			args:     "-f .*E(\\d+).* -r $1.mkv -x",
			undoArgs: []string{"-u", "-x"},
		},
		{
			want: []Change{
				{Source: "morepics", IsDir: true, Target: "moreimages"},
			},
			args:     "-f pic -r image -d -x",
			undoArgs: []string{"-u", "-x"},
		},
	}

	for i, v := range table {
		testDir := setupFileSystem(t)

		for i := range v.want {
			v.want[i].BaseDir = testDir
		}

		argsSlice := strings.Split(v.args, " ")
		argsSlice = append(argsSlice, testDir)

		args := os.Args[0:1]
		args = append(args, argsSlice...)
		result, _ := testRun(args) // err will be nil

		if len(result.conflicts) > 0 {
			t.Fatalf(
				"Test(%d) — Expected no conflicts but got some: %v",
				i+1,
				result.conflicts,
			)
		}

		if result.applyError != nil {
			t.Fatalf(
				"Test(%d) — Unexpected apply error: %v\n",
				i+1,
				result.applyError,
			)
		}

		// Test if the backup file was written successfully
		if result.backupFile != "" {
			file, err := os.ReadFile(result.backupFile)
			if err != nil {
				t.Fatalf(
					"Test (%s) — Unexpected error when trying to read backup file: %v\n",
					v.name,
					err,
				)
			}

			var bf backupFile

			err = json.Unmarshal(file, &bf)
			if err != nil {
				t.Fatalf(
					"Test (%s) — Unexpected error when trying to unmarshal map file contents: %v\n",
					v.name,
					err,
				)
			}

			ch := bf.Operations

			sortChanges(ch)

			if !cmp.Equal(v.want, ch, cmpopts.IgnoreUnexported(Change{})) &&
				len(v.want) != 0 {
				t.Fatalf(
					"Test (%s) — Expected: %+v, got: %+v\n",
					v.name,
					prettyPrint(v.want),
					prettyPrint(ch),
				)
			}
		}

		// Test Undo function
		args = os.Args[0:1]
		args = append(args, v.undoArgs...)

		result, err := testRun(args)
		if err != nil {
			t.Fatalf("Test(%d) — Unexpected error in undo mode: %v\n", i+1, err)
		}

		if _, err := os.Stat(result.backupFile); err == nil ||
			errors.Is(err, os.ErrExist) {
			t.Fatalf(
				"Test (%d) - Backup file was not removed after undo operation: %v",
				i+1,
				err,
			)
		}
	}
}

func TestHandleErrors(t *testing.T) {
	testDir := setupFileSystem(t)

	cases := []testCase{
		{
			name: "Replace Pressure with Limits in string mode",
			want: []Change{
				{
					Source:  "No Pressure (2021) S1.E1.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Limits (2021) S1.E1.1080p.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E2.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Limits (2021) S1.E2.1080p.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E3.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Limits (2021) S1.E3.1080p.mkv",
					Error:   "renaming failed",
				},
			},
			expectedErrors: []int{2},
			args:           "-f Pressure -r Limits -s " + testDir,
		},
	}

	for _, tc := range cases {
		var buf bytes.Buffer

		op := &Operation{
			stdout: &buf,
		}
		op.matches = tc.want
		op.errors = tc.expectedErrors

		err := op.handleErrors()
		if err == nil {
			t.Fatalf(
				"Expected case '%s' to yield an error, but got nil",
				tc.name,
			)
		}

		str, err := op.retrieveBackupFile()
		if err != nil {
			t.Fatalf(
				"Test (%s) -> Error while retrieving backup file: %v",
				tc.name,
				err,
			)
		}

		os.Remove(str)
	}
}

func TestShortHelp(t *testing.T) {
	help := shortHelp(newApp())

	g := goldie.New(t, goldie.WithFixtureDir(fixtures))
	g.Assert(t, "help", []byte(help))
}
