package f2_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ayoisaiah/f2/v2"
	"github.com/ayoisaiah/f2/v2/internal/testutil"
)

func TestImagePairRenaming(t *testing.T) {
	var stdout bytes.Buffer

	var stdin bytes.Buffer

	var stderr bytes.Buffer

	app, err := f2.New(&stdin, &stdout)
	if err != nil {
		t.Fatal(err)
	}

	app.ErrWriter = &stderr

	err = app.Run(t.Context(), []string{
		"f2_test",
		"-r",
		"{x.cdt.YYYY}/{x.cdt.MM}-{x.cdt.MMM}/{x.cdt.YYYY}-{x.cdt.MM}-{x.cdt.DD}/{%03d}",
		"-R",
		"--target-dir",
		".",
		"--pair",
		"--reset-index-per-dir",
		"-F",
		"--fix-conflicts-pattern",
		"%03d",
		"--sort",
		"time_var",
		"--sort-var",
		"{x.cdt}",
		"--pair-order",
		"dng,jpg",
		"--exclude",
		"golden",
		"testdata",
	})
	if err != nil {
		t.Fatal(err)
	}

	tc := &testutil.TestCase{
		Name: "image pair renaming",
	}

	tc.SnapShot.Stdout = stdout.Bytes()
	tc.SnapShot.Stderr = stderr.Bytes()

	testutil.CompareGoldenFile(t, tc)
}

func TestConditionalSearch(t *testing.T) {
	cases := []testutil.TestCase{
		{
			Name: "use dates to conditionally match files",
			Args: []string{
				"f2_test",
				"-f",
				"{{x.cdt.DD} in [27, 28]}",
				"-r",
				"{f.up}",
				"-f",
				"{{x.cdt.DD} > 27}",
				"-r",
				"{%03d}",
				"-R",
				"--pair",
			},
		},
		{
			Name: "only match files named img33.dng",
			Args: []string{
				"f2_test",
				"-f",
				"{`{xt.FileName}` == `img33.dng`}",
				"-r",
				"{f.up}{ext.up}",
				"-R",
			},
		},
	}

	for _, tc := range cases {
		var stdout bytes.Buffer

		var stdin bytes.Buffer

		var stderr bytes.Buffer

		app, err := f2.New(&stdin, &stdout)
		if err != nil {
			t.Fatal(err)
		}

		app.ErrWriter = &stderr

		err = app.Run(t.Context(), tc.Args)
		if err != nil {
			t.Fatal(err)
		}

		tc.SnapShot.Stdout = stdout.Bytes()
		tc.SnapShot.Stderr = stderr.Bytes()

		testutil.CompareGoldenFile(t, &tc)
	}
}

// Undo must restore a recursive rename that also renamed directories: the
// files are restored inside directories that are restored first.
func TestUndoRecursiveRenameOfDirectories(t *testing.T) {
	root := t.TempDir()
	t.Setenv("TMPDIR", t.TempDir())
	t.Chdir(root)

	for _, path := range []string{"My Dir/My File.txt", "My Dir/Sub Dir/Other File.txt"} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(path, []byte(path), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	run := func(args ...string) {
		t.Helper()

		var stdin, stdout, stderr bytes.Buffer

		app, err := f2.New(&stdin, &stdout)
		if err != nil {
			t.Fatal(err)
		}

		app.ErrWriter = &stderr

		if err := app.Run(t.Context(), append([]string{"f2_test"}, args...)); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, stderr.String())
		}
	}

	run("-f", " ", "-r", "_", "-R", "-d", "-x")
	run("-u", "-x")

	for _, path := range []string{"My Dir/My File.txt", "My Dir/Sub Dir/Other File.txt"} {
		if got, err := os.ReadFile(path); err != nil || string(got) != path {
			t.Errorf("%s not restored: %v", path, err)
		}
	}
}
