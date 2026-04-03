package variables

import (
	"context"
	"regexp"
	"strconv"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
)

// Replace inserts the appropriate CSV column
// in the replacement target or an empty string if the column
// is not present in the row.
func (cv csvVars) Replace(
	_ context.Context,
	conf *config.Config,
	change *file.Change,
) error {
	if !csvVarRegex.MatchString(change.Target) {
		return nil
	}

	target := replaceCSVVars(conf, change.Target, change.CSVRow, cv)

	change.Target = target

	return nil
}

// replaceCSVVars inserts the appropriate CSV column
// in the replacement target or an empty string if the column
// is not present in the row.
func replaceCSVVars(
	conf *config.Config,
	target string,
	csvRow []string,
	cv csvVars,
) string {
	for i := range cv.submatches {
		current := cv.values[i]
		column := current.column - 1

		var value string

		if len(csvRow) > column && column >= 0 {
			value = csvRow[column]
		}

		value = transformString(conf, value, current.transformToken)

		target = RegexReplace(current.regex, target, value, 0, nil)
	}

	return target
}

// getCSVVars retrieves all the csv variables in the replacement
// string if any.
func getCSVVars(replacementInput string) (csvVars, error) {
	var csv csvVars
	if csvVarRegex.MatchString(replacementInput) {
		csv.submatches = csvVarRegex.FindAllStringSubmatch(replacementInput, -1)
		expectedLength := 3

		for _, submatch := range csv.submatches {
			if len(submatch) < expectedLength {
				return csv, errInvalidSubmatches
			}

			var match csvVarMatch

			regex, err := regexp.Compile(submatch[0])
			if err != nil {
				return csv, err
			}

			match.regex = regex

			n, err := strconv.Atoi(submatch[1])
			if err != nil {
				return csv, err
			}

			match.column = n
			match.transformToken = submatch[2]

			csv.values = append(csv.values, match)
		}
	}

	return csv, nil
}
