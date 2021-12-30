//go:build darwin
// +build darwin

package f2

import (
	"path/filepath"
	"testing"
)

func TestDarwinSpecificConflicts(t *testing.T) {
	testDir := setupFileSystem(t)

	table := []conflictTable{
		{
			name: "File name must not contain : character",
			want: map[conflictType][]Conflict{
				invalidCharacters: {
					{
						source: []string{filepath.Join(testDir, "abc.pdf")},
						target: filepath.Join(testDir, ":::.pdf"),
						cause:  ":",
					},
				},
			},
			args: "-f abc.pdf -r ':::.pdf' " + testDir,
		},
	}

	runConflictCheckHelper(t, table)
}

func TestDarwinFixConflict(t *testing.T) {
	testDir := setupFileSystem(t)

	table := []testCase{
		{
			name: "Fix invalid characters present",
			want: []Change{
				{
					Source:  "abc.pdf",
					BaseDir: testDir,
					Target:  "name.pdf",
				},
			},
			args: "-f abc.pdf -r 'name:::.pdf' -F " + testDir,
		},
	}

	runFixConflictHelper(t, table)
}
