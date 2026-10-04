package find_test

import (
	"path/filepath"
	"testing"

	"github.com/ayoisaiah/f2/v2/internal/testutil"
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
	{
		Name: "record renaming steps from csv file",
		Want: []string{
			filepath.Join("testdata", "aa.txt"),
			filepath.Join("testdata", "myname.txt"),
		},
		Args: []string{"--csv", "testdata/input.csv"},
	},
	// TODO: Add more tests
}

// TestFindCSV tests file matching with CSV files.
func TestFindCSV(t *testing.T) {
	findTest(t, csvCases, "")
}

// // TestCSVRename tests that replacement steps are recorded for CSV renames.
// func TestCSVRename(t *testing.T) {
// 	for i := range csvRenameCases {
// 		tc := csvRenameCases[i]
//
// 		t.Run(tc.Name, func(t *testing.T) {
// 			conf := testutil.GetConfig(t, &tc, "")
//
// 			changes, err := find.Find(t.Context(), conf)
// 			if err != nil {
// 				t.Fatal(err)
// 			}
//
// 			changes, err = replace.Replace(t.Context(), conf, changes)
// 			if err != nil {
// 				t.Fatal(err)
// 			}
//
// 			for j, ch := range changes {
// 				if len(ch.Steps) < 2 {
// 					t.Fatalf(
// 						"expected change %d (%s) to have target in steps, got: %v",
// 						j,
// 						ch.Source,
// 						ch.Steps,
// 					)
// 				}
//
// 				if ch.Steps[len(ch.Steps)-1] != tc.Want[j] {
// 					t.Fatalf(
// 						"expected last step to be %q, got %q",
// 						tc.Want[j],
// 						ch.Steps[len(ch.Steps)-1],
// 					)
// 				}
// 			}
// 		})
// 	}
// }
//
