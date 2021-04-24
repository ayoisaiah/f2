package f2

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/urfave/cli/v2"
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
		op, err := newOperation(c)
		if err != nil {
			return err
		}

		op.quiet = true
		if op.undoFile != "" {
			result.outputFile = op.undoFile
			return op.undo()
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
			op.sortMatches()
		}

		err = op.replace()
		if err != nil {
			return err
		}

		result.changes = op.matches

		if op.outputFile != "" {
			result.outputFile = op.outputFile
			err = op.writeToFile(op.outputFile)
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
				prettyPrint(v.want),
				prettyPrint(result.changes),
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
					prettyPrint(v.want),
					prettyPrint(ch),
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
				"$1 98%d.mkv",
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
