package f2

import (
	"path/filepath"
	"testing"
)

func TestReplaceLongPath(t *testing.T) {
	testDir := setupFileSystem(t)

	longPath := "weirdo/Data Structures and Algorithms/1. Asymptotic Analysis and Insertion Sort, Merge Sort/2.Sorting & Searching why bother with these simple tasks/this is a long path/1. Sorting & Searching- why bother with these simple tasks- - Data Structure & Algorithms - Part-2.mp4"

	dir := filepath.Join(testDir, filepath.Dir(longPath))

	cases := []testCase{
		{
			name: "Overwriting abc.pdf",
			want: []Change{
				{
					BaseDir: dir,
					Source:  "1. Sorting & Searching- why bother with these simple tasks- - Data Structure & Algorithms - Part-2.mp4",
					Target:  "part2.mp4",
				},
			},
			args: "-f '^1\\..*' -r part2.mp4 -R " + testDir,
		},
	}

	runFindReplaceHelper(t, cases)
}
