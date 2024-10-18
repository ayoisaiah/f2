package find_test

import (
	"testing"

	"github.com/ayoisaiah/f2/internal/testutil"
)

var csvCases = []testutil.TestCase{
	{
		Name: "find matches from csv file",
		Want: []string{
			"a.txt",
			"c.txt",
		},
		Args: []string{"--csv", "testdata/input.csv"},
	},
	// TODO: Add more tests
}

// TestFindCSV tests file matching with CSV files.
func TestFindCSV(t *testing.T) {
	findTest(t, csvCases, "")
}
