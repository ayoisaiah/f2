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
}

// TestFindCSV tests file matching with CSV files.
// TODO: Test --csv.
func TestFindCSV(t *testing.T) {
	_ = csvCases

	t.Skip("not implemented")
}
