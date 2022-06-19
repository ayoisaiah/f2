package f2

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type conflictTable struct {
	name string
	want map[conflictType][]Conflict
	args string
}

func runConflictCheckHelper(t *testing.T, table []conflictTable) {
	t.Helper()

	for _, tc := range table {
		args := parseArgs(t, tc.name, tc.args)

		result, err := testRun(args)
		if err != nil {
			if !errors.Is(err, errConflictDetected) {
				t.Fatalf("Test (%s) — Unexpected error: %v\n", tc.name, err)
			}
		}

		if len(result.conflicts) == 0 {
			t.Fatalf(
				"Test (%s) — Expected some conflicts but got none",
				tc.name,
			)
		}

		if !cmp.Equal(
			tc.want,
			result.conflicts,
			cmp.AllowUnexported(Conflict{}),
		) {
			t.Fatalf(
				"Test (%s) — Expected: %+v, got: %+v\n",
				tc.name,
				tc.want,
				result.conflicts,
			)
		}
	}
}

func runFixConflictHelper(t *testing.T, table []testCase) {
	t.Helper()

	for _, tc := range table {
		args := parseArgs(t, tc.name, tc.args)

		result, err := testRun(args)
		if err != nil {
			t.Fatalf("Test (%s) — Unexpected error from F2: %v", tc.name, err)
		}

		if len(result.conflicts) == 0 {
			t.Fatalf(
				"Test (%s) — Expected some conflicts but got none",
				tc.name,
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
				"Test (%s) — Expected: %+v, got: %+v\n",
				tc.name,
				prettyPrint(tc.want),
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
						Sources: []string{filepath.Join(testDir, "abc.pdf")},
						Target:  filepath.Join(testDir, "abc.epub"),
					},
				},
			},
			args: "-f pdf -r epub " + testDir,
		},
		{
			name: "Empty filename",
			want: map[conflictType][]Conflict{
				emptyFilename: {
					{
						Sources: []string{filepath.Join(testDir, "abc.pdf")},
						Target:  filepath.Join(testDir, ""),
					},
				},
			},
			args: "-f abc.pdf -r '' " + testDir,
		},
		{
			name: "Overwriting newly renamed path",
			want: map[conflictType][]Conflict{
				overwritingNewPath: {
					{
						Sources: []string{
							filepath.Join(testDir, "abc.epub"),
							filepath.Join(testDir, "abc.pdf"),
						},
						Target: filepath.Join(testDir, "abc.mobi"),
					},
				},
			},
			args: "-f pdf|epub -r mobi " + testDir,
		},
	}

	runConflictCheckHelper(t, table)
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
			args: "-f abc|xyz -r 123 -F " + filepath.Join(testDir, "conflicts"),
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
			args: "-f 123 -r abc -F " + filepath.Join(testDir, "conflicts"),
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
			args: "-f abc|xyz -r man -F " + filepath.Join(testDir, "conflicts"),
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
			args: "-f xyz.txt -F " + filepath.Join(testDir, "conflicts"),
		},
	}

	runFixConflictHelper(t, table)
}

func TestReportConflicts(t *testing.T) {
	testDir := setupFileSystem(t)

	table := map[conflictType][]Conflict{
		fileExists: {
			{
				Sources: []string{filepath.Join(testDir, "abc.pdf")},
				Target:  filepath.Join(testDir, "abc.epub"),
			},
		},
		emptyFilename: {
			{
				Sources: []string{filepath.Join(testDir, "abc.pdf")},
				Target:  filepath.Join(testDir, ""),
			},
		},
		trailingPeriod: {
			{
				Sources: []string{filepath.Join(testDir, "abc.pdf")},
				Target:  filepath.Join(testDir, "abc.pdf."),
			},
		},
		invalidCharacters: {
			{
				Sources: []string{filepath.Join(testDir, "abc.pdf")},
				Target:  filepath.Join(testDir, "%^&*().pdf"),
			},
		},
		overwritingNewPath: {
			{
				Sources: []string{
					filepath.Join(testDir, "abc.epub"),
					filepath.Join(testDir, "abc.pdf"),
				},
				Target: filepath.Join(testDir, "abc.mobi"),
			},
		},
		maxFilenameLengthExceeded: {
			{
				Sources: []string{
					filepath.Join(testDir, "abc.pdf"),
				},
				Target: filepath.Join(
					testDir,
					"😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀😀.mobi",
				),
			},
		},
	}

	var buf bytes.Buffer

	op := &Operation{
		stdout: &buf,
	}
	op.conflicts = table

	op.reportConflicts()

	if buf.String() == "" {
		t.Fatal(
			"Expected output to be a non-empty string but, got an empty string",
		)
	}
}

func TestGetNewPath(t *testing.T) {
	type m map[string][]struct {
		sourcePath string
		index      int
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
						sourcePath: "img.png",
						index:      3,
					},
				},
				"an_image (9).png": {
					{
						sourcePath: "img-2.png",
						index:      5,
					},
				},
				"an_image (10).png": {
					{
						sourcePath: "img-3.png",
						index:      8,
					},
				},
				"an_image (11).png": {
					{
						sourcePath: "img-4.png",
						index:      6,
					},
				},
			},
		},
	}

	for _, v := range cases {
		ch := Change{
			Target:  v.input,
			BaseDir: ".",
		}

		out := newTarget(&ch, v.m)
		if out != v.output {
			t.Fatalf(
				"Incorrect output from getNewPath. Want: %s, got %s",
				v.output,
				out,
			)
		}
	}
}

func TestExstingTargetChanging(t *testing.T) {
	testDir := setupFileSystem(t)

	cases := []testCase{
		{
			name: "Existing target paths that are changing should not trigger a conflict",
			want: []Change{
				{
					BaseDir: filepath.Join(testDir, "morepics"),
					Source:  "pic-1.avif",
					Target:  "pic-2.avif",
				},
				{
					BaseDir: filepath.Join(testDir, "morepics"),
					Source:  "pic-2.avif",
					Target:  "pic-3.avif",
				},
			},
			args: "-f '\\d' -r 2%d " + filepath.Join(testDir, "morepics"),
		},
	}

	runFindReplaceHelper(t, cases)
}
