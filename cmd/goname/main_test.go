package main

import (
	"encoding/json"
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
	mapFile   string
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

		if op.outputFile != "" {
			result.mapFile = op.outputFile
			op.WriteToFile()
		}

		return nil
	}

	return result, app.Run(args)
}

func sortChanges(s []Change) {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Source < s[j].Source
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
				{Source: "No Pressure (2021) S1.E1.1080p.mkv", BaseDir: testDir, Target: "1.mkv"},
				{Source: "No Pressure (2021) S1.E2.1080p.mkv", BaseDir: testDir, Target: "2.mkv"},
				{Source: "No Pressure (2021) S1.E3.1080p.mkv", BaseDir: testDir, Target: "3.mkv"},
			},
			args: []string{"-f", ".*E(\\d+).*", "-r", "$1.mkv", "-o", "map.json", testDir},
		},
		{
			want: []Change{
				{Source: "No Pressure (2021) S1.E1.1080p.mkv", BaseDir: testDir, Target: "No Pressure 98.mkv"},
				{Source: "No Pressure (2021) S1.E2.1080p.mkv", BaseDir: testDir, Target: "No Pressure 99.mkv"},
				{Source: "No Pressure (2021) S1.E3.1080p.mkv", BaseDir: testDir, Target: "No Pressure 100.mkv"},
			},
			args: []string{"-f", "(No Pressure).*", "-r", "$1 %d.mkv", "-n", "98", testDir},
		},
		{
			want: []Change{
				{Source: "a.jpg", BaseDir: filepath.Join(testDir, "images"), Target: "a.jpeg"},
			},
			args: []string{"-f", "jpg", "-r", "jpeg", "-R", testDir},
		},
		{
			want: []Change{
				{Source: "456.webp", BaseDir: filepath.Join(testDir, "images"), Target: "456-001.webp"},
				{Source: "a.jpg", BaseDir: filepath.Join(testDir, "images"), Target: "a-002.jpg"},
				{Source: "abc.png", BaseDir: filepath.Join(testDir, "images"), Target: "abc-003.png"},
			},
			args: []string{"-f", ".*(jpg|png|webp)", "-r", "{og}-%03d.$1", filepath.Join(testDir, "images")},
		},
		{
			want: []Change{
				{Source: "456.webp", BaseDir: filepath.Join(testDir, "images"), Target: "001.webp"},
				{Source: "a.jpg", BaseDir: filepath.Join(testDir, "images"), Target: "002.jpg"},
				{Source: "abc.png", BaseDir: filepath.Join(testDir, "images"), Target: "003.png"},
			},
			args: []string{"-f", ".*(jpg|png|webp)", "-r", "%03d{ext}", filepath.Join(testDir, "images")},
		},
		{
			want: []Change{
				{Source: "index.js", BaseDir: filepath.Join(testDir, "scripts"), Target: "index.ts"},
				{Source: "main.js", BaseDir: filepath.Join(testDir, "scripts"), Target: "main.ts"},
			},
			args: []string{"-f", "js", "-r", "ts", filepath.Join(testDir, "scripts")},
		},
		{
			want: []Change{
				{Source: "index.js", BaseDir: filepath.Join(testDir, "scripts"), Target: "i n d e x .js"},
				{Source: "main.js", BaseDir: filepath.Join(testDir, "scripts"), Target: "m a i n .js"},
			},
			args: []string{"-f", "(.)", "-r", "$1 ", "-e", filepath.Join(testDir, "scripts")},
		},
		{
			want: []Change{
				{Source: "a.jpg", BaseDir: filepath.Join(testDir, "images"), Target: "a.jpeg"},
				{Source: "123.JPG", BaseDir: filepath.Join(testDir, "images", "pics"), Target: "123.jpeg"},
			},
			args: []string{"-f", "jpg", "-r", "jpeg", "-R", "-i", "-o", "map.json", testDir},
		},
		{
			want: []Change{
				{Source: "pics", IsDir: true, BaseDir: filepath.Join(testDir, "images"), Target: "images"},
				{Source: "morepics", IsDir: true, BaseDir: testDir, Target: "moreimages"},
				{Source: "pic-1.avif", BaseDir: filepath.Join(testDir, "morepics"), Target: "image-1.avif"},
				{Source: "pic-2.avif", BaseDir: filepath.Join(testDir, "morepics"), Target: "image-2.avif"},
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

		// Test if the map file was written successfully
		if result.mapFile != "" {
			file, err := os.ReadFile(result.mapFile)
			if err != nil {
				t.Fatalf("Unexpected error when trying to read map file: %v\n", err)
			}

			ch := []Change{}
			err = json.Unmarshal([]byte(file), &ch)
			if err != nil {
				t.Fatalf("Unexpected error when trying to unmarshal map file contents: %v\n", err)
			}

			sortChanges(ch)

			if !reflect.DeepEqual(v.want, ch) && len(v.want) != 0 {
				t.Fatalf("Test(%d) — Expected: %+v, got: %+v\n", i+1, v.want, ch)
			}

			err = os.Remove(result.mapFile)
			if err != nil {
				t.Log("Failed to remove log file")
			}
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
