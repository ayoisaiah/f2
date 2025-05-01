package variables

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var errInvalidSubmatches = errors.New("Invalid number of submatches")

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

// getIndexingVars retrieves all the index variables in the replacement string
// if any.
func getIndexingVars(replacementInput string) (indexVars, error) {
	var indexMatches indexVars

	submatches := indexVarRegex.FindAllStringSubmatch(
		replacementInput,
		-1,
	)

	if submatches == nil {
		return indexMatches, nil
	}

	expectedLength := 9

	for i, submatch := range submatches {
		if len(submatch) < expectedLength {
			panic(errInvalidSubmatches)
		}

		regex, err := regexp.Compile(submatch[0])
		if err != nil {
			return indexMatches, err
		}

		match := indexVarMatch{
			regex:        regex,
			submatch:     submatch,
			startNumber:  1,
			indexFormat:  submatch[3],
			numberSystem: submatch[5],
		}

		if submatch[1] != "" {
			indexMatches.capturVarIndex = append(indexMatches.capturVarIndex, i)
		}

		if submatch[2] != "" {
			match.startNumber, err = strconv.Atoi(submatch[2])
			if err != nil {
				return indexMatches, err
			}
		}

		if submatch[6] != "" {
			match.step.isSet = true

			match.step.value, err = strconv.Atoi(submatch[6])
			if err != nil {
				return indexMatches, err
			}
		}

		skipNumbers := submatch[7]
		if skipNumbers != "" {
			for val := range strings.SplitSeq(skipNumbers, ";") {
				if strings.Contains(val, "-") {
					numRange := strings.Split(val, "-")

					startNum, err := strconv.Atoi(numRange[0])
					if err != nil {
						return indexMatches, err
					}

					endNum, err := strconv.Atoi(numRange[1])
					if err != nil {
						return indexMatches, err
					}

					match.skip = append(match.skip, numbersToSkip{
						max: int(math.Max(float64(startNum), float64(endNum))),
						min: int(math.Min(float64(startNum), float64(endNum))),
					})

					continue
				}

				num, err := strconv.Atoi(val)
				if err != nil {
					return indexMatches, err
				}

				match.skip = append(match.skip, numbersToSkip{
					max: num,
					min: num,
				})
			}
		}

		if submatch[8] != "" {
			match.isCaptureVar = true
		}

		indexMatches.matches = append(indexMatches.matches, match)
	}

	for range indexMatches.matches {
		indexMatches.offset = append(indexMatches.offset, 0)
	}

	return indexMatches, nil
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

func getExtVars(replacementInput string) (extVars, error) {
	var evMatches extVars

	if !extensionVarRegex.MatchString(replacementInput) {
		return evMatches, nil
	}

	submatches := extensionVarRegex.FindAllStringSubmatch(replacementInput, -1)

	expectedLength := 3

	for _, submatch := range submatches {
		if len(submatch) < expectedLength {
			return evMatches, errInvalidSubmatches
		}

		var match extVarMatch

		regex, err := regexp.Compile(submatch[0])
		if err != nil {
			return evMatches, err
		}

		match.regex = regex

		if submatch[1] != "" {
			match.doubleExt = true
		}

		match.transformToken = submatch[2]

		evMatches.matches = append(evMatches.matches, match)
	}

	return evMatches, nil
}

func getParentDirVars(replacementInput string) (parentDirVars, error) {
	var pvMatches parentDirVars

	if !parentDirVarRegex.MatchString(replacementInput) {
		return pvMatches, nil
	}

	submatches := parentDirVarRegex.FindAllStringSubmatch(replacementInput, -1)

	expectedLength := 3

	for _, submatch := range submatches {
		if len(submatch) < expectedLength {
			return pvMatches, errInvalidSubmatches
		}

		var match parentDirVarMatch

		regex, err := regexp.Compile(submatch[0])
		if err != nil {
			return pvMatches, err
		}

		match.regex = regex
		match.parent = 1

		if submatch[1] != "" {
			match.parent, err = strconv.Atoi(submatch[1])
			if err != nil {
				return pvMatches, err
			}
		}

		match.transformToken = submatch[2]

		pvMatches.matches = append(pvMatches.matches, match)
	}

	return pvMatches, nil
}

func getFilenameVars(replacementInput string) (filenameVars, error) {
	var fvMatches filenameVars

	if !filenameVarRegex.MatchString(replacementInput) {
		return fvMatches, nil
	}

	submatches := filenameVarRegex.FindAllStringSubmatch(replacementInput, -1)

	expectedLength := 2

	for _, submatch := range submatches {
		if len(submatch) < expectedLength {
			return fvMatches, errInvalidSubmatches
		}

		var match filenameVarMatch

		regex, err := regexp.Compile(submatch[0])
		if err != nil {
			return fvMatches, err
		}

		match.regex = regex

		match.transformToken = submatch[1]

		fvMatches.matches = append(fvMatches.matches, match)
	}

	return fvMatches, nil
}

// Extract retrieves all the variables present in the replacement
// string.
func Extract(replacement string) (Variables, error) {
	var vars Variables

	var err error

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
