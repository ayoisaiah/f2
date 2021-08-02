package f2

import (
	"encoding/json"
	"errors"
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
	"github.com/urfave/cli/v2"
)

type testCase struct {
	name           string
	want           []Change
	args           []string
	undoArgs       []string
	expectedErrors []renameError
}

var (
	backupFilePath string
)

var fileSystem = []string{
	"No Pressure (2021) S1.E1.1080p.mkv",
	"No Pressure (2021) S1.E2.1080p.mkv",
	"No Pressure (2021) S1.E3.1080p.mkv",
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
}

func init() {
	workingDir, err := filepath.Abs(".")
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}

	rand.Seed(time.Now().UnixNano())
}

// setupFileSystem creates all required files and folders for
// the tests and returns a function that is used as
// a teardown function when the tests are done.
func setupFileSystem(tb testing.TB) string {
	tb.Helper()

	testDir, err := ioutil.TempDir(".", "")
	if err != nil {
		tb.Fatal(err)
	}

	tb.Cleanup(func() {
		if err = os.RemoveAll(testDir); err != nil {
			tb.Fatal(err)
		}
	})

	directories := []string{
		"images/pics",
		"scripts",
		"morepics/nested",
		"conflicts",
		".dir",
		"regex",
	}
	for _, v := range directories {
		filePath := filepath.Join(testDir, v)

		err = os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			tb.Fatal(err)
		}
	}

	for _, f := range fileSystem {
		filePath := filepath.Join(testDir, f)
		if err = ioutil.WriteFile(filePath, []byte{}, 0600); err != nil {
			tb.Fatal(err)
		}
	}

	abs, err := filepath.Abs(testDir)
	if err != nil {
		tb.Fatal(err)
	}

	return abs
}

type ActionResult struct {
	changes         []Change
	conflicts       map[conflictType][]Conflict
	backupFile      string
	applyError      error
	operationErrors []renameError
}

func action(args []string) (ActionResult, error) {
	var result ActionResult

	app := GetApp()
	app.Action = func(c *cli.Context) error {
		op, err := newOperation(c)
		if err != nil {
			return err
		}

		op.quiet = true

		result.applyError = op.run()
		result.changes = op.matches
		result.backupFile = backupFilePath
		result.conflicts = op.conflicts
		result.operationErrors = op.errors

		return nil
	}

	return result, app.Run(args)
}

func sortChanges(s []Change) {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Source < s[j].Source
	})
}

func runFindReplace(t *testing.T, cases []testCase) {
	t.Helper()

	for _, v := range cases {
		args := os.Args[0:1]
		args = append(args, v.args...)

		result, err := action(args) // err will be nil
		if err != nil {
			t.Fatalf("Test (%s) — Unexpected error: %v", v.name, err)
		}

		if len(v.expectedErrors) != len(result.operationErrors) {
			t.Fatalf(
				"Test (%s) — Expected %d errors but got %d errors: %v",
				v.name,
				len(v.expectedErrors),
				len(result.operationErrors),
				result.operationErrors,
			)
		}

		if len(result.conflicts) > 0 {
			t.Fatalf(
				"Test (%s) — Expected no conflicts but got some: %v",
				v.name,
				result.conflicts,
			)
		}

		sortChanges(v.want)
		sortChanges(result.changes)

		if !cmp.Equal(
			v.want,
			result.changes,
			cmpopts.IgnoreUnexported(Change{}),
		) &&
			len(v.want) != 0 {
			t.Fatalf(
				"Test (%s) — Expected: %+v, got: %+v\n",
				v.name,
				prettyPrint(v.want),
				prettyPrint(result.changes),
			)
		}
	}
}

func TestFilePaths(t *testing.T) {
	testDir := setupFileSystem(t)
	cases := []testCase{
		{
			name: "Target a specific mkv file",
			want: []Change{
				{
					Source:  "No Pressure (2021) S1.E3.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure (2021) S1.E3.1080p.mp4",
				},
			},
			args: []string{
				"-f",
				"mkv",
				"-r",
				"mp4",
				filepath.Join(testDir, "No Pressure (2021) S1.E3.1080p.mkv"),
			},
		},
		{
			name: "Combine file paths and directory paths",
			want: []Change{
				{
					Source:  "abc.pdf",
					BaseDir: testDir,
					Target:  "qqq.pdf",
				},
				{
					Source:  "abc.epub",
					BaseDir: testDir,
					Target:  "qqq.epub",
				},
				{
					Source:  "abc.png",
					BaseDir: filepath.Join(testDir, "images"),
					Target:  "qqq.png",
				},
			},
			args: []string{
				"-f",
				"abc",
				"-r",
				"qqq",
				testDir,
				filepath.Join(testDir, "images", "abc.png"),
			},
		},
		{
			name: "No side effects should result from specifying a directory and a file inside the directory",
			want: []Change{
				{
					Source:  "abc.png",
					BaseDir: filepath.Join(testDir, "images"),
					Target:  "qqq.png",
				},
			},
			args: []string{
				"-f",
				"abc",
				"-r",
				"qqq",
				filepath.Join(testDir, "images"),
				filepath.Join(testDir, "images", "abc.png"),
			},
		},
		{
			name: "Specifying a file path should be unaffected by recursion",
			want: []Change{
				{
					Source:  "abc.pdf",
					BaseDir: testDir,
					Target:  "qqq.pdf",
				},
			},
			args: []string{
				"-f",
				"abc",
				"-r",
				"qqq",
				"-R",
				filepath.Join(testDir, "abc.pdf"),
			},
		},
	}

	runFindReplace(t, cases)
}

func TestHidden(t *testing.T) {
	testDir := setupFileSystem(t)
	cases := []testCase{
		{
			name: "Hidden files are ignored by default",
			want: []Change{
				{
					Source:  "abc.pdf",
					BaseDir: testDir,
					Target:  "abc.pdf.bak",
				},
			},
			args: []string{"-f", "pdf", "-r", "pdf.bak", "-R", testDir},
		},
		{
			name: "Hidden files are included",
			want: []Change{
				{
					Source:  "abc.pdf",
					BaseDir: testDir,
					Target:  "abc.pdf.bak",
				},
				{
					Source:  "sample.pdf",
					BaseDir: filepath.Join(testDir, ".dir"),
					Target:  "sample.pdf.bak",
				},
				{
					Source:  ".forbidden.pdf",
					BaseDir: testDir,
					Target:  ".forbidden.pdf.bak",
				},
			},
			args: []string{"-f", "pdf", "-r", "pdf.bak", "-H", "-R", testDir},
		},
	}

	runFindReplace(t, cases)
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
			args: []string{"-f", "jpg", "-r", "jpeg", "-R", testDir},
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
			args: []string{"-f", "jpg", "-r", "jpeg", "-R", "-m", "0", testDir},
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
			args: []string{"-f", "jpg", "-r", "jpeg", "-R", "-m", "1", testDir},
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
			args: []string{"-f", "jpg", "-r", "jpeg", "-R", "-m", "2", testDir},
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
			args: []string{
				"-f",
				"mp4",
				"-r",
				"mp4.bak",
				"-R",
				"-m",
				"1",
				filepath.Join(testDir, "images"),
				filepath.Join(testDir, "morepics"),
			},
		},
	}

	runFindReplace(t, cases)
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
			args: []string{
				"-f",
				"Pressure",
				"-r",
				"Limits",
				"-s",
				"-E",
				"S1.E3",
				testDir,
			},
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
			args: []string{
				"-f",
				"txt",
				"-r",
				"md",
				"-R",
				"-E",
				"\\d+",
				testDir,
			},
		},
	}

	runFindReplace(t, cases)
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
			args: []string{
				"-f",
				"$",
				"-r",
				"#",
				"-f",
				"+",
				"-r",
				"_",
				"-f",
				"(",
				"-r",
				"[",
				"-f",
				")",
				"-r",
				"]",
				"-se",
				filepath.Join(testDir, "regex"),
			},
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
			args: []string{"-f", "Pressure", "-r", "Limits", "-s", testDir},
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
			args: []string{
				"-r",
				"%03d{{ext}}",
				"-sE",
				"abc|pics",
				testDir,
			},
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
			args: []string{
				"-f",
				"jpg",
				"-r",
				"jpeg",
				"-siR",
				filepath.Join(testDir, "images"),
			},
		},
	}

	runFindReplace(t, cases)
}

func TestApplyUndo(t *testing.T) {
	table := []testCase{
		{
			want: []Change{
				{Source: "No Pressure (2021) S1.E1.1080p.mkv", Target: "1.mkv"},
				{Source: "No Pressure (2021) S1.E2.1080p.mkv", Target: "2.mkv"},
				{Source: "No Pressure (2021) S1.E3.1080p.mkv", Target: "3.mkv"},
			},
			args: []string{
				"-f",
				".*E(\\d+).*",
				"-r",
				"$1.mkv",
				"-x",
			},
			undoArgs: []string{"-u", "-x"},
		},
		{
			want: []Change{
				{Source: "morepics", IsDir: true, Target: "moreimages"},
			},
			args: []string{
				"-f",
				"pic",
				"-r",
				"image",
				"-d",
				"-x",
			},
			undoArgs: []string{"-u", "-x"},
		},
	}

	for i, v := range table {
		testDir := setupFileSystem(t)

		for i := range v.want {
			v.want[i].BaseDir = testDir
		}

		v.args = append(v.args, testDir)

		args := os.Args[0:1]
		args = append(args, v.args...)
		result, _ := action(args) // err will be nil

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

		result, err := action(args)
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
				},
			},
			expectedErrors: []renameError{
				{
					entry: Change{
						Source:  "No Pressure (2021) S1.E3.1080p.mkv",
						BaseDir: testDir,
						Target:  "No Limits (2021) S1.E3.1080p.mkv",
					},
					err: errors.New("Missing permissions"),
				},
			},
			args: []string{"-f", "Pressure", "-r", "Limits", "-s", testDir},
		},
	}

	for _, v := range cases {
		op := &Operation{}
		op.matches = v.want
		op.errors = v.expectedErrors

		err := op.handleErrors()
		if err == nil {
			t.Fatal("Expected an error not got nil")
		}

		str, err := op.retrieveBackupFile()
		if err != nil {
			t.Fatalf("Unexpected error while retrieving backup file: %v", err)
		}

		os.Remove(str)
	}
}
