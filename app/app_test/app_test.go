package app_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"

	"github.com/ayoisaiah/f2/v2/app"
	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/osutil"
	"github.com/ayoisaiah/f2/v2/internal/testutil"
)

// buildBinary builds the package at pkgPath to a temp dir and returns
// the path to the binary.
func buildBinary(t *testing.T, pkgPath string) string {
	t.Helper()

	tmp := t.TempDir()
	bin := filepath.Join(tmp, "f2")

	if runtime.GOOS == osutil.Windows {
		bin = filepath.Join(tmp, "f2.exe")
	}

	cmd := exec.Command("go", "build", "-o", bin, pkgPath)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	return bin
}

func run(
	t *testing.T,
	bin string,
	args ...string,
) (stdout, stderr string, exitCode int) {
	t.Helper()

	var outBuf, errBuf bytes.Buffer

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err == nil {
		return outBuf.String(), errBuf.String(), 0
	}

	var ee *exec.ExitError

	if errors.As(err, &ee) {
		return outBuf.String(), errBuf.String(), ee.ExitCode()
	}

	// If we get here, the process didn't even start or was killed before exec;
	// treat as an infrastructure failure.
	t.Fatalf("failed to run %q: %v (stderr: %s)", bin, err, errBuf.String())

	return "", "", -1 // unreachable
}

func TestHelp(t *testing.T) {
	bin := buildBinary(t, "../../cmd/f2")

	langs := []string{"en", "fr", "es", "de", "ru", "pt", "zh"}

	for _, v := range langs {
		t.Run(v, func(t *testing.T) {
			t.Setenv("LANG", v)

			stdout, stderr, code := run(t, bin, "--help")

			if code != 0 {
				t.Fatalf("expected exit 0, got %d; stderr:\n%s", code, stderr)
			}

			if stdout == "" {
				t.Fatalf("expected help output, got empty stdout")
			}
		})
	}
}

func TestShortHelp_Localized(t *testing.T) {
	bin := buildBinary(t, "../../cmd/f2")

	langs := []string{"en", "fr", "es", "de", "ru", "pt", "zh"}

	for _, v := range langs {
		t.Run(v, func(t *testing.T) {
			t.Setenv("LANG", v)

			stdout, stderr, code := run(t, bin)

			if code != 0 {
				t.Fatalf("expected exit 0, got %d; stderr:\n%s", code, stderr)
			}

			if stdout != "" {
				t.Fatalf("expected empty stdout, but got: %s", stdout)
			}

			if stderr == "" {
				t.Fatalf("expected short help output, got empty stderr")
			}
		})
	}
}

func TestShortHelp(t *testing.T) {
	tc := &testutil.TestCase{
		Name: "short help",
		Args: []string{"f2_test"},
	}

	var stdout bytes.Buffer

	config.Stderr = &stdout

	t.Cleanup(func() {
		config.Stderr = os.Stderr
	})

	renamer, err := app.Get(os.Stdin, os.Stdin)
	if err != nil {
		t.Fatal(err)
	}

	// renamer.Run() calls os.Exit() which causes the test to panic
	// This will recover and make the relevant assertion
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected a panic due to os.Exit(0) but got none")
		}

		tc.SnapShot.Stdout = stdout.Bytes()

		testutil.CompareGoldenFile(t, tc)
	}()

	err = renamer.Run(t.Context(), tc.Args)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVersion(t *testing.T) {
	t.Skip("versioning is no longer hard coded")

	tc := &testutil.TestCase{
		Name: "version",
		Args: []string{"f2_test", "--version"},
	}

	var stdout bytes.Buffer

	renamer, err := app.Get(os.Stdin, &stdout)
	if err != nil {
		t.Fatal(err)
	}

	err = renamer.Run(t.Context(), tc.Args)
	if err != nil {
		t.Fatal(err)
	}

	tc.SnapShot.Stdout = stdout.Bytes()
	testutil.CompareGoldenFile(t, tc)
}

func TestDefaultEnv(t *testing.T) {
	t.Skip("need to update how default env is tested")

	cases := []struct {
		Assert      func(t *testing.T, cmd *cli.Command)
		Name        string
		DefaultOpts string
		Args        []string
	}{
		{
			Name:        "enable hidden files",
			Args:        []string{"f2_test", "--find", "jpeg"},
			DefaultOpts: "--hidden",
			Assert: func(t *testing.T, cmd *cli.Command) {
				t.Helper()

				if !cmd.Bool("hidden") {
					t.Fatal("expected --hidden default option to be true")
				}
			},
		},
		{
			Name:        "set a custom --fix-conflicts-pattern",
			Args:        []string{"f2_test", "--find", "jpeg"},
			DefaultOpts: "--fix-conflicts-pattern _%03d",
			Assert: func(t *testing.T, cmd *cli.Command) {
				t.Helper()

				if got := cmd.String("fix-conflicts-pattern"); got != "_%03d" {
					t.Fatalf(
						"expected --fix-conflicts-pattern to default option to be _%%03d, but got: %s",
						got,
					)
				}
			},
		},
		{
			Name: "override --fix-conflicts-pattern",
			Args: []string{
				"f2_test",
				"--find",
				"jpeg",
				"--fix-conflicts-pattern",
				"_%02d",
			},
			DefaultOpts: "--fix-conflicts-pattern _%03d",
			Assert: func(t *testing.T, cmd *cli.Command) {
				t.Helper()

				if got := cmd.String("fix-conflicts-pattern"); got != "_%02d" {
					t.Fatalf(
						"expected --fix-conflicts-pattern to default option to be _%%02d, but got: %s",
						got,
					)
				}
			},
		},
		// TODO: Should repeatable options be overridden?
		{
			Name: "exclude node_modules and git",
			Args: []string{
				"f2_test",
				"--find",
				"jpeg",
				"--exclude-dir",
				".git",
			},
			DefaultOpts: "--exclude-dir node_modules",
			Assert: func(t *testing.T, cmd *cli.Command) {
				t.Helper()

				want := []string{"node_modules", ".git"}
				if got := cmd.StringSlice("exclude-dir"); !slices.Equal(
					got,
					want,
				) {
					t.Fatalf(
						"expected --exclude-dir to be %v, but got %v",
						want,
						got,
					)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Setenv(app.EnvDefaultOpts, tc.DefaultOpts)

			var buf bytes.Buffer

			renamer, err := app.Get(os.Stdin, &buf)
			if err != nil {
				t.Fatal(err)
			}

			err = renamer.Run(t.Context(), tc.Args)
			if err != nil {
				t.Fatal("expected no errors, but got:", err)
			}

			v, exists := renamer.Metadata["ctx"]
			if !exists {
				t.Fatal("default context is not set")
			}

			cmd, ok := v.(*cli.Command)
			if !ok {
				t.Fatal(
					"Unexpected type assertion failure: expected *cli.Command",
				)
			}

			tc.Assert(t, cmd)
		})
	}
}

func TestStringSliceFlag(t *testing.T) {
	cases := []*testutil.TestCase{
		{
			Name: "commas should not be interpreted as a separator",
			Args: []string{
				"f2_test",
				"--replace",
				"Windows, Linux Episode {%d}{ext}",
			},
			Want: []string{"Windows, Linux Episode {%d}{ext}"},
		},
		{
			Name: "multiple flags should add a separate value to the slice",
			Args: []string{
				"f2_test",
				"--replace",
				"Windows",
				"--replace",
				"Linux Episode {%d}{ext}",
			},
			Want: []string{"Windows", "Linux Episode {%d}{ext}"},
		},
	}

	for _, tc := range cases {
		var stdout bytes.Buffer

		renamer, err := app.Get(os.Stdin, &stdout)
		if err != nil {
			t.Fatal(err)
		}

		renamer.Action = func(_ context.Context, cmd *cli.Command) error {
			assert.Equal(t, tc.Want, cmd.StringSlice("replace"))

			return nil
		}

		_ = renamer.Run(t.Context(), tc.Args)
	}
}
