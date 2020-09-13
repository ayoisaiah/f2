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
				{source: "20200112_100034.jpg"},
				{source: "Screenshot from 2020-09-12 12-04-15_1000"},
			},
		},
		{
			input: []string{"img1.jpeg", "img2.png", "img3.jpeg", "img4.png", "img5.gif"},
			regex: "jpeg",
			want: []Change{
				{source: "img1.jpeg"},
				{source: "img3.jpeg"},
			},
		},
		{
			input: []string{"namecheap-order-3891443.pdf", "mysplits.pdf", "tmux-cheatsheet.pdf", "Blokada-204.conf", "tower heist (2011)-English.srt"},
			regex: "[0-9]+",
			want: []Change{
				{source: "namecheap-order-3891443.pdf"},
				{source: "Blokada-204.conf"},
				{source: "tower heist (2011)-English.srt"},
			},
		},
	}

	for i, tc := range tests {
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
			t.Errorf("Test %d — An error occurred while finding matches: %v", i+1, err)
		}

		if !reflect.DeepEqual(tc.want, op.matches) {
			t.Fatalf("Test %d — Expected: %v, got: %v", i+1, tc.want, op.matches)
		}
	}
}

func TestReplace(t *testing.T) {
	tests := []struct {
		matches       []Change
		searchRegex   *regexp.Regexp
		replaceString string
		want          []Change
	}{
		{
			matches: []Change{
				{source: "How I Met Your Mother - S09E21.mp4"},
				{source: "Dear Mother.epub"},
				{source: "Mother - Charlie Puth.mp3"},
			},
			searchRegex:   regexp.MustCompile("Mother"),
			replaceString: "Father",
			want: []Change{
				{source: "How I Met Your Mother - S09E21.mp4", target: "How I Met Your Father - S09E21.mp4"},
				{source: "Dear Mother.epub", target: "Dear Father.epub"},
				{source: "Mother - Charlie Puth.mp3", target: "Father - Charlie Puth.mp3"},
			},
		},
		{
			matches: []Change{
				{source: "pic-1.jpg"},
				{source: "dir/pic-2.jpg"},
				{source: "pic-3/pic-3.jpg"},
				{source: "deep/nested/dir/pic-4.jpg"},
			},
			searchRegex: regexp.MustCompile("pic-"),
			want: []Change{
				{source: "pic-1.jpg", target: "1.jpg"},
				{source: "dir/pic-2.jpg", target: "dir/2.jpg"},
				{source: "pic-3/pic-3.jpg", target: "pic-3/3.jpg"},
				{source: "deep/nested/dir/pic-4.jpg", target: "deep/nested/dir/4.jpg"},
			},
		},
	}

	for i, tc := range tests {
		op := &Operation{}
		op.searchRegex = tc.searchRegex
		op.replaceString = tc.replaceString
		op.matches = tc.matches
		err := op.Replace()
		if err != nil {
			t.Fatalf("Test %d — Error: %v", i+1, err)
		}

		if !reflect.DeepEqual(tc.want, op.matches) {
			t.Fatalf("Test %d — Expected: %v, got: %v", i+1, tc.want, op.matches)
		}
	}
}
