package find_test

import (
	"testing"

	"github.com/ayoisaiah/f2/internal/testutil"
)

var csvCases = []testutil.TestCase{
	{
		Name: "find matches from csv file",
		Want: []string{
			"LICENSE.txt",
			"backup/documents/.hidden_old_resume.txt",
			"projects/project1/index.html",
			"projects/project2/index.html",
			"videos/funny_cats (3).mp4",
		},
		Args: []string{"--csv", "input.csv"},
	},
}

// TestFindCSV tests file matching with CSV files.
func TestFindCSV(t *testing.T) {
	findTest(t, csvCases)
}
