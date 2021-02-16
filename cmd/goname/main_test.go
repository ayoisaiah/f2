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
	"scripts/index.js",
	"scripts/main.js",
	"a-b-c.txt",
}

type Table struct {
	want []Change
	args []string
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

	directories := []string{"images/pics", "scripts"}
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

func replaceAction(args []string) ([]Change, error) {
	var result []Change

	app := getApp()
	app.Action = func(c *cli.Context) error {
		op, err := NewOperation(c)
		if err != nil {
			return err
		}

		op.FindMatches()
		if err != nil {
			return err
		}

		if op.includeDir {
			op.SortMatches()
		}

		op.Replace()

		result = op.matches

		return nil
	}

	return result, app.Run(args)
}

func sortChanges(s []Change) {
	sort.Slice(s, func(i, j int) bool {
		return s[i].source < s[j].source
	})
}

func loop(t *testing.T, fn func([]string) ([]Change, error), table []Table) {
	for i, v := range table {
		args := os.Args[0:1]
		args = append(args, v.args...)
		result, err := fn(args)
		if err != nil {
			t.Fatalf("Unexpected error occurred: %v", err)
		}

		if err != nil {
			t.Errorf("Test(%d) — Unexpected error: %v", i+1, err)
		}

		sortChanges(v.want)
		sortChanges(result)

		if !reflect.DeepEqual(v.want, result) && len(v.want) != 0 {
			t.Fatalf("Test(%d) — Expected: %+v, got: %+v", i+1, v.want, result)
		}
	}
}

func TestFindReplace(t *testing.T) {
	testDir, teardown := setupFileSystem(t)

	defer teardown()

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
	}

	loop(t, replaceAction, table)
}
