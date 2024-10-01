//go:build !windows
// +build !windows

package app_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ayoisaiah/f2/app"
)

func simulatePipe(t *testing.T, name string, arg ...string) *exec.Cmd {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(name, arg...)
	cmd.Stdin = r
	cmd.Stdout = w

	oldStdin := os.Stdin

	t.Cleanup(func() {
		os.Stdin = oldStdin
	})

	os.Stdin = r

	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	w.Close()

	return cmd
}

// TODO: Write equivalent for Windows.
func TestPipingInputFromFind(t *testing.T) {
	cases := []struct {
		name     string
		findArgs []string
		expected []string
	}{
		{
			name:     "find all txt files",
			findArgs: []string{"testdata", "-name", "*.txt"},
			expected: []string{
				"testdata/a.txt",
				"testdata/b.txt",
				"testdata/c.txt",
				"testdata/d.txt",
			},
		},
		{
			name:     "find a.txt file",
			findArgs: []string{"testdata", "-name", "a.txt"},
			expected: []string{
				"testdata/a.txt",
			},
		},
		{
			name:     "find a.txt and b.txt files",
			findArgs: []string{"testdata", "-regex", `.*\/\(a\|b\)\.txt$`},
			expected: []string{
				"testdata/a.txt",
				"testdata/b.txt",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			simulatePipe(t, "find", tc.findArgs...)

			_, _ = app.Get(os.Stdin, os.Stdout)

			got := os.Args[len(os.Args)-len(tc.expected):]

			assert.Equal(t, tc.expected, got)
		})
	}
}
