package sortfiles_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/osutil"
	"github.com/ayoisaiah/f2/internal/sortfiles"
	"github.com/ayoisaiah/f2/internal/testutil"
)

type sortTestCase struct {
	Name        string
	Unsorted    []string
	Sorted      []string
	Order       []string
	TimeSort    config.Sort
	ReverseSort bool
	SortPerDir  bool
	Revert      bool
}

func sortTest(t *testing.T, unsorted []string) file.Changes {
	t.Helper()

	changes := make(file.Changes, len(unsorted))

	for i := range unsorted {
		v := unsorted[i]

		changes[i] = &file.Change{
			Source:     filepath.Base(v),
			BaseDir:    filepath.Dir(v),
			SourcePath: v,
		}

		f, err := os.Stat(v)
		if err == nil {
			changes[i].IsDir = f.IsDir()
		}
	}

	return changes
}

func TestSortFiles_EnforceHierarchicalOrder(t *testing.T) {
	testCases := []sortTestCase{
		{
			Name: "enforce parent-child directory sorting",
			Unsorted: []string{
				"testdata/dir1/folder/15k.txt",
				"testdata/20k.txt",
				"testdata/dir1/10k.txt",
			},
			Sorted: []string{
				"testdata/20k.txt",
				"testdata/dir1/10k.txt",
				"testdata/dir1/folder/15k.txt",
			},
		},
		{
			Name: "enforce parent-child directory sorting with files and dirs",
			Unsorted: []string{
				"testdata/dir1",
				"testdata/dir1/folder",
				"testdata/dir1/folder/15k.txt",
				"testdata/20k.txt",
				"testdata/dir1/10k.txt",
			},
			Sorted: []string{
				"testdata/dir1",
				"testdata/20k.txt",
				"testdata/dir1/folder",
				"testdata/dir1/10k.txt",
				"testdata/dir1/folder/15k.txt",
			},
		},
		{
			Name: "enforce parent-child directory sorting with multiple files",
			Unsorted: []string{
				"f.txt",
				"dir1/c.txt",
				"dir1/a.txt",
				"e.txt",
			},
			Sorted: []string{
				"f.txt",
				"e.txt",
				"dir1/c.txt",
				"dir1/a.txt",
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.Name, func(t *testing.T) {
			unsorted := sortTest(t, tc.Unsorted)

			sortfiles.EnforceHierarchicalOrder(unsorted)

			testutil.CompareSourcePath(t, tc.Sorted, unsorted)
		})
	}
}

func TestSortFiles_BySize(t *testing.T) {
	testCases := []sortTestCase{
		{
			Name: "sort in ascending order",
			Unsorted: []string{
				"testdata/20k.txt",
				"testdata/4k.txt",
				"testdata/11k.txt",
				"testdata/10k.txt",
			},
			Sorted: []string{
				"testdata/4k.txt",
				"testdata/10k.txt",
				"testdata/11k.txt",
				"testdata/20k.txt",
			},
		},
		{
			Name: "sort in descending order",
			Unsorted: []string{
				"testdata/20k.txt",
				"testdata/4k.txt",
				"testdata/11k.txt",
				"testdata/10k.txt",
			},
			Sorted: []string{
				"testdata/20k.txt",
				"testdata/11k.txt",
				"testdata/10k.txt",
				"testdata/4k.txt",
			},
			ReverseSort: true,
		},
		{
			Name: "sort recursively without --sort-per-dir ",
			Unsorted: []string{
				"testdata/dir1/folder/15k.txt",
				"testdata/20k.txt",
				"testdata/4k.txt",
				"testdata/dir1/folder/3k.txt",
				"testdata/11k.txt",
				"testdata/10k.txt",
				"testdata/dir1/20k.txt",
				"testdata/dir1/10k.txt",
			},
			Sorted: []string{
				"testdata/dir1/folder/3k.txt",
				"testdata/4k.txt",
				"testdata/10k.txt",
				"testdata/dir1/10k.txt",
				"testdata/11k.txt",
				"testdata/dir1/folder/15k.txt",
				"testdata/20k.txt",
				"testdata/dir1/20k.txt",
			},
		},
		{
			Name: "sort recursively in reverse without --sort-per-dir",
			Unsorted: []string{
				"testdata/dir1/folder/15k.txt",
				"testdata/dir1/20k.txt",
				"testdata/20k.txt",
				"testdata/4k.txt",
				"testdata/dir1/folder/3k.txt",
				"testdata/11k.txt",
				"testdata/10k.txt",
				"testdata/dir1/10k.txt",
			},
			Sorted: []string{
				"testdata/dir1/20k.txt",
				"testdata/20k.txt",
				"testdata/dir1/folder/15k.txt",
				"testdata/11k.txt",
				"testdata/10k.txt",
				"testdata/dir1/10k.txt",
				"testdata/4k.txt",
				"testdata/dir1/folder/3k.txt",
			},
			ReverseSort: true,
		},
		{
			Name: "sort recursively with --sort-per-dir",
			Unsorted: []string{
				"testdata/dir1/folder/15k.txt",
				"testdata/20k.txt",
				"testdata/4k.txt",
				"testdata/dir1/folder/3k.txt",
				"testdata/11k.txt",
				"testdata/10k.txt",
				"testdata/dir1/20k.txt",
				"testdata/dir1/10k.txt",
			},
			Sorted: []string{
				"testdata/4k.txt",
				"testdata/10k.txt",
				"testdata/11k.txt",
				"testdata/20k.txt",
				"testdata/dir1/10k.txt",
				"testdata/dir1/20k.txt",
				"testdata/dir1/folder/3k.txt",
				"testdata/dir1/folder/15k.txt",
			},
			SortPerDir: true,
		},
		{
			Name: "sort recursively in reverse with --sort-per-dir",
			Unsorted: []string{
				"testdata/dir1/folder/15k.txt",
				"testdata/20k.txt",
				"testdata/4k.txt",
				"testdata/dir1/folder/3k.txt",
				"testdata/11k.txt",
				"testdata/10k.txt",
				"testdata/dir1/20k.txt",
				"testdata/dir1/10k.txt",
			},
			Sorted: []string{
				"testdata/20k.txt",
				"testdata/11k.txt",
				"testdata/10k.txt",
				"testdata/4k.txt",
				"testdata/dir1/20k.txt",
				"testdata/dir1/10k.txt",
				"testdata/dir1/folder/15k.txt",
				"testdata/dir1/folder/3k.txt",
			},
			SortPerDir:  true,
			ReverseSort: true,
		},
		{
			Name: "sort files without --sort-per-dir",
			Unsorted: []string{
				"testdata/dir1/folder/15k.txt",
				"testdata/20k.txt",
				"testdata/4k.txt",
				"testdata/dir1/folder/3k.txt",
				"testdata/11k.txt",
				"testdata/10k.txt",
				"testdata/dir1/20k.txt",
				"testdata/dir1/10k.txt",
			},
			Sorted: []string{
				"testdata/dir1/folder/3k.txt",
				"testdata/4k.txt",
				"testdata/10k.txt",
				"testdata/dir1/10k.txt",
				"testdata/11k.txt",
				"testdata/dir1/folder/15k.txt",
				"testdata/20k.txt",
				"testdata/dir1/20k.txt",
			},
		},
		{
			Name: "sort files with --sort-per-dir",
			Unsorted: []string{
				"testdata/dir1/folder/15k.txt",
				"testdata/20k.txt",
				"testdata/4k.txt",
				"testdata/dir1/folder/3k.txt",
				"testdata/11k.txt",
				"testdata/10k.txt",
				"testdata/dir1/20k.txt",
				"testdata/dir1/10k.txt",
			},
			Sorted: []string{
				"testdata/4k.txt",
				"testdata/10k.txt",
				"testdata/11k.txt",
				"testdata/20k.txt",
				"testdata/dir1/10k.txt",
				"testdata/dir1/20k.txt",
				"testdata/dir1/folder/3k.txt",
				"testdata/dir1/folder/15k.txt",
			},
			SortPerDir: true,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.Name, func(t *testing.T) {
			unsorted := sortTest(t, tc.Unsorted)

			sortfiles.Changes(
				unsorted,
				config.SortSize,
				tc.ReverseSort,
				tc.SortPerDir,
			)

			testutil.CompareSourcePath(t, tc.Sorted, unsorted)
		})
	}
}

func TestSortFiles_Natural(t *testing.T) {
	testCases := []sortTestCase{
		{
			Name: "sort files numerically",
			Unsorted: []string{
				"file10.txt",
				"file2.txt",
				"file1.txt",
			},
			Sorted: []string{
				"file1.txt",
				"file2.txt",
				"file10.txt",
			},
		},
		{
			Name: "sort files numerically in reverse",
			Unsorted: []string{
				"file1.txt",
				"file10.txt",
				"file2.txt",
			},
			Sorted: []string{
				"file10.txt",
				"file2.txt",
				"file1.txt",
			},
			ReverseSort: true,
		},
		{
			Name: "sort files numerically in reverse",
			Unsorted: []string{
				"01.txt",
				"02.txt",
				"03.txt",
			},
			Sorted: []string{
				"03.txt",
				"02.txt",
				"01.txt",
			},
			ReverseSort: true,
		},
		{
			Name: "sort files with different extensions",
			Unsorted: []string{
				"file1.jpg",
				"file10.txt",
				"file2.png",
			},
			Sorted: []string{
				"file1.jpg",
				"file2.png",
				"file10.txt",
			},
		},
		{
			Name: "sort files with mixed alphanumeric",
			Unsorted: []string{
				"file-2.txt",
				"file10.txt",
				"file-1.txt",
				"file1.txt",
			},
			Sorted: []string{
				"file-1.txt",
				"file-2.txt",
				"file1.txt",
				"file10.txt",
			},
		},
		{
			Name: "sort files with special characters",
			Unsorted: []string{
				"file-2.txt",
				"file1.txt",
				"file_1.txt",
			},
			Sorted: []string{
				"file-2.txt",
				"file1.txt",
				"file_1.txt",
			},
		},
		{
			Name: "sort files with mixed case",
			Unsorted: []string{
				"File10.txt",
				"file2.txt",
				"FILE1.txt",
			},
			Sorted: []string{
				"FILE1.txt",
				"File10.txt",
				"file2.txt",
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.Name, func(t *testing.T) {
			unsorted := sortTest(t, tc.Unsorted)

			sortfiles.Changes(
				unsorted,
				config.SortNatural,
				tc.ReverseSort,
				tc.SortPerDir,
			)

			testutil.CompareSourcePath(t, tc.Sorted, unsorted)
		})
	}
}

func TestSortFiles_ByTime(t *testing.T) {
	testCases := []sortTestCase{
		{
			Name: "sort files by modification time",
			Unsorted: []string{
				"testdata/4k.txt",
				"testdata/10k.txt",
				"testdata/11k.txt",
				"testdata/20k.txt",
				"testdata/dir1/10k.txt",
				"testdata/dir1",
				"testdata/dir1/folder/3k.txt",
				"testdata/dir1/folder/15k.txt",
			},
			Sorted: []string{
				"testdata/11k.txt",
				"testdata/10k.txt",
				"testdata/dir1/10k.txt",
				"testdata/20k.txt",
				"testdata/dir1/folder/3k.txt",
				"testdata/dir1",
				"testdata/4k.txt",
				"testdata/dir1/folder/15k.txt",
			},
			TimeSort: config.SortMtime,
			Order: []string{
				"2025-05-30T06:58:00+01:00",
				"2023-03-30T12:30:00+01:00",
				"2022-05-30T06:58:00+01:00",
				"2023-05-30T12:30:00+01:00",
				"2023-04-30T12:30:00+01:00",
				"2024-06-20T00:29:00+01:00",
				"2024-05-30T06:58:00+01:00",
				"2025-06-20T00:29:00+01:00",
			},
		},
		{
			Name: "sort files by modification time with --sort-per-dir",
			Unsorted: []string{
				"testdata/4k.txt",
				"testdata/10k.txt",
				"testdata/11k.txt",
				"testdata/20k.txt",
				"testdata/dir1/10k.txt",
				"testdata/dir1",
				"testdata/dir1/folder/3k.txt",
				"testdata/dir1/folder/15k.txt",
			},
			Sorted: []string{
				"testdata/10k.txt",
				"testdata/4k.txt",
				"testdata/20k.txt",
				"testdata/11k.txt",
				"testdata/dir1",
				"testdata/dir1/10k.txt",
				"testdata/dir1/folder/3k.txt",
				"testdata/dir1/folder/15k.txt",
			},
			TimeSort: config.SortMtime,
			Order: []string{
				"2023-03-30T12:30:00+01:00",
				"2022-05-30T06:58:00+01:00",
				"2023-05-30T12:30:00+01:00",
				"2023-04-30T12:30:00+01:00",
				"2024-06-20T00:29:00+01:00",
				"2024-05-30T06:58:00+01:00",
				"2025-05-30T06:58:00+01:00",
				"2025-06-20T00:29:00+01:00",
			},
			SortPerDir: true,
		},
		{
			Name: "sort files by modification time in reverse",
			Unsorted: []string{
				"testdata/4k.txt",
				"testdata/10k.txt",
				"testdata/11k.txt",
				"testdata/20k.txt",
			},
			Sorted: []string{
				"testdata/4k.txt",
				"testdata/11k.txt",
				"testdata/20k.txt",
				"testdata/10k.txt",
			},
			TimeSort: config.SortMtime,
			Order: []string{
				"2024-06-20T00:29:00+01:00",
				"2022-05-30T06:58:00+01:00",
				"2024-05-30T06:58:00+01:00",
				"2023-03-30T12:30:00+01:00",
			},
			ReverseSort: true,
		},
		{
			Name: "sort files by access time",
			Unsorted: []string{
				"testdata/4k.txt",
				"testdata/10k.txt",
				"testdata/11k.txt",
				"testdata/20k.txt",
			},
			Sorted: []string{
				"testdata/10k.txt",
				"testdata/20k.txt",
				"testdata/11k.txt",
				"testdata/4k.txt",
			},
			TimeSort: config.SortAtime,
			Order: []string{
				"2024-06-20T00:29:00+01:00",
				"2022-05-30T06:58:00+01:00",
				"2024-05-30T06:58:00+01:00",
				"2023-03-30T12:30:00+01:00",
			},
		},
		{
			Name: "sort files by access time in reverse",
			Unsorted: []string{
				"testdata/4k.txt",
				"testdata/10k.txt",
				"testdata/11k.txt",
				"testdata/20k.txt",
			},
			Sorted: []string{
				"testdata/4k.txt",
				"testdata/11k.txt",
				"testdata/20k.txt",
				"testdata/10k.txt",
			},
			TimeSort: config.SortAtime,
			Order: []string{
				"2024-06-20T00:29:00+01:00",
				"2022-05-30T06:58:00+01:00",
				"2024-05-30T06:58:00+01:00",
				"2023-03-30T12:30:00+01:00",
			},
			ReverseSort: true,
		},
		{
			Name: "sort files by birth time",
			Unsorted: []string{
				"testdata/4.txt",
				"testdata/1.txt",
				"testdata/2.txt",
				"testdata/3.txt",
			},
			Sorted: []string{
				"testdata/1.txt",
				"testdata/2.txt",
				"testdata/3.txt",
				"testdata/4.txt",
			},
			Order: []string{
				"testdata/1.txt",
				"testdata/2.txt",
				"testdata/3.txt",
				"testdata/4.txt",
			},
			TimeSort: config.SortBtime,
		},
		{
			Name: "sort files by birth time in reverse",
			Unsorted: []string{
				"testdata/4.txt",
				"testdata/1.txt",
				"testdata/2.txt",
				"testdata/3.txt",
			},
			Order: []string{
				"testdata/1.txt",
				"testdata/2.txt",
				"testdata/3.txt",
				"testdata/4.txt",
			},
			Sorted: []string{
				"testdata/4.txt",
				"testdata/3.txt",
				"testdata/2.txt",
				"testdata/1.txt",
			},
			TimeSort:    config.SortBtime,
			ReverseSort: true,
		},
		{
			Name: "sort files by change time",
			Unsorted: []string{
				"testdata/4k.txt",
				"testdata/10k.txt",
				"testdata/11k.txt",
				"testdata/20k.txt",
			},
			Sorted: []string{
				"testdata/20k.txt",
				"testdata/4k.txt",
				"testdata/11k.txt",
				"testdata/10k.txt",
			},
			Order: []string{
				"testdata/20k.txt",
				"testdata/4k.txt",
				"testdata/11k.txt",
				"testdata/10k.txt",
			},
			TimeSort: config.SortCtime,
		},
		{
			Name: "sort files by change time in reverse",
			Unsorted: []string{
				"testdata/4k.txt",
				"testdata/10k.txt",
				"testdata/11k.txt",
				"testdata/20k.txt",
			},
			Sorted: []string{
				"testdata/10k.txt",
				"testdata/11k.txt",
				"testdata/4k.txt",
				"testdata/20k.txt",
			},
			Order: []string{
				"testdata/20k.txt",
				"testdata/4k.txt",
				"testdata/11k.txt",
				"testdata/10k.txt",
			},
			TimeSort:    config.SortCtime,
			ReverseSort: true,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		if tc.TimeSort == config.SortAtime || tc.TimeSort == config.SortMtime {
			for i, v := range tc.Unsorted {
				mtime, err := time.Parse(time.RFC3339, tc.Order[i])
				if err != nil {
					t.Fatal(err)
				}

				err = os.Chtimes(v, mtime, mtime)
				if err != nil {
					t.Fatal(err)
				}
			}
		}

		if tc.TimeSort == config.SortBtime {
			for _, v := range tc.Order {
				f, err := os.Create(v)
				if err != nil {
					t.Fatal(err)
				}

				f.Close()

				time.Sleep(10 * time.Millisecond)
			}
		}

		if tc.TimeSort == config.SortCtime {
			if runtime.GOOS == osutil.Windows {
				t.Skip("skip change time test in Windows")
			}

			for _, v := range tc.Order {
				err := os.Chmod(v, 0o755)
				if err != nil {
					t.Fatal(err)
				}

				time.Sleep(10 * time.Millisecond)
			}
		}

		t.Run(tc.Name, func(t *testing.T) {
			unsorted := sortTest(t, tc.Unsorted)

			sortfiles.Changes(
				unsorted,
				tc.TimeSort,
				tc.ReverseSort,
				tc.SortPerDir,
			)

			testutil.CompareSourcePath(t, tc.Sorted, unsorted)

			if tc.TimeSort == config.SortBtime {
				t.Cleanup(func() {
					for _, v := range tc.Order {
						err := os.Remove(v)
						if err != nil {
							t.Log(err)
						}
					}
				})
			}
		})
	}
}

func TestSortFiles_ForRenamingAndUndo(t *testing.T) {
	testCases := []sortTestCase{
		{
			Name: "sort for file renaming",
			Unsorted: []string{
				"testdata/dir1/10k.txt",
				"testdata/dir1",
				"testdata/4k.txt",
				"testdata/dir1/folder/15k.txt",
				"testdata/dir1/folder",
			},
			Sorted: []string{
				"testdata/dir1/folder/15k.txt",
				"testdata/dir1/10k.txt",
				"testdata/dir1/folder",
				"testdata/4k.txt",
				"testdata/dir1",
			},
		},
		{
			Name: "sort for undo",
			Unsorted: []string{
				"testdata/dir1/10k.txt",
				"testdata/dir1",
				"testdata/4k.txt",
				"testdata/dir1/folder/15k.txt",
				"testdata/dir1/folder",
			},
			Sorted: []string{
				"testdata/4k.txt",
				"testdata/dir1",
				"testdata/dir1/10k.txt",
				"testdata/dir1/folder",
				"testdata/dir1/folder/15k.txt",
			},
			Revert: true,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.Name, func(t *testing.T) {
			unsorted := sortTest(t, tc.Unsorted)

			sortfiles.ForRenamingAndUndo(unsorted, tc.Revert)

			testutil.CompareSourcePath(t, tc.Sorted, unsorted)
		})
	}
}

func TestSortFiles_Pairs(t *testing.T) {
	testCases := []sortTestCase{
		{
			Name: "sort file pairs",
			Unsorted: []string{
				"image.dng",
				"a.jpg",
				"image.heif",
				"b.arw",
				"image.jpg",
			},
			Sorted: []string{
				"a.jpg",
				"b.arw",
				"image.dng",
				"image.heif",
				"image.jpg",
			},
		},
		{
			Name: "provide a pair order",
			Unsorted: []string{
				"image.dng",
				"a.jpg",
				"image.heif",
				"b.arw",
				"image.jpg",
			},
			Sorted: []string{
				"a.jpg",
				"b.arw",
				"image.heif",
				"image.dng",
				"image.jpg",
			},
			Order: []string{"heif", "dng", "jpg"},
		},
		{
			Name: "sort multiple file pairs",
			Unsorted: []string{
				"image.dng",
				"a.jpg",
				"image.heif",
				"b.arw",
				"image.jpg",
				"a.pdf",
			},
			Sorted: []string{
				"a.jpg",
				"a.pdf",
				"b.arw",
				"image.jpg",
				"image.dng",
				"image.heif",
			},
			Order: []string{"jpg"},
		},
		{
			Name: "order multiple file pairings",
			Unsorted: []string{
				"image.dng",
				"a.jpg",
				"image.heif",
				"b.arw",
				"image.jpg",
				"a.pdf",
			},
			Sorted: []string{
				"a.pdf",
				"a.jpg",
				"b.arw",
				"image.jpg",
				"image.dng",
				"image.heif",
			},
			Order: []string{"pdf", "jpg"},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.Name, func(t *testing.T) {
			unsorted := sortTest(t, tc.Unsorted)

			sortfiles.Pairs(unsorted, tc.Order)

			testutil.CompareSourcePath(t, tc.Sorted, unsorted)
		})
	}
}
