package f2

import "testing"

func TestSortBySize(t *testing.T) {
	testDir := "../testdata/images"

	cases := []testCase{
		{
			name: "Sort files by size in descending order",
			want: []Change{
				{
					Source:  "tractor-raw.cr2",
					BaseDir: testDir,
					Target:  "001.cr2",
				},
				{
					Source:  "proraw.dng",
					BaseDir: testDir,
					Target:  "002.dng",
				},
				{
					Source:  "bike.jpeg",
					BaseDir: testDir,
					Target:  "003.jpeg",
				},
				{
					Source:  "tractor-raw.json",
					BaseDir: testDir,
					Target:  "004.json",
				},
				{
					Source:  "proraw.json",
					BaseDir: testDir,
					Target:  "005.json",
				},
				{
					Source:  "bike.json",
					BaseDir: testDir,
					Target:  "006.json",
				},
			},
			args: []string{
				"-f",
				".*",
				"-r",
				"%03d",
				"-e",
				"-sort",
				"size",
				testDir,
			},
		},
		{
			name: "Sort files by size in ascending order",
			want: []Change{
				{
					Source:  "tractor-raw.cr2",
					BaseDir: testDir,
					Target:  "006.cr2",
				},
				{
					Source:  "proraw.dng",
					BaseDir: testDir,
					Target:  "005.dng",
				},
				{
					Source:  "bike.jpeg",
					BaseDir: testDir,
					Target:  "004.jpeg",
				},
				{
					Source:  "tractor-raw.json",
					BaseDir: testDir,
					Target:  "003.json",
				},
				{
					Source:  "proraw.json",
					BaseDir: testDir,
					Target:  "002.json",
				},
				{
					Source:  "bike.json",
					BaseDir: testDir,
					Target:  "001.json",
				},
			},
			args: []string{
				"-f",
				".*",
				"-r",
				"%03d",
				"-e",
				"-sortr",
				"size",
				testDir,
			},
		},
	}

	runFindReplace(t, cases)
}

func TestDefaultSort(t *testing.T) {
	testDir := "../testdata/images"

	cases := []testCase{
		{
			name: "Sort files alphabetically in a descending order",
			want: []Change{
				{
					Source:  "tractor-raw.cr2",
					BaseDir: testDir,
					Target:  "005.cr2",
				},
				{
					Source:  "proraw.dng",
					BaseDir: testDir,
					Target:  "003.dng",
				},
				{
					Source:  "bike.jpeg",
					BaseDir: testDir,
					Target:  "001.jpeg",
				},
				{
					Source:  "tractor-raw.json",
					BaseDir: testDir,
					Target:  "006.json",
				},
				{
					Source:  "proraw.json",
					BaseDir: testDir,
					Target:  "004.json",
				},
				{
					Source:  "bike.json",
					BaseDir: testDir,
					Target:  "002.json",
				},
			},
			args: []string{
				"-f",
				".*",
				"-r",
				"%03d",
				"-e",
				"-sort",
				"default",
				testDir,
			},
		},
		{
			name: "Sort files alphabetically in a descending order",
			want: []Change{
				{
					Source:  "tractor-raw.json",
					BaseDir: testDir,
					Target:  "001.json",
				},
				{
					Source:  "tractor-raw.cr2",
					BaseDir: testDir,
					Target:  "002.cr2",
				},
				{
					Source:  "proraw.json",
					BaseDir: testDir,
					Target:  "003.json",
				},
				{
					Source:  "proraw.dng",
					BaseDir: testDir,
					Target:  "004.dng",
				},
				{
					Source:  "bike.json",
					BaseDir: testDir,
					Target:  "005.json",
				},
				{
					Source:  "bike.jpeg",
					BaseDir: testDir,
					Target:  "006.jpeg",
				},
			},
			args: []string{
				"-f",
				".*",
				"-r",
				"%03d",
				"-e",
				"-sortr",
				"default",
				testDir,
			},
		},
	}

	runFindReplace(t, cases)
}
