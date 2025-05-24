package f2_test

import (
	"bytes"
	"testing"

	"github.com/ayoisaiah/f2/v2"
	"github.com/ayoisaiah/f2/v2/internal/config"
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

	config.Stderr = &stderr

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
