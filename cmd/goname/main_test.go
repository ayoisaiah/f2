package main

import (
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestFindMatches(t *testing.T) {
	tests := []struct {
		input []string
		regex string
		want  []Change
	}{
		{
			input: []string{"20200112_100034.jpg", "How I Met Your Mother - S09E19.mp4", "Screenshot from 2020-09-12 12-04-15_1000", "How I Met Your Mother - S09E21.mp4"},
			regex: "2020",
			want: []Change{
				Change{source: "20200112_100034.jpg", isDir: false},
				Change{source: "Screenshot from 2020-09-12 12-04-15_1000", isDir: false},
			},
		},
		{
			input: []string{"img1.jpeg", "img2.png", "img3.jpeg", "img4.png", "img5.gif"},
			regex: "jpeg",
			want: []Change{
				Change{source: "img1.jpeg", isDir: false},
				Change{source: "img3.jpeg", isDir: false},
			},
		},
		{
			input: []string{"namecheap-order-3891443.pdf", "mysplits.pdf", "tmux-cheatsheet.pdf", "Blokada-204.conf", "tower heist (2011)-English.srt"},
			regex: "[0-9]+",
			want: []Change{
				Change{source: "namecheap-order-3891443.pdf", isDir: false},
				Change{source: "Blokada-204.conf", isDir: false},
				Change{source: "tower heist (2011)-English.srt", isDir: false},
			},
		},
	}

	for _, tc := range tests {
		op := &Operation{}
		for _, v := range tc.input {
			isDir := strings.HasSuffix(v, "/")
			op.paths = append(op.paths, Change{
				isDir:  isDir,
				source: filepath.Clean(v),
			})
		}

		op.searchRegex = regexp.MustCompile(tc.regex)
		err := op.FindMatches()
		if err != nil {
			t.Errorf("An error occurred while finding matches: %v", err)
		}

		if !reflect.DeepEqual(tc.want, op.matches) {
			t.Fatalf("expected: %v, got: %v", tc.want, op.matches)
		}
	}
}
