package variables

import (
	"errors"
	"regexp"

	"github.com/ayoisaiah/f2/v2/internal/localize"
)

var errInvalidSubmatches = errors.New(localize.T("error.invalid_submatches"))

// getDateVars retrieves all the date variables in the replacement
// string if any.
func getDateVars(replacementInput string) (dateVars, error) {
	var dateVarMatches dateVars

	if !dateVarRegex.MatchString(replacementInput) {
		return dateVarMatches, nil
	}

	submatches := dateVarRegex.FindAllStringSubmatch(
		replacementInput,
		-1,
	)
	expectedLength := 4

	for _, submatch := range submatches {
		if len(submatch) < expectedLength {
			return dateVarMatches, errInvalidSubmatches
		}

		var match dateVarMatch

		regex, err := regexp.Compile(submatch[0])
		if err != nil {
			return dateVarMatches, err
		}

		match.regex = regex
		match.val = submatch
		match.attr = submatch[1]
		match.token = submatch[2]
		match.transformToken = submatch[3]

		dateVarMatches.matches = append(dateVarMatches.matches, match)
	}

	return dateVarMatches, nil
}

// getHashVars retrieves all the hash variables in the replacement
// string if any.
func getHashVars(replacementInput string) (hashVars, error) {
	var hashMatches hashVars

	if !hashVarRegex.MatchString(replacementInput) {
		return hashMatches, nil
	}

	submatches := hashVarRegex.FindAllStringSubmatch(
		replacementInput,
		-1,
	)
	expectedLength := 3

	for _, submatch := range submatches {
		if len(submatch) < expectedLength {
			return hashMatches, errInvalidSubmatches
		}

		var match hashVarMatch

		regex, err := regexp.Compile(submatch[0])
		if err != nil {
			return hashMatches, err
		}

		match.regex = regex
		match.val = submatch
		match.hashFn = hashAlgorithm(submatch[1])
		match.transformToken = submatch[2]

		hashMatches.matches = append(hashMatches.matches, match)
	}

	return hashMatches, nil
}

// getTransformVars retrieves all the string transformation variables
// in the replacement string if any.
func getTransformVars(replacementInput string) (transformVars, error) {
	var transformVarMatches transformVars

	if !transformVarRegex.MatchString(replacementInput) {
		return transformVarMatches, nil
	}

	submatches := transformVarRegex.FindAllStringSubmatch(
		replacementInput,
		-1,
	)
	expectedLength := 5

	for _, submatch := range submatches {
		if len(submatch) < expectedLength {
			return transformVarMatches, errInvalidSubmatches
		}

		var match transformVarMatch

		regex, err := regexp.Compile(submatch[0])
		if err != nil {
			return transformVarMatches, err
		}

		match.regex = regex
		match.val = submatch
		match.captureVar = submatch[1]
		match.inputStr = submatch[2]
		match.token = submatch[3]
		match.timeStr = submatch[4]

		transformVarMatches.matches = append(transformVarMatches.matches, match)
	}

	return transformVarMatches, nil
}

// getExifVars retrieves all the exif variables in the replacement
// string if any.
func getExifVars(replacementInput string) (exifVars, error) {
	var exifMatches exifVars

	if !exifVarRegex.MatchString(replacementInput) {
		return exifMatches, nil
	}

	submatches := exifVarRegex.FindAllStringSubmatch(
		replacementInput,
		-1,
	)
	expectedLength := 4

	for _, submatch := range submatches {
		if len(submatch) < expectedLength {
			return exifMatches, errInvalidSubmatches
		}

		var match exifVarMatch

		regex, err := regexp.Compile(submatch[0])
		if err != nil {
			return exifMatches, err
		}

		match.regex = regex
		match.val = submatch

		match.attr = submatch[1]
		if submatch[2] != "" {
			match.attr = submatch[2]
		}

		match.timeStr = submatch[3]

		match.transformToken = submatch[4]

		exifMatches.matches = append(exifMatches.matches, match)
	}

	return exifMatches, nil
}

// getExifToolVars retrieves all the exiftool variables in the
// replacement string if any.
func getExifToolVars(replacementInput string) (exiftoolVars, error) {
	var exiftoolMatches exiftoolVars

	if !exiftoolVarRegex.MatchString(replacementInput) {
		return exiftoolMatches, nil
	}

	submatches := exiftoolVarRegex.FindAllStringSubmatch(
		replacementInput,
		-1,
	)
	expectedLength := 3

	for _, submatch := range submatches {
		if len(submatch) < expectedLength {
			return exiftoolMatches, errInvalidSubmatches
		}

		var match exiftoolVarMatch

		regex, err := regexp.Compile(submatch[0])
		if err != nil {
			return exiftoolMatches, err
		}

		match.regex = regex
		match.attr = submatch[1]
		match.val = submatch
		match.transformToken = submatch[2]

		exiftoolMatches.matches = append(exiftoolMatches.matches, match)
	}

	return exiftoolMatches, nil
}

// getID3Vars retrieves all the id3 variables in the
// replacement string if any.
func getID3Vars(replacementInput string) (id3Vars, error) {
	var id3Matches id3Vars

	if !id3VarRegex.MatchString(replacementInput) {
		return id3Matches, nil
	}

	submatches := id3VarRegex.FindAllStringSubmatch(
		replacementInput,
		-1,
	)
	expectedLength := 2

	for _, submatch := range submatches {
		if len(submatch) < expectedLength {
			return id3Matches, errInvalidSubmatches
		}

		var match id3VarMatch

		regex, err := regexp.Compile(submatch[0])
		if err != nil {
			return id3Matches, err
		}

		match.regex = regex
		match.tag = submatch[1]
		match.transformToken = submatch[2]
		match.val = submatch

		id3Matches.matches = append(id3Matches.matches, match)
	}

	return id3Matches, nil
}

// Extract retrieves all the variables present in the replacement string.
func Extract(replacement string) (vars Variables, err error) {
	if replacement == "" {
		return
	}

	vars.filename, err = getFilenameVars(replacement)
	if err != nil {
		return vars, err
	}

	vars.ext, err = getExtVars(replacement)
	if err != nil {
		return vars, err
	}

	vars.parentDir, err = getParentDirVars(replacement)
	if err != nil {
		return vars, err
	}

	vars.exif, err = getExifVars(replacement)
	if err != nil {
		return vars, err
	}

	vars.index, err = getIndexingVars(replacement)
	if err != nil {
		return vars, err
	}

	vars.id3, err = getID3Vars(replacement)
	if err != nil {
		return vars, err
	}

	vars.hash, err = getHashVars(replacement)
	if err != nil {
		return vars, err
	}

	vars.date, err = getDateVars(replacement)
	if err != nil {
		return vars, err
	}

	vars.exiftool, err = getExifToolVars(replacement)
	if err != nil {
		return vars, err
	}

	vars.transform, err = getTransformVars(replacement)
	if err != nil {
		return vars, err
	}

	vars.csv, err = getCSVVars(replacement)
	if err != nil {
		return vars, err
	}

	return vars, nil
}
