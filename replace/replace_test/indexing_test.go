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
	}

	replaceTest(t, testCases)
}
