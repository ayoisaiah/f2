package f2

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type conflictTable struct {
	name string
	want map[conflictType][]Conflict
	args []string
}

func runConflictCheck(t *testing.T, table []conflictTable) {
	for _, v := range table {
		args := os.Args[0:1]
		args = append(args, v.args...)
		result, err := action(args)
		if err != nil {
			t.Fatalf("Test (%s) — Unexpected error: %v\n", v.name, err)
		}

		if len(result.conflicts) == 0 {
			t.Fatalf("Test (%s) — Expected some conflicts but got none", v.name)
		}

		if !cmp.Equal(
			v.want,
			result.conflicts,
			cmp.AllowUnexported(Conflict{}),
		) {
			t.Fatalf(
				"Test (%s) — Expected: %+v, got: %+v\n",
				v.name,
				v.want,
				result.conflicts,
			)
		}
	}
}

func runFixConflict(t *testing.T, table []testCase) {
	for _, v := range table {
		args := os.Args[0:1]
		args = append(args, v.args...)
		result, _ := action(args) // err will be nil

		if len(result.conflicts) == 0 {
			t.Fatalf("Test (%s) — Expected some conflicts but got none", v.name)
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

func TestDetectConflicts(t *testing.T) {
	testDir := setupFileSystem(t)

	table := []conflictTable{
		{
			name: "File exists",
			want: map[conflictType][]Conflict{
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
			name: "Empty filename",
			want: map[conflictType][]Conflict{
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
			name: "Overwriting newly renamed path",
			want: map[conflictType][]Conflict{
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

	runConflictCheck(t, table)
}

func TestFixConflicts(t *testing.T) {
	testDir := setupFileSystem(t)

	table := []testCase{
		{
			name: "Fix path already exists conflict",
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
			name: "Fix path exists conflict",
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
			name: "Fix overwriting new path conflict",
			want: []Change{
				{
					Source:  "abc.txt",
					BaseDir: filepath.Join(testDir, "conflicts"),
					Target:  "man.txt",
				},
				{
					Source:  "xyz.txt",
					BaseDir: filepath.Join(testDir, "conflicts"),
					Target:  "man (2).txt",
				},
			},
			args: []string{
				"-f",
				"abc|xyz",
				"-r",
				"man",
				"-F",
				filepath.Join(testDir, "conflicts"),
			},
		},
		{
			name: "Fix empty filename conflict",
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

	runFixConflict(t, table)
}

func TestReportConflicts(t *testing.T) {
	testDir := setupFileSystem(t)

	table := map[conflictType][]Conflict{
		fileExists: {
			{
				source: []string{filepath.Join(testDir, "abc.pdf")},
				target: filepath.Join(testDir, "abc.epub"),
			},
		},
		emptyFilename: {
			{
				source: []string{filepath.Join(testDir, "abc.pdf")},
				target: filepath.Join(testDir, ""),
			},
		},
		invalidCharacters: {
			{
				source: []string{filepath.Join(testDir, "abc.pdf")},
				target: filepath.Join(testDir, "%^&*().pdf"),
			},
		},
		overwritingNewPath: {
			{
				source: []string{
					filepath.Join(testDir, "abc.epub"),
					filepath.Join(testDir, "abc.pdf"),
				},
				target: filepath.Join(testDir, "abc.mobi"),
			},
		},
		maxLengthExceeded: {
			{
				source: []string{
					filepath.Join(testDir, "abc.pdf"),
				},
				target: filepath.Join(
					testDir,
					"😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀.mobi",
				),
			},
		},
	}

	op := &Operation{}
	op.conflicts = table
	rescueStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	os.Stdout = w

	op.reportConflicts()

	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	os.Stdout = rescueStdout

	if string(out) == "" {
		t.Fatal(
			"Expected output to be a non-empty string but, got an empty string",
		)
	}
}

func TestGetNewPath(t *testing.T) {
	type m map[string][]struct {
		source string
		index  int
	}
	cases := []struct {
		input  string
		output string
		m      m
	}{
		{
			input:  "an_image.png",
			output: "an_image (2).png",
			m:      nil,
		},
		{
			input:  "an_image (2).png",
			output: "an_image (3).png",
			m:      nil,
		},
		{
			input:  "an_image (4).png",
			output: "an_image (5).png",
			m:      nil,
		},
		{
			input:  "an_image (8).png",
			output: "an_image (12).png",
			m: m{
				"an_image (8).png": {
					{
						source: "img.png",
						index:  3,
					},
				},
				"an_image (9).png": {
					{
						source: "img-2.png",
						index:  5,
					},
				},
				"an_image (10).png": {
					{
						source: "img-3.png",
						index:  8,
					},
				},
				"an_image (11).png": {
					{
						source: "img-4.png",
						index:  6,
					},
				},
			},
		},
	}

	for _, v := range cases {
		out := getNewPath(v.input, ".", v.m)
		if out != v.output {
			t.Fatalf(
				"Incorrect output from getNewPath. Want: %s, got %s",
				v.output,
				out,
			)
		}
	}
}
