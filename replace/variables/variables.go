package variables

import (
	"context"
	"regexp"
	"slices"
	"strings"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
)

func greatestCommonDivisor(a, b int) int {
	precision := 0.0001
	if float64(b) < precision {
		return a
	}

	return greatestCommonDivisor(b, a%b)
}

// replaceSlashes replaces forward and backward slashes in the input with an
// underscore character.
func replaceSlashes(input string) string {
	r := strings.NewReplacer("/", "_", "\\", "_")
	return r.Replace(input)
}

// RegexReplace replaces matched substrings in the input with the replacement.
// It respects the specified replacement limit. A negative limit indicates that
// replacement should start from the end of the fileName.
func RegexReplace(
	regex *regexp.Regexp,
	input, replacement string,
	replaceLimit int,
	replaceRange []int,
) string {
	if len(replaceRange) > 0 {
		return replaceByRange(regex, input, replacement, replaceRange)
	}

	return replaceByLimit(regex, input, replacement, replaceLimit)
}

func replaceByRange(
	regex *regexp.Regexp,
	input, replacement string,
	replaceRange []int,
) string {
	matchCount := len(regex.FindAllString(input, -1))
	counter := 1

	output := regex.ReplaceAllStringFunc(input, func(val string) string {
		defer func() { counter++ }()

		if slices.Contains(replaceRange, counter) {
			return replacement
		}

		for _, v := range replaceRange {
			if v < 0 {
				if counter == matchCount+v+1 {
					return replacement
				}
			}
		}

		return val
	})

	return output
}

func replaceByLimit(
	regex *regexp.Regexp,
	input, replacement string,
	limit int,
) string {
	if limit == 0 {
		return regex.ReplaceAllString(input, replacement)
	}

	var output string

	if limit > 0 {
		counter := 0

		output = regex.ReplaceAllStringFunc(input, func(val string) string {
			if counter < limit {
				counter++
				return replacement
			}

			return val
		})
	} else {
		matchCount := len(regex.FindAllString(input, -1))
		limit = matchCount + limit
		counter := 0

		output = regex.ReplaceAllStringFunc(input, func(val string) string {
			if counter >= limit {
				return replacement
			}

			counter++

			return val
		})
	}

	return output
}

// Replace checks if any variables are present in the target filename
// and delegates the variable replacement to the appropriate function.
func Replace(
	ctx context.Context,
	conf *config.Config,
	change *file.Change,
	vars *Variables,
) error {
	providers := []VariableProvider{
		vars.filename,
		vars.ext,
		vars.parentDir,
		vars.date,
		vars.exiftool,
		vars.exif,
		vars.id3,
		vars.csv,
		vars.hash,
		vars.transform,
		&vars.index,
	}

	for _, p := range providers {
		err := p.Replace(ctx, conf, change)
		if err != nil {
			return err
		}
	}

	return nil
}
