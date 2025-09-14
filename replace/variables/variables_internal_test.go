package variables

import (
	"regexp"
	"testing"
)

func TestVariableRegex(t *testing.T) {
	testCases := []struct {
		name      string
		regex     *regexp.Regexp
		matches   []string
		unmatches []string
	}{
		{
			name:  "filenameVarRegex",
			regex: filenameVarRegex,
			matches: []string{
				"{f}",
				"{f.up}",
				"{f.lw}",
				"{f.ti}",
				"{f.win}",
				"{f.mac}",
				"{f.di}",
				"{f.norm}",
				"{f.dt}",
				"{f.dt.YYYY}",
			},
			unmatches: []string{
				"f",
				"{f.}",
				"{f.unknown}",
			},
		},
		{
			name:  "extensionVarRegex",
			regex: extensionVarRegex,
			matches: []string{
				"{ext}",
				"{2ext}",
				"{ext.up}",
			},
			unmatches: []string{
				"ext",
				"{ext.}",
				"{3ext}",
			},
		},
		{
			name:  "parentDirVarRegex",
			regex: parentDirVarRegex,
			matches: []string{
				"{p}",
				"{1p}",
				"{10p}",
				"{p.up}",
				"{p.dt.hh}",
			},
			unmatches: []string{
				"p",
				"{p.}",
			},
		},
		{
			name:  "indexVarRegex",
			regex: indexVarRegex,
			matches: []string{
				"{%d}",
				"{$1%d}",
				"{%2d}",
				"{%05d}",
			},
			unmatches: []string{
				"d",
			},
		},
		{
			name:  "hashVarRegex",
			regex: hashVarRegex,
			matches: []string{
				"{hash.sha1}",
				"{hash.sha256.up}",
			},
			unmatches: []string{
				"{hash}",
				"{hash.}",
				"{hash.sha1.}",
			},
		},
		{
			name:  "transformVarRegex",
			regex: transformVarRegex,
			matches: []string{
				"{.up}",
				"{<$1>.up}",
				"{<text>.up}",
			},
			unmatches: []string{
				"{up}",
				"{.}",
			},
		},
		{
			name:  "csvVarRegex",
			regex: csvVarRegex,
			matches: []string{
				"{csv.1}",
				"{csv.10.up}",
			},
			unmatches: []string{
				"{csv}",
				"{csv.}",
				"{csv.1.}",
			},
		},
		{
			name:  "exiftoolVarRegex",
			regex: exiftoolVarRegex,
			matches: []string{
				"{xt.Artist}",
				"{xt.Comment.up}",
			},
			unmatches: []string{
				"{xt}",
				"{xt.}",
				"{xt.Artist.}",
			},
		},
		{
			name:  "id3VarRegex",
			regex: id3VarRegex,
			matches: []string{
				"{id3.title}",
				"{id3.artist.up}",
			},
			unmatches: []string{
				"{id3}",
				"{id3.}",
				"{id3.title.}",
			},
		},
		{
			name:  "exifVarRegex",
			regex: exifVarRegex,
			matches: []string{
				"{exif.iso}",
				"{x.cdt.YYYY}",
			},
			unmatches: []string{
				"{exif}",
				"{exif.}",
				"{exif.iso.}",
			},
		},
		{
			name:  "dateVarRegex",
			regex: dateVarRegex,
			matches: []string{
				"{mtime.YYYY}",
				"{ctime.H}",
				"{btime.DDDD.up}",
				"{atime.MMM}",
				"{mtime}",
				"{ctime}",
				"{btime}",
				"{atime}",
				"{now}",
				"{now.hh.lw}",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, match := range tc.matches {
				if !tc.regex.MatchString(match) {
					t.Errorf(
						"expected %q to match %q",
						match,
						tc.regex.String(),
					)
				}
			}

			for _, unmatch := range tc.unmatches {
				if tc.regex.MatchString(unmatch) {
					t.Errorf(
						"expected %q to not match %q",
						unmatch,
						tc.regex.String(),
					)
				}
			}
		})
	}
}
