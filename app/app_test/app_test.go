package app_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/urfave/cli/v2"

	"github.com/ayoisaiah/f2/app"
	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/testutil"
)

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

	err = renamer.Run(tc.Args)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHelp(t *testing.T) {
	tc := &testutil.TestCase{
		Name: "help",
		Args: []string{"f2_test", "--help"},
	}

	var stdout bytes.Buffer

	renamer, err := app.Get(os.Stdin, &stdout)
	if err != nil {
		t.Fatal(err)
	}

	err = renamer.Run(tc.Args)
	if err != nil {
		t.Fatal(err)
	}

	tc.SnapShot.Stdout = stdout.Bytes()
	testutil.CompareGoldenFile(t, tc)
}

func TestVersion(t *testing.T) {
	tc := &testutil.TestCase{
		Name: "version",
		Args: []string{"f2_test", "--version"},
	}

	var stdout bytes.Buffer

	renamer, err := app.Get(os.Stdin, &stdout)
	if err != nil {
		t.Fatal(err)
	}

	err = renamer.Run(tc.Args)
	if err != nil {
		t.Fatal(err)
	}

	tc.SnapShot.Stdout = stdout.Bytes()
	testutil.CompareGoldenFile(t, tc)
}

func TestDefaultEnv(t *testing.T) {
	cases := []struct {
		Assert      func(t *testing.T, ctx *cli.Context)
		Name        string
		DefaultOpts string
		Args        []string
	}{
		{
			Name:        "enable hidden files",
			Args:        []string{"f2_test", "--find", "jpeg"},
			DefaultOpts: "--hidden",
			Assert: func(t *testing.T, ctx *cli.Context) {
				t.Helper()

				if !ctx.Bool("hidden") {
					t.Fatal("expected --hidden default option to be true")
				}
			},
		},
		{
			Name:        "set a custom --fix-conflicts-pattern",
			Args:        []string{"f2_test", "--find", "jpeg"},
			DefaultOpts: "--fix-conflicts-pattern _%03d",
			Assert: func(t *testing.T, ctx *cli.Context) {
				t.Helper()

				if got := ctx.String("fix-conflicts-pattern"); got != "_%03d" {
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
			Assert: func(t *testing.T, ctx *cli.Context) {
				t.Helper()

				if got := ctx.String("fix-conflicts-pattern"); got != "_%02d" {
					t.Fatalf(
						"expected --fix-conflicts-pattern to default option to be _%%02d, but got: %s",
						got,
					)
				}
			},
		},
		// TODO: Should repeatable options be overridden?
		// {
		// 	Name: "exclude node_modules and git",
		// 	Args: []string{
		// 		"f2_test",
		// 		"--find",
		// 		"jpeg",
		// 		"--exclude-dir",
		// 		".git",
		// 	},
		// 	DefaultOpts: "--exclude-dir node_modules",
		// 	Assert: func(t *testing.T, ctx *cli.Context) {
		// 		want := []string{"node_modules", ".git"}
		// 		if got := ctx.StringSlice("exclude-dir"); !slices.Equal(
		// 			got,
		// 			want,
		// 		) {
		// 			t.Fatalf(
		// 				"expected --exclude-dir to be %v, but got %v",
		// 				want,
		// 				got,
		// 			)
		// 		}
		// 	},
		// },
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Setenv(app.EnvDefaultOpts, tc.DefaultOpts)

			var buf bytes.Buffer

			renamer, err := app.Get(os.Stdin, &buf)
			if err != nil {
				t.Fatal(err)
			}

			err = renamer.Run(tc.Args)
			if err != nil {
				t.Fatal("expected no errors, but got:", err)
			}

			v, exists := renamer.Metadata["ctx"]
			if !exists {
				t.Fatal("default context is not set")
			}

			ctx, ok := v.(*cli.Context)
			if !ok {
				t.Fatal(
					"Unexpected type assertion failure: expected *cli.Context",
				)
			}

			tc.Assert(t, ctx)
		})
	}
}
