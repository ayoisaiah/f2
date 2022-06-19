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
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/adrg/xdg"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/pterm/pterm"
)

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
	"movies/No Pressure (2021) S1.E1.1080p.mkv",
	"movies/No Pressure (2021) S1.E2.1080p.mkv",
	"movies/No Pressure (2021) S1.E3.1080p.mkv",
	"movies/green-mile_1999.mp4",
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
		return nil, err
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
	Name     string   `json:"name"`
	Want     []Change `json:"want"`
	Args     string   `json:"args"`
	PathArgs []string `json:"path_args"`
}

type TestCase2 struct {
	Name     string   `json:"name"`
	Want     []string `json:"want"`
	Args     string   `json:"args"`
	PathArgs []string `json:"path_args"`
}

func h1(t *testing.T, filename string) []TestCase {
	t.Helper()

	var cases []TestCase

	b, err := os.ReadFile(filepath.Join("..", "testdata", filename))
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(b, &cases)
	if err != nil {
		t.Fatal(err)
	}

	return cases
}

func h2(t *testing.T, filename string) []TestCase {
	t.Helper()

	var cases []TestCase2

	b, err := os.ReadFile(filepath.Join("..", "testdata", filename))
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
			Name:     v.Name,
			Args:     v.Args,
			PathArgs: v.PathArgs,
		}

		for _, v2 := range v.Want {
			var ch Change

			sl := strings.Split(v2, "|")

			for k, v3 := range sl {
				if k == 0 {
					ch.Source = v3
				}

				if k == 1 {
					ch.Target = v3
				}

				if k == 2 {
					ch.BaseDir = v3
				}
			}

			ca.Want = append(ca.Want, ch)
		}

		c[i] = ca
	}

	return c
}

func h(t *testing.T, cases []TestCase) {
	t.Helper()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			testDir := setupNewFileSystem(t)

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

			cargs := tc.Args + " --json " + pathArgs

			args := parseArgs(t, tc.Name, cargs)

			result, err := newTestRun(args)
			if err != nil {
				t.Fatal(err)
			}

			var o jsonOutput

			err = json.Unmarshal(result, &o)
			if err != nil {
				t.Fatal(err)
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
		})
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

func TestRecursive(t *testing.T) {
	testDir := setupFileSystem(t)

	cases := []testCase{
		{
			name: "Recursively match jpg files without max depth specified",
			want: []Change{
				{
					Source:  "a.jpg",
					BaseDir: filepath.Join(testDir, "images"),
					Target:  "a.jpeg",
				},
				{
					Source:  "free.jpg",
					BaseDir: filepath.Join(testDir, "images", "pics"),
					Target:  "free.jpeg",
				},
				{
					Source:  "img.jpg",
					BaseDir: filepath.Join(testDir, "morepics", "nested"),
					Target:  "img.jpeg",
				},
			},
			args: "-f jpg -r jpeg -R " + testDir,
		},
		{
			name: "Recursively match jpg files with max depth set to zero",
			want: []Change{
				{
					Source:  "a.jpg",
					BaseDir: filepath.Join(testDir, "images"),
					Target:  "a.jpeg",
				},
				{
					Source:  "free.jpg",
					BaseDir: filepath.Join(testDir, "images", "pics"),
					Target:  "free.jpeg",
				},
				{
					Source:  "img.jpg",
					BaseDir: filepath.Join(testDir, "morepics", "nested"),
					Target:  "img.jpeg",
				},
			},
			args: "-f jpg -r jpeg -R -m 0 " + testDir,
		},
		{
			name: "Recursively match jpg files with max depth of 1",
			want: []Change{
				{
					Source:  "a.jpg",
					BaseDir: filepath.Join(testDir, "images"),
					Target:  "a.jpeg",
				},
			},
			args: "-f jpg -r jpeg -R -m 1 " + testDir,
		},
		{
			name: "Recursively match jpg files with max depth set to 2",
			want: []Change{
				{
					Source:  "a.jpg",
					BaseDir: filepath.Join(testDir, "images"),
					Target:  "a.jpeg",
				},
				{
					Source:  "free.jpg",
					BaseDir: filepath.Join(testDir, "images", "pics"),
					Target:  "free.jpeg",
				},
				{
					Source:  "img.jpg",
					BaseDir: filepath.Join(testDir, "morepics", "nested"),
					Target:  "img.jpeg",
				},
			},
			args: "-f jpg -r jpeg -R -m 2 " + testDir,
		},
		{
			name: "Recursively rename with multiple paths",
			want: []Change{
				{
					Source:  "ios.mp4",
					BaseDir: filepath.Join(testDir, "images", "pics"),
					Target:  "ios.mp4.bak",
				},
				{
					Source:  "linux.mp4",
					BaseDir: filepath.Join(testDir, "morepics", "nested"),
					Target:  "linux.mp4.bak",
				},
			},
			args: "-f mp4 -r mp4.bak -R -m 1 " + filepath.Join(
				testDir,
				"images",
			) + " " + filepath.Join(
				testDir,
				"morepics",
			),
		},
	}

	runFindReplaceHelper(t, cases)
}

func TestExcludeFilter(t *testing.T) {
	testDir := setupFileSystem(t)

	cases := []testCase{
		{
			name: "Exclude S1.E3 from matches",
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
			},
			args: "-f Pressure -r Limits -s -E S1.E3 " + testDir,
		},
		{
			name: "Exclude files that contain any number",
			want: []Change{
				{
					Source:  "abc.txt",
					BaseDir: filepath.Join(testDir, "conflicts"),
					Target:  "abc.md",
				},
				{
					Source:  "xyz.txt",
					BaseDir: filepath.Join(testDir, "conflicts"),
					Target:  "xyz.md",
				},
			},
			args: "-f txt -r md -R -E '\\d+' " + testDir,
		},
	}

	runFindReplaceHelper(t, cases)
}

func TestStringLiteralMode(t *testing.T) {
	testDir := setupFileSystem(t)

	cases := []testCase{
		{
			name: "String literal mode: match regex special characters without escaping them",
			want: []Change{
				{
					Source:  "100$-(boring+company).com.ng",
					BaseDir: filepath.Join(testDir, "regex"),
					Target:  "100#-[boring_company].com.ng",
				},
			},
			args: "-f $ -r # -f + -r _ -f ( -r [ -f ) -r ] -se " + filepath.Join(
				testDir,
				"regex",
			),
		},
		{
			name: "String literal mode: Basic find and replace",
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
				},
			},
			args: "-f Pressure -r Limits -s " + testDir,
		},
		{
			name: "String literal mode: replace entire string if find pattern is empty",
			want: []Change{
				{
					Source:  "No Pressure (2021) S1.E1.1080p.mkv",
					BaseDir: testDir,
					Target:  "001.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E2.1080p.mkv",
					BaseDir: testDir,
					Target:  "002.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E3.1080p.mkv",
					BaseDir: testDir,
					Target:  "003.mkv",
				},
			},
			args: "-r %03d{{ext}} -sE abc|pics " + testDir,
		},
		{
			name: "String literal mode: respect case insensitive option",
			want: []Change{
				{
					Source:  "a.jpg",
					BaseDir: filepath.Join(testDir, "images"),
					Target:  "a.jpeg",
				},
				{
					Source:  "b.jPg",
					BaseDir: filepath.Join(testDir, "images"),
					Target:  "b.jpeg",
				},
				{
					Source:  "123.JPG",
					BaseDir: filepath.Join(testDir, "images", "pics"),
					Target:  "123.jpeg",
				},
				{
					Source:  "free.jpg",
					BaseDir: filepath.Join(testDir, "images", "pics"),
					Target:  "free.jpeg",
				},
			},
			args: "-f jpg -r jpeg -siR " + filepath.Join(testDir, "images"),
		},
	}

	runFindReplaceHelper(t, cases)
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

func TestCSV(t *testing.T) {
	testDir := setupFileSystem(t)

	csv := filepath.Join("..", "testdata", "input.csv")

	cases := []testCase{
		{
			name: "Rename from CSV file",
			want: []Change{
				{
					Source:  "ios.mp4",
					BaseDir: filepath.Join(testDir, "images", "pics"),
					Target:  "a podcast on ios 15.mp4",
				},
				{
					Source:  "abc.pdf",
					BaseDir: testDir,
					Target:  "A book about africa.pdf",
				},
			},
			args: "-csv " + csv + " -r {{csv.3}}{{ext}} " + testDir,
		},
	}

	runFindReplaceHelper(t, cases)
}
