package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
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
	"images/pics/123.jpg",
}

// setup creates all required files and folders for
// the tests and returns a function that is used as
// a teardown function when the tests are done.
func setup(t testing.TB) (string, func()) {
	testDir, err := ioutil.TempDir(".", "")
	if err != nil {
		os.RemoveAll(testDir)
		t.Fatal(err)
	}

	directories := []string{"images/pics"}
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

func TestFindMatches(t *testing.T) {
	testDir, teardown := setup(t)

	defer teardown()

	var result []Change

	app := getApp()
	app.Action = func(c *cli.Context) error {
		op, err := NewOperation(c)
		if err != nil {
			return err
		}

		err = op.FindMatches()
		if err != nil {
			return err
		}

		result = op.matches

		return nil
	}

	table := []struct {
		want []Change
		args []string
	}{
		{
			want: []Change{
				{source: "No Pressure (2021) S1.E1.1080p.mkv", baseDir: testDir},
				{source: "No Pressure (2021) S1.E2.1080p.mkv", baseDir: testDir},
				{source: "No Pressure (2021) S1.E3.1080p.mkv", baseDir: testDir},
			},
			args: []string{"-f", ".*E(\\d+).*", "-r", "", testDir},
		},
		{
			want: []Change{},
			args: []string{"-f", "images", "-r", "", testDir},
		},
		{
			want: []Change{
				{source: "images", isDir: true, baseDir: testDir},
			},
			args: []string{"-f", "images", "-r", "", "-D", testDir},
		},
		{
			want: []Change{
				{source: "a.jpg", baseDir: filepath.Join(testDir, "images")},
				{source: "abc.png", baseDir: filepath.Join(testDir, "images")},
				{source: "123.jpg", baseDir: filepath.Join(testDir, "images", "pics")},
			},
			args: []string{"-f", "jpg|png", "-r", "", "-D", "-R", testDir},
		},
	}

	for i, v := range table {
		args := os.Args[0:1]
		args = append(args, v.args...)
		err := app.Run(args)
		if err != nil {
			t.Fatalf("Unexpected error occurred: %v", err)
		}

		if err != nil {
			t.Errorf("Test(%d) — An error occurred while finding matches: %v", i+1, err)
		}

		if !reflect.DeepEqual(v.want, result) && len(v.want) != 0 {
			t.Fatalf("Test(%d) — Expected: %v, got: %v", i+1, v.want, result)
		}
	}
}
