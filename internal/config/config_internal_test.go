package config

import (
	"regexp"
	"testing"
)

func TestRegex(t *testing.T) {
	testCases := []struct {
		name      string
		regex     *regexp.Regexp
		matches   []string
		unmatches []string
	}{
		{
			name:  "sortVarRegex",
			regex: sortVarRegex,
			matches: []string{
				"{sort}",
				"{any_string}",
			},
			unmatches: []string{
				"sort",
				"{sort",
				"sort}",
				"s{or}t}",
				"{}",
			},
		},
		{
			name:  "defaultFixConflictsPatternRegex",
			regex: defaultFixConflictsPatternRegex,
			matches: []string{
				"file(1)",
				"file (1)",
			},
			unmatches: []string{
				"file",
			},
		},
		{
			name:  "customFixConflictsPatternRegex",
			regex: customFixConflictsPatternRegex,
			matches: []string{
				"%d",
				"_%d",
				"__%2d",
				"-%d-",
			},
			unmatches: []string{
				"%x",
			},
		},
		{
			name:  "capturVarIndexRegex",
			regex: capturVarIndexRegex,
			matches: []string{
				"{$1%d}",
				"{$1%2d}",
				"{$1%db}",
				"{$1%do}",
				"{$1%dr}",
				"{$1%dh}",
				"{$1%d-1}",
				"{$1%d<1>}",
				"{$1%d<1-2>}",
				"{$1%d<1-2;3-4>}",
			},
			unmatches: []string{
				"{%d}",
				"{$1}",
				"{$1%d<1-2;3-4;>}",
			},
		},
		{
			name:  "indexVarRegex",
			regex: indexVarRegex,
			matches: []string{
				"{%d}",
				"{1%d}",
				"{$1%d}",
				"{%2d}",
				"{%db}",
				"{%do}",
				"{%dr}",
				"{%dh}",
				"{%d-1}",
				"{%d<1>}",
				"{%d<1-2>}",
				"{%d<1-2;3-4>}",
				"{%d##}",
			},
			unmatches: []string{
				"{d}",
				"{%d<1-2;3-4;>}",
			},
		},
		{
			name:  "findVariableRegex",
			regex: findVariableRegex,
			matches: []string{
				"{find}",
				"{any_string}",
				"{([0-9]{4})-([0-9]{2})}",
			},
			unmatches: []string{
				"find",
				"{find",
				"find}",
				"([0-9]{4})-([0-9]{2})",
				"{}",
			},
		},
		{
			name:  "exifToolVarRegex",
			regex: exifToolVarRegex,
			matches: []string{
				"{xt.Artist}",
				"{xt.Comment}",
			},
			unmatches: []string{
				"{xt.}",
				"{xt}",
			},
		},
		{
			name:  "dateTokenRegex",
			regex: dateTokenRegex,
			matches: []string{
				"{YYYY}",
				"{YY}",
				"{MMMM}",
				"{MMM}",
				"{MM}",
				"{M}",
				"{DDDD}",
				"{DDD}",
				"{DD}",
				"{D}",
				"{H}",
				"{hh}",
				"{h}",
				"{mm}",
				"{m}",
				"{ss}",
				"{s}",
				"{A}",
				"{a}",
				"{unix}",
				"{since}",
			},
			unmatches: []string{
				"{yyyy}",
				"{yy}",
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
