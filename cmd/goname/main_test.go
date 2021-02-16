package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/urfave/cli/v2"
)

var fileSystem = []string{
	"No Pressure (2021) S1.E1.1080p.mkv",
	"No Pressure (2021) S1.E2.1080p.mkv",
	"No Pressure (2021) S1.E3.1080p.mkv",
	"images/a.jpg",
	"images/abc.png",
	"images/456.webp",
	"images/pics/123.JPG",
	"morepics/pic-1.avif",
	"morepics/pic-2.avif",
	"scripts/index.js",
	"scripts/main.js",
	"abc.pdf",
	"abc.epub",
	".pics",
}

// setupFileSystem creates all required files and folders for
// the tests and returns a function that is used as
// a teardown function when the tests are done.
func setupFileSystem(t testing.TB) (string, func()) {
	testDir, err := ioutil.TempDir(".", "")
	if err != nil {
		os.RemoveAll(testDir)
		t.Fatal(err)
	}

	directories := []string{"images/pics", "scripts", "morepics"}
	for _, v := range directories {
		filePath := filepath.Join(testDir, v)
		err = os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			os.RemoveAll(testDir)
			t.Fatal(err)
		}
	}

	for _, f := range fileSystem {
		filePath := filepath.Join(testDir, f)
		if err := ioutil.WriteFile(filePath, []byte{}, 0755); err != nil {
			os.RemoveAll(testDir)
			t.Fatal(err)
		}
	}

	abs, err := filepath.Abs(testDir)
	if err != nil {
		os.RemoveAll(testDir)
		t.Fatal(err)
	}

	return abs, func() {
		if os.RemoveAll(testDir); err != nil {
			t.Fatal(err)
		}
	}
}

type ActionResult struct {
	changes   []Change
	conflicts map[conflict][]Conflict
}

func action(args []string) (ActionResult, error) {
	var result ActionResult

	app := getApp()
	app.Action = func(c *cli.Context) error {
		op, err := NewOperation(c)
		if err != nil {
			return err
		}

		op.FindMatches()

		if op.includeDir {
			op.SortMatches()
		}

		op.Replace()

		result.changes = op.matches
		result.conflicts = op.DetectConflicts()

		return nil
	}

	return result, app.Run(args)
}

func sortChanges(s []Change) {
	sort.Slice(s, func(i, j int) bool {
		return s[i].source < s[j].source
	})
}

func TestFindReplace(t *testing.T) {
	testDir, teardown := setupFileSystem(t)

	defer teardown()

	type Table struct {
		want []Change
		args []string
	}

	table := []Table{
		{
			want: []Change{
				{source: "No Pressure (2021) S1.E1.1080p.mkv", baseDir: testDir, target: "1.mkv"},
				{source: "No Pressure (2021) S1.E2.1080p.mkv", baseDir: testDir, target: "2.mkv"},
				{source: "No Pressure (2021) S1.E3.1080p.mkv", baseDir: testDir, target: "3.mkv"},
			},
			args: []string{"-f", ".*E(\\d+).*", "-r", "$1.mkv", testDir},
		},
		{
			want: []Change{
				{source: "No Pressure (2021) S1.E1.1080p.mkv", baseDir: testDir, target: "No Pressure 98.mkv"},
				{source: "No Pressure (2021) S1.E2.1080p.mkv", baseDir: testDir, target: "No Pressure 99.mkv"},
				{source: "No Pressure (2021) S1.E3.1080p.mkv", baseDir: testDir, target: "No Pressure 100.mkv"},
			},
			args: []string{"-f", "(No Pressure).*", "-r", "$1 %d.mkv", "-n", "98", testDir},
		},
		{
			want: []Change{
				{source: "a.jpg", baseDir: filepath.Join(testDir, "images"), target: "a.jpeg"},
			},
			args: []string{"-f", "jpg", "-r", "jpeg", "-R", testDir},
		},
		{
			want: []Change{
				{source: "456.webp", baseDir: filepath.Join(testDir, "images"), target: "456-001.webp"},
				{source: "a.jpg", baseDir: filepath.Join(testDir, "images"), target: "a-002.jpg"},
				{source: "abc.png", baseDir: filepath.Join(testDir, "images"), target: "abc-003.png"},
			},
			args: []string{"-f", ".*(jpg|png|webp)", "-r", "{og}-%03d.$1", filepath.Join(testDir, "images")},
		},
		{
			want: []Change{
				{source: "456.webp", baseDir: filepath.Join(testDir, "images"), target: "001.webp"},
				{source: "a.jpg", baseDir: filepath.Join(testDir, "images"), target: "002.jpg"},
				{source: "abc.png", baseDir: filepath.Join(testDir, "images"), target: "003.png"},
			},
			args: []string{"-f", ".*(jpg|png|webp)", "-r", "%03d{ext}", filepath.Join(testDir, "images")},
		},
		{
			want: []Change{
				{source: "index.js", baseDir: filepath.Join(testDir, "scripts"), target: "index.ts"},
				{source: "main.js", baseDir: filepath.Join(testDir, "scripts"), target: "main.ts"},
			},
			args: []string{"-f", "js", "-r", "ts", filepath.Join(testDir, "scripts")},
		},
		{
			want: []Change{
				{source: "index.js", baseDir: filepath.Join(testDir, "scripts"), target: "i n d e x .js"},
				{source: "main.js", baseDir: filepath.Join(testDir, "scripts"), target: "m a i n .js"},
			},
			args: []string{"-f", "(.)", "-r", "$1 ", "-e", filepath.Join(testDir, "scripts")},
		},
		{
			want: []Change{
				{source: "a.jpg", baseDir: filepath.Join(testDir, "images"), target: "a.jpeg"},
				{source: "123.JPG", baseDir: filepath.Join(testDir, "images", "pics"), target: "123.jpeg"},
			},
			args: []string{"-f", "jpg", "-r", "jpeg", "-R", "-i", testDir},
		},
		{
			want: []Change{
				{source: "pics", isDir: true, baseDir: filepath.Join(testDir, "images"), target: "images"},
				{source: "morepics", isDir: true, baseDir: testDir, target: "moreimages"},
				{source: "pic-1.avif", baseDir: filepath.Join(testDir, "morepics"), target: "image-1.avif"},
				{source: "pic-2.avif", baseDir: filepath.Join(testDir, "morepics"), target: "image-2.avif"},
			},
			args: []string{"-f", "pic", "-r", "image", "-D", "-R", testDir},
		},
	}

	for i, v := range table {
		args := os.Args[0:1]
		args = append(args, v.args...)
		result, err := action(args)
		if err != nil {
			t.Fatalf("Test(%d) — Unexpected error: %v\n", i+1, err)
		}

		if len(result.conflicts) > 0 {
			t.Fatalf("Test(%d) — Expected no conflicts but got some: %v", i+1, result.conflicts)
		}

		sortChanges(v.want)
		sortChanges(result.changes)

		if !reflect.DeepEqual(v.want, result.changes) && len(v.want) != 0 {
			t.Fatalf("Test(%d) — Expected: %+v, got: %+v\n", i+1, v.want, result.changes)
		}
	}
}

func TestDetectConflicts(t *testing.T) {
	testDir, teardown := setupFileSystem(t)

	defer teardown()

	type Table struct {
		want map[conflict][]Conflict
		args []string
	}

	table := []Table{
		{
			want: map[conflict][]Conflict{
				FILE_EXISTS: []Conflict{
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
				EMPTY_FILENAME: []Conflict{
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
				OVERWRITNG_NEW_PATH: []Conflict{
					{
						source: []string{filepath.Join(testDir, "abc.epub"), filepath.Join(testDir, "abc.pdf")},
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

		if !reflect.DeepEqual(v.want, result.conflicts) {
			t.Fatalf("Test(%d) — Expected: %+v, got: %+v\n", i+1, v.want, result.conflicts)
		}
	}
}
