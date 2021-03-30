package f2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/urfave/cli/v2"
	"gopkg.in/djherbis/times.v1"
)

type testCase struct {
	name     string
	want     []Change
	args     []string
	undoArgs []string
}

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
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// setupFileSystem creates all required files and folders for
// the tests and returns a function that is used as
// a teardown function when the tests are done.
func setupFileSystem(t testing.TB) string {
	testDir, err := ioutil.TempDir(".", "")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err = os.RemoveAll(testDir); err != nil {
			t.Fatal(err)
		}
	})

	directories := []string{
		"images/pics",
		"scripts",
		"morepics/nested",
		"conflicts",
		".dir",
	}
	for _, v := range directories {
		filePath := filepath.Join(testDir, v)
		err = os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, f := range fileSystem {
		filePath := filepath.Join(testDir, f)
		if err = ioutil.WriteFile(filePath, []byte{}, 0600); err != nil {
			t.Fatal(err)
		}
	}

	abs, err := filepath.Abs(testDir)
	if err != nil {
		t.Fatal(err)
	}

	return abs
}

type ActionResult struct {
	changes    []Change
	conflicts  map[conflict][]Conflict
	outputFile string
	applyError error
}

func action(args []string) (ActionResult, error) {
	var result ActionResult

	app := GetApp()
	app.Action = func(c *cli.Context) error {
		op, err := NewOperation(c)
		if err != nil {
			return err
		}

		if op.undoFile != "" {
			result.outputFile = op.undoFile
			return op.Undo()
		}

		err = op.findMatches()
		if err != nil {
			return err
		}

		if len(op.excludeFilter) != 0 {
			err = op.filterMatches()
			if err != nil {
				return err
			}
		}

		if op.includeDir {
			op.SortMatches()
		}

		err = op.replace()
		if err != nil {
			return err
		}

		result.changes = op.matches

		if op.outputFile != "" {
			result.outputFile = op.outputFile
			err = op.WriteToFile()
			if err != nil {
				return err
			}
		}

		result.applyError = op.apply()
		result.conflicts = op.conflicts

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
	for _, v := range cases {
		args := os.Args[0:1]
		args = append(args, v.args...)
		result, _ := action(args) // err will be nil

		if len(result.conflicts) > 0 {
			t.Fatalf(
				"Test (%s) — Expected no conflicts but got some: %v",
				v.name,
				result.conflicts,
			)
		}

		sortChanges(v.want)
		sortChanges(result.changes)

		if !cmp.Equal(v.want, result.changes) && len(v.want) != 0 {
			t.Fatalf(
				"Test (%s) — Expected: %+v, got: %+v\n",
				v.name,
				v.want,
				result.changes,
			)
		}

		// Test if the map file was written successfully
		if result.outputFile != "" {
			file, err := os.ReadFile(result.outputFile)
			if err != nil {
				t.Fatalf(
					"Test (%s) — Unexpected error when trying to read map file: %v\n",
					v.name,
					err,
				)
			}

			var mf mapFile
			err = json.Unmarshal(file, &mf)
			if err != nil {
				t.Fatalf(
					"Test (%s) — Unexpected error when trying to unmarshal map file contents: %v\n",
					v.name,
					err,
				)
			}
			ch := mf.Operations

			sortChanges(ch)

			if !cmp.Equal(v.want, ch) && len(v.want) != 0 {
				t.Fatalf(
					"Test (%s) — Expected: %+v, got: %+v\n",
					v.name,
					v.want,
					ch,
				)
			}

			err = os.Remove(result.outputFile)
			if err != nil {
				t.Log("Failed to remove output file")
			}
		}
	}
}

func TestFindReplace(t *testing.T) {
	testDir := setupFileSystem(t)

	cases := []testCase{
		{
			want: []Change{
				{
					Source:  "No Pressure (2021) S1.E1.1080p.mkv",
					BaseDir: testDir,
					Target:  "1.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E2.1080p.mkv",
					BaseDir: testDir,
					Target:  "2.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E3.1080p.mkv",
					BaseDir: testDir,
					Target:  "3.mkv",
				},
			},
			args: []string{
				"-f",
				".*E(\\d+).*",
				"-r",
				"$1.mkv",
				"-o",
				"map.json",
				testDir,
			},
		},
		{
			want: []Change{
				{
					Source:  "No Pressure (2021) S1.E1.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure 98.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E2.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure 99.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E3.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure 100.mkv",
				},
			},
			args: []string{
				"-f",
				"(No Pressure).*",
				"-r",
				"$1 %d.mkv",
				"-n",
				"98",
				testDir,
			},
		},
		{
			want: []Change{
				{
					Source:  "index.js",
					BaseDir: filepath.Join(testDir, "scripts"),
					Target:  "index.ts",
				},
				{
					Source:  "main.js",
					BaseDir: filepath.Join(testDir, "scripts"),
					Target:  "main.ts",
				},
			},
			args: []string{
				"-f",
				"js",
				"-r",
				"ts",
				filepath.Join(testDir, "scripts"),
			},
		},
		{
			want: []Change{
				{
					Source:  "index.js",
					BaseDir: filepath.Join(testDir, "scripts"),
					Target:  "i n d e x .js",
				},
				{
					Source:  "main.js",
					BaseDir: filepath.Join(testDir, "scripts"),
					Target:  "m a i n .js",
				},
			},
			args: []string{
				"-f",
				"(.)",
				"-r",
				"$1 ",
				"-e",
				filepath.Join(testDir, "scripts"),
			},
		},
		{
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
				{
					Source:  "img.jpg",
					BaseDir: filepath.Join(testDir, "morepics", "nested"),
					Target:  "img.jpeg",
				},
			},
			args: []string{
				"-f",
				"jpg",
				"-r",
				"jpeg",
				"-R",
				"-i",
				"-o",
				"map.json",
				testDir,
			},
		},
		{
			want: []Change{
				{
					Source:  "pics",
					IsDir:   true,
					BaseDir: filepath.Join(testDir, "images"),
					Target:  "images",
				},
				{
					Source:  "morepics",
					IsDir:   true,
					BaseDir: testDir,
					Target:  "moreimages",
				},
				{
					Source:  "pic-1.avif",
					BaseDir: filepath.Join(testDir, "morepics"),
					Target:  "image-1.avif",
				},
				{
					Source:  "pic-2.avif",
					BaseDir: filepath.Join(testDir, "morepics"),
					Target:  "image-2.avif",
				},
			},
			args: []string{"-f", "pic", "-r", "image", "-d", "-R", testDir},
		},
		{
			want: []Change{
				{
					Source:  "pics",
					IsDir:   true,
					BaseDir: filepath.Join(testDir, "images"),
					Target:  "images",
				},
				{
					Source:  "morepics",
					IsDir:   true,
					BaseDir: testDir,
					Target:  "moreimages",
				},
			},
			args: []string{"-f", "pic", "-r", "image", "-D", "-R", testDir},
		},
	}

	runFindReplace(t, cases)
}

func TestDetectConflicts(t *testing.T) {
	testDir := setupFileSystem(t)

	type Table struct {
		want map[conflict][]Conflict
		args []string
	}

	table := []Table{
		{
			want: map[conflict][]Conflict{
				fileExists: {
					{
						source: []string{filepath.Join(testDir, "abc.pdf")},
						target: filepath.Join(testDir, "abc.epub"),
					},
				},
			},
			args: []string{"-f", "pdf", "-r", "epub", testDir},
		},
		{
			want: map[conflict][]Conflict{
				emptyFilename: {
					{
						source: []string{filepath.Join(testDir, "abc.pdf")},
						target: filepath.Join(testDir, ""),
					},
				},
			},
			args: []string{"-f", "abc.pdf", "-r", "", testDir},
		},
		{
			want: map[conflict][]Conflict{
				overwritingNewPath: {
					{
						source: []string{
							filepath.Join(testDir, "abc.epub"),
							filepath.Join(testDir, "abc.pdf"),
						},
						target: filepath.Join(testDir, "abc.mobi"),
					},
				},
			},
			args: []string{"-f", "pdf|epub", "-r", "mobi", testDir},
		},
	}

	for i, v := range table {
		args := os.Args[0:1]
		args = append(args, v.args...)
		result, err := action(args)
		if err != nil {
			t.Fatalf("Test(%d) — Unexpected error: %v\n", i+1, err)
		}

		if len(result.conflicts) == 0 {
			t.Fatalf("Test(%d) — Expected some conflicts but got none", i+1)
		}

		if !cmp.Equal(
			v.want,
			result.conflicts,
			cmp.AllowUnexported(Conflict{}),
		) {
			t.Fatalf(
				"Test(%d) — Expected: %+v, got: %+v\n",
				i+1,
				v.want,
				result.conflicts,
			)
		}
	}
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

func TestStringMode(t *testing.T) {
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
			args: []string{"-f", "Pressure", "-r", "Limits", "-s", testDir},
		},
		{
			name: "Replace entire string if find pattern is empty",
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
				"-s",
				"-E",
				"abc|pics",
				testDir,
			},
		},
		{
			name: "Respect case insensitive option",
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
				"-R",
				"-si",
				filepath.Join(testDir, "images"),
			},
		},
	}

	runFindReplace(t, cases)
}

func TestFixConflicts(t *testing.T) {
	testDir := setupFileSystem(t)

	table := []testCase{
		{
			want: []Change{
				{
					Source:  "abc.txt",
					BaseDir: filepath.Join(testDir, "conflicts"),
					Target:  "123 (2).txt",
				},
				{
					Source:  "xyz.txt",
					BaseDir: filepath.Join(testDir, "conflicts"),
					Target:  "123 (4).txt",
				},
			},
			args: []string{
				"-f",
				"abc|xyz",
				"-r",
				"123",
				"-F",
				filepath.Join(testDir, "conflicts"),
			},
		},
		{
			want: []Change{
				{
					Source:  "123.txt",
					BaseDir: filepath.Join(testDir, "conflicts"),
					Target:  "abc (2).txt",
				},
				{
					Source:  "123 (3).txt",
					BaseDir: filepath.Join(testDir, "conflicts"),
					Target:  "abc (3).txt",
				},
			},
			args: []string{
				"-f",
				"123",
				"-r",
				"abc",
				"-F",
				filepath.Join(testDir, "conflicts"),
			},
		},
		{
			want: []Change{
				{
					Source:  "xyz.txt",
					BaseDir: filepath.Join(testDir, "conflicts"),
					Target:  "123 (2).txt",
				},
			},
			args: []string{
				"-f",
				"xyz",
				"-r",
				"123",
				"-F",
				filepath.Join(testDir, "conflicts"),
			},
		},
		{
			want: []Change{
				{
					Source:  "xyz.txt",
					BaseDir: filepath.Join(testDir, "conflicts"),
					Target:  "xyz.txt",
				},
			},
			args: []string{
				"-f",
				"xyz.txt",
				"-F",
				filepath.Join(testDir, "conflicts"),
			},
		},
	}

	for i, v := range table {
		args := os.Args[0:1]
		args = append(args, v.args...)
		result, _ := action(args) // err will be nil

		if len(result.conflicts) == 0 {
			t.Fatalf("Test(%d) — Expected some conflicts but got none", i+1)
		}

		sortChanges(v.want)
		sortChanges(result.changes)

		if !cmp.Equal(v.want, result.changes) && len(v.want) != 0 {
			t.Fatalf(
				"Test(%d) — Expected: %+v, got: %+v\n",
				i+1,
				v.want,
				result.changes,
			)
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
			args: []string{
				"-f",
				".*E(\\d+).*",
				"-r",
				"$1.mkv",
				"-o",
				"map.json",
				"-x",
			},
			undoArgs: []string{"-u", "map.json", "-x"},
		},
		{
			want: []Change{
				{Source: "pics", IsDir: true, Target: "images"},
				{Source: "morepics", IsDir: true, Target: "moreimages"},
				{Source: "pic-1.avif", Target: "image-1.avif"},
				{Source: "pic-2.avif", Target: "image-2.avif"},
			},
			args: []string{
				"-f",
				"pic",
				"-r",
				"image",
				"-d",
				"-R",
				"-o",
				"map.json",
				"-x",
			},
			undoArgs: []string{"-u", "map.json", "-x"},
		},
	}

	for i, v := range table {
		testDir := setupFileSystem(t)

		for _, ch := range v.want {
			ch.BaseDir = testDir
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

		// Test Undo function
		args = os.Args[0:1]
		args = append(args, v.undoArgs...)
		result, err := action(args)
		if err != nil {
			t.Fatalf("Test(%d) — Unexpected error in undo mode: %v\n", i+1, err)
		}

		err = os.Remove(result.outputFile)
		if err != nil {
			t.Log("Failed to remove output file")
		}
	}
}

func randate() time.Time {
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

func TestAutoIncrementingNumber(t *testing.T) {
	testDir := setupFileSystem(t)
	op := &Operation{}
	op.replacement = "%03d{{ext}}"
	op.searchRegex = regexp.MustCompile(".*")
	slice := []int{4, 10, 100, 150, 44, 82, 1000, 321}
	for _, start := range slice {
		op.startNumber = start

		for _, path := range fileSystem {
			dir := filepath.Dir(path)
			ch := Change{
				BaseDir: filepath.Join(testDir, dir),
				Source:  filepath.Base(path),
			}
			op.matches = append(op.matches, ch)
		}

		op.SortMatches()
		err := op.replace()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		sort.SliceStable(op.matches, func(i, j int) bool {
			regex := regexp.MustCompile(`\d+`)
			inum, err := strconv.Atoi(regex.FindString(op.matches[i].Target))
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			jnum, err := strconv.Atoi(regex.FindString(op.matches[j].Target))
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			return inum < jnum
		})

		for i, v := range op.matches {
			ext := filepath.Ext(v.Source)
			index := fmt.Sprintf("%03d", i+start)

			want := index + ext
			if want != v.Target {
				t.Fatalf("Expected: %s, but got: %s", want, v.Target)
			}
		}
	}
}

func TestReplaceFilenameVariables(t *testing.T) {
	testDir := setupFileSystem(t)

	for _, path := range fileSystem {
		fullPath := filepath.Join(testDir, path)
		base := filenameWithoutExtension(filepath.Base(path))
		ext := filepath.Ext(path)
		dir := filepath.Dir(path)
		ch := Change{
			BaseDir: filepath.Join(testDir, dir),
			Source:  filepath.Base(path),
		}

		op := &Operation{}
		got, err := op.handleVariables("{{p}}-{{f}}{{ext}}", ch)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		want := fmt.Sprintf(
			"%s-%s%s",
			filepath.Base(filepath.Dir(fullPath)),
			base,
			ext,
		)
		if got != want {
			t.Fatalf("Expected: %s, but got: %s", want, got)
		}
	}
}

func TestReplaceDateVariables(t *testing.T) {
	testDir := setupFileSystem(t)

	for _, file := range fileSystem {
		path := filepath.Join(testDir, file)

		// change the atime and mtime to a random value
		mtime, atime := randate(), randate()
		err := os.Chtimes(path, atime, mtime)
		if err != nil {
			t.Fatalf("Expected no errors, but got one: %v\n", err)
		}

		timeInfo, err := times.Stat(path)
		if err != nil {
			t.Fatalf("Expected no errors, but got one: %v\n", err)
		}

		want := make(map[string]string)
		got := make(map[string]string)

		accessTime := timeInfo.AccessTime()
		modTime := timeInfo.ModTime()

		fileTimes := []string{"mtime", "atime", "ctime", "now", "btime"}

		for _, v := range fileTimes {
			var timeValue time.Time
			switch v {
			case "mtime":
				timeValue = modTime
			case "atime":
				timeValue = accessTime
			case "ctime":
				timeValue = modTime
				if timeInfo.HasChangeTime() {
					timeValue = timeInfo.ChangeTime()
				}
			case "btime":
				timeValue = modTime
				if timeInfo.HasBirthTime() {
					timeValue = timeInfo.BirthTime()
				}
			case "now":
				timeValue = time.Now()
			}

			for key, token := range dateTokens {
				want[v+"."+key] = timeValue.Format(token)
				out, err := replaceDateVariables(path, "{{"+v+"."+key+"}}")
				if err != nil {
					t.Fatalf("Expected no errors, but got one: %v\n", err)
				}
				got[v+"."+key] = out
			}
		}

		if !cmp.Equal(want, got) {
			t.Fatalf("Expected %v, but got %v\n", want, got)
		}
	}
}

func TestReplaceExifVariables(t *testing.T) {
	rootDir := filepath.Join("testdata", "images")

	type FileExif struct {
		Year         string `json:"year"`
		Make         string `json:"make"`
		Model        string `json:"model"`
		ISO          int    `json:"iso"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		Dimensions   string `json:"dimensions"`
		ExposureTime string `json:"exposure_time"`
		FocalLength  string `json:"focal_length"`
		Aperture     string `json:"aperture"`
	}

	cases := []testCase{
		{
			name: "Use EXIF data to rename CR2 file",
			want: []Change{
				{
					Source:  "tractor-raw.cr2",
					BaseDir: rootDir,
				},
			},
			args: []string{
				"-f",
				"tractor-raw.cr2",
				"-r",
				"{{exif.dt.YYYY}}_{{exif.make}}_{{exif.model}}_{{exif.iso}}_w{{exif.w}}_h{{exif.h}}_{{exif.wh}}_{{exif.et}}_{{exif.fl}}_{{exif.fnum}}{{ext}}",
				rootDir,
			},
		},
		{
			name: "Use EXIF data to rename JPEG file",
			want: []Change{
				{
					Source:  "bike.jpeg",
					BaseDir: rootDir,
				},
			},
			args: []string{
				"-f",
				"bike.jpeg",
				"-r",
				"{{exif.dt.YYYY}}_{{exif.make}}_{{exif.model}}_{{exif.iso}}_w{{exif.w}}_h{{exif.h}}_{{exif.wh}}_{{exif.et}}_{{exif.fl}}_{{exif.fnum}}{{ext}}",
				rootDir,
			},
		},
		{
			name: "Use EXIF data to rename DNG file",
			want: []Change{
				{
					Source:  "proraw.dng",
					BaseDir: rootDir,
				},
			},
			args: []string{
				"-f",
				"proraw.dng",
				"-r",
				"{{exif.dt.YYYY}}_{{exif.make}}_{{exif.model}}_{{exif.iso}}_w{{exif.h}}_h{{exif.w}}_{{exif.h}}x{{exif.w}}_{{exif.et}}_{{exif.fl}}_{{exif.fnum}}{{ext}}",
				rootDir,
			},
		},
	}

	for _, c := range cases {
		f := filenameWithoutExtension(c.want[0].Source)
		ext := filepath.Ext(c.want[0].Source)

		jsonFile, err := os.ReadFile(filepath.Join(rootDir, f+".json"))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		var exif FileExif
		err = json.Unmarshal(jsonFile, &exif)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		target := fmt.Sprintf(
			"%s_%s_%s_ISO%d_w%d_h%d_%s_%ss_%smm_f%s%s",
			exif.Year,
			exif.Make,
			exif.Model,
			exif.ISO,
			exif.Width,
			exif.Height,
			exif.Dimensions,
			exif.ExposureTime,
			exif.FocalLength,
			exif.Aperture,
			ext,
		)

		c.want[0].Target = target
	}

	runFindReplace(t, cases)
}
