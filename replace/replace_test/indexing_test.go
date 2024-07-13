package replace_test

import (
	"testing"

	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/testutil"
)

func TestIndexing(t *testing.T) {
	testCases := []testutil.TestCase{
		{
			Name: "replace with auto incrementing integers",
			Changes: []*file.Change{
				{
					Source: "a.txt",
				},
				{
					Source: "b.txt",
				},
				{
					Source: "c.txt",
				},
			},
			Want: []string{"1.txt", "2.txt", "3.txt"},
			Args: []string{"-f", "a|b|c", "-r", "{%d}"},
		},
		{
			Name: "replace with multiple incrementing integers",
			Changes: []*file.Change{
				{
					Source: "a.txt",
				},
				{
					Source: "b.txt",
				},
				{
					Source: "c.txt",
				},
			},
			Want: []string{"1_10_0100.txt", "2_20_0200.txt", "3_30_0300.txt"},
			Args: []string{"-f", "a|b|c", "-r", "{%d}_{10%02d10}_{100%04d100}"},
		},
		{
			Name: "replace with non-arabic numerals",
			Changes: []*file.Change{
				{
					Source: "a.txt",
				},
				{
					Source: "b.txt",
				},
				{
					Source: "c.txt",
				},
			},
			Want: []string{"I_1 1_1.txt", "II_2 2_10.txt", "III_3 3_11.txt"},
			Args: []string{"-f", "a|b|c", "-r", "{%dr}_{%do} {%dh}_{%db}"},
		},
		{
			Name: "skip some numbers when incrementing",
			Changes: []*file.Change{
				{
					Source: "a.txt",
				},
				{
					Source: "b.txt",
				},
				{
					Source: "c.txt",
				},
			},
			Want: []string{"16.txt", "17.txt", "18.txt"},
			Args: []string{"-f", "a|b|c", "-r", "{10%d<10-15>}"},
		},
		{
			Name: "use integer capture variables",
			Changes: []*file.Change{
				{
					Source: "doc1.txt",
				},
				{
					Source: "doc4.txt",
				},
				{
					Source: "doc99.txt",
				},
			},
			Want: []string{"001.txt", "004.txt", "099.txt"},
			Args: []string{"-f", "doc(\\d+)", "-r", "{$1%03d}"},
		},
		{
			Name: "use integer capture variables with explicit step",
			Changes: []*file.Change{
				{
					Source: "doc1.txt",
				},
				{
					Source: "doc4.txt",
				},
				{
					Source: "doc99.txt",
				},
			},
			Want: []string{"006.txt", "009.txt", "104.txt"},
			Args: []string{"-f", "doc(\\d+)", "-r", "{$1%03d5}"},
		},
		{
			Name: "skip some numbers while indexing with capture variables",
			Changes: []*file.Change{
				{
					Source: "doc1.txt",
				},
				{
					Source: "doc4.txt",
				},
				{
					Source: "doc99.txt",
				},
			},
			Want: []string{"002.txt", "005.txt", "099.txt"},
			Args: []string{"-f", "doc(\\d+)", "-r", "{$1%03d<1;4>}"},
		},
		{
			Name: "reset index per directory",
			Changes: []*file.Change{
				{
					BaseDir: "folder1",
					Source:  "f1.log",
				},
				{
					BaseDir: "folder1",
					Source:  "f2.log",
				},
				{
					BaseDir: "folder2",
					Source:  "f3.log",
				},
				{
					BaseDir: "folder2",
					Source:  "f4.log",
				},
				{
					BaseDir: "folder3",
					Source:  "f5.log",
				},
				{
					BaseDir: "folder3",
					Source:  "f6.log",
				},
			},
			Want: []string{
				"folder1/f1_001.log",
				"folder1/f2_002.log",
				"folder2/f3_001.log",
				"folder2/f4_002.log",
				"folder3/f5_001.log",
				"folder3/f6_002.log",
			},
			Args: []string{
				"-f",
				".*",
				"-r",
				"{f}_{%03d}{ext}",
				"--reset-index-per-dir",
			},
		},
	}

	replaceTest(t, testCases)
}
