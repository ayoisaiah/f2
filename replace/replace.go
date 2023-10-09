// Package replace substitutes each match according to the configured
// replacement directives which could be plain strings, builtin variables, or
// regex capture variables
package replace

import (
	"errors"
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/pathutil"
	"github.com/ayoisaiah/f2/internal/sortfiles"
	"github.com/ayoisaiah/f2/internal/status"
)

var errInvalidSubmatches = errors.New("Invalid number of submatches")

type numbersToSkip struct {
	min int
	max int
}

type indexVarMatch struct {
	regex  *regexp.Regexp
	index  string
	format string
	skip   []numbersToSkip
	val    []string
	step   struct {
		isSet bool
		value int
	}
	startNumber int
}

type indexVars struct {
	capturVarIndex []int
	matches        []indexVarMatch
}

type transformVarMatch struct {
	regex      *regexp.Regexp
	token      string
	captureVar string
	inputStr   string
	timeStr    string
	val        []string
}

type transformVars struct {
	matches []transformVarMatch
}

type exiftoolVarMatch struct {
	regex          *regexp.Regexp
	attr           string
	transformToken string
	val            []string
}

type exiftoolVars struct {
	matches []exiftoolVarMatch
}

type exifVarMatch struct {
	regex          *regexp.Regexp
	attr           string
	timeStr        string
	transformToken string
	val            []string
}

type exifVars struct {
	matches []exifVarMatch
}

type id3VarMatch struct {
	regex          *regexp.Regexp
	tag            string
	transformToken string
	val            []string
}

type id3Vars struct {
	matches []id3VarMatch
}

type dateVarMatch struct {
	regex          *regexp.Regexp
	attr           string
	token          string
	transformToken string
	val            []string
}

type dateVars struct {
	matches []dateVarMatch
}

type hashVarMatch struct {
	regex          *regexp.Regexp
	hashFn         hashAlgorithm
	transformToken string
	val            []string
}

type hashVars struct {
	matches []hashVarMatch
}

type randomVarMatch struct {
	regex          *regexp.Regexp
	characters     string
	transformToken string
	val            []string
	length         int
}

type randomVars struct {
	matches []randomVarMatch
}

type csvVarMatch struct {
	regex          *regexp.Regexp
	transformToken string
	column         int
}

type csvVars struct {
	submatches [][]string
	values     []csvVarMatch
}

type filenameVarMatch struct {
	regex          *regexp.Regexp
	transformToken string
}

type filenameVars struct {
	matches []filenameVarMatch
}

type extVarMatch struct {
	regex          *regexp.Regexp
	transformToken string
}

type extVars struct {
	matches []extVarMatch
}

type parentDirVarMatch struct {
	regex          *regexp.Regexp
	transformToken string
	parent         int
}

type parentDirVars struct {
	matches []parentDirVarMatch
}

type variables struct {
	exif      exifVars
	exiftool  exiftoolVars
	index     indexVars
	id3       id3Vars
	hash      hashVars
	date      dateVars
	random    randomVars
	transform transformVars
	csv       csvVars
	filename  filenameVars
	ext       extVars
	parentDir parentDirVars
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

	if !indexVarRegex.MatchString(replacementInput) {
		return indexMatches, nil
	}

	submatches := indexVarRegex.FindAllStringSubmatch(
		replacementInput,
		-1,
	)
	expectedLength := 8

	for i, submatch := range submatches {
		if len(submatch) < expectedLength {
			return indexMatches, errInvalidSubmatches
		}

		var match indexVarMatch

		regex, err := regexp.Compile(submatch[0])
		if err != nil {
			return indexMatches, err
		}

		match.regex = regex
		match.val = submatch

		if submatch[1] != "" {
			indexMatches.capturVarIndex = append(indexMatches.capturVarIndex, i)
		}

		if submatch[2] != "" {
			match.startNumber, err = strconv.Atoi(submatch[2])
			if err != nil {
				return indexMatches, err
			}
		} else {
			match.startNumber = 1
		}

		match.index = submatch[3]
		match.format = submatch[5]

		if submatch[6] != "" {
			match.step.isSet = true

			match.step.value, err = strconv.Atoi(submatch[6])
			if err != nil {
				return indexMatches, err
			}
		}

		skipNumbers := submatch[7]
		if skipNumbers != "" {
			numRanges := strings.Split(skipNumbers, ";")
			for _, val := range numRanges {
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

		indexMatches.matches = append(indexMatches.matches, match)
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

// getRandomVars retrieves all the random variables in the
// replacement string if any.
func getRandomVars(replacementInput string) (randomVars, error) {
	var rvMatches randomVars

	if !randomVarRegex.MatchString(replacementInput) {
		return rvMatches, nil
	}

	submatches := randomVarRegex.FindAllStringSubmatch(
		replacementInput,
		-1,
	)
	expectedLength := 5

	for _, submatch := range submatches {
		if len(submatch) < expectedLength {
			return rvMatches, errInvalidSubmatches
		}

		var match randomVarMatch

		match.length = 10
		match.val = submatch

		regex, err := regexp.Compile(submatch[0])
		if err != nil {
			return rvMatches, err
		}

		match.regex = regex

		strLen := submatch[1]
		if strLen != "" {
			match.length, err = strconv.Atoi(strLen)
			if err != nil {
				return rvMatches, err
			}
		}

		match.characters = submatch[2]

		if submatch[3] != "" {
			match.characters = submatch[3]
		}

		match.transformToken = submatch[4]

		rvMatches.matches = append(rvMatches.matches, match)
	}

	return rvMatches, nil
}

func getExtVars(replacementInput string) (extVars, error) {
	var evMatches extVars

	if !extensionVarRegex.MatchString(replacementInput) {
		return evMatches, nil
	}

	submatches := extensionVarRegex.FindAllStringSubmatch(replacementInput, -1)

	expectedLength := 2

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

		match.transformToken = submatch[1]

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

// extractVariables retrieves all the variables present in the replacement
// string.
func extractVariables(replacement string) (variables, error) {
	var vars variables

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

	vars.random, err = getRandomVars(replacement)
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

// regexReplace replaces matched substrings in the input with the replacement.
// It respects the specified replacement limit. A negative limit indicates that
// replacement should start from the end of the fileName.
func regexReplace(
	regex *regexp.Regexp,
	input, replacement string,
	replaceLimit int,
) string {
	var output string

	switch limit := replaceLimit; {
	case limit > 0:
		counter := 0
		output = regex.ReplaceAllStringFunc(
			input,
			func(val string) string {
				if counter == replaceLimit {
					return val
				}

				counter++
				return regex.ReplaceAllString(val, replacement)
			},
		)
	case limit < 0:
		matches := regex.FindAllString(input, -1)

		l := len(matches) + limit
		counter := 0
		output = regex.ReplaceAllStringFunc(
			input,
			func(val string) string {
				if counter >= l {
					return regex.ReplaceAllString(val, replacement)
				}

				counter++
				return val
			},
		)
	default:
		output = regex.ReplaceAllString(input, replacement)
	}

	return output
}

// replaceString replaces all matches in the filename
// with the replacement string.
func replaceString(conf *config.Config, originalName string) string {
	return regexReplace(
		conf.SearchRegex,
		originalName,
		conf.Replacement,
		conf.ReplaceLimit,
	)
}

// replaceMatches handles the replacement of matches in each file with the
// replacement string.
func replaceMatches(
	conf *config.Config,
	matches []*file.Change,
) ([]*file.Change, error) {
	vars, err := extractVariables(conf.Replacement)
	if err != nil {
		return nil, err
	}

	if len(vars.index.matches) > 0 {
		matches = sortfiles.DirectoryHierarchy(matches)
	}

	for i := range matches {
		change := matches[i]
		change.Index = i
		originalName := change.Source
		fileExt := filepath.Ext(originalName)

		if conf.IgnoreExt && !change.IsDir {
			originalName = pathutil.StripExtension(originalName)
		}

		change.Target = replaceString(conf, originalName)

		// Replace any variables present with their corresponding values
		err = replaceVariables(conf, change, &vars)
		if err != nil {
			return nil, err
		}

		// Reattach the original extension to the new file name
		if conf.IgnoreExt && !change.IsDir {
			change.Target += fileExt
		}

		change.Target = strings.TrimSpace(filepath.Clean(change.Target))
		change.Status = status.OK
		change.RelTargetPath = filepath.Join(change.BaseDir, change.Target)
		matches[i] = change
	}

	return matches, nil
}

func handleReplacementChain(
	conf *config.Config,
	matches []*file.Change,
) ([]*file.Change, error) {
	replacementSlice := conf.ReplacementSlice

	for i, v := range replacementSlice {
		config.SetReplacement(v)

		var err error

		matches, err = replaceMatches(conf, matches)
		if err != nil {
			return nil, err
		}

		for j := range matches {
			change := matches[j]

			// Update the source to the target from the previous replacement
			// in preparation for the next replacement
			if i != len(replacementSlice)-1 {
				matches[j].Source = change.Target
			}

			// After the last replacement, update the Source
			// back to the original
			if i > 0 && i == len(replacementSlice)-1 {
				matches[j].Source = change.OriginalSource
			}
		}

		if i != len(replacementSlice)-1 {
			err := conf.SetFindStringRegex(i + 1)
			if err != nil {
				return nil, err
			}
		}
	}

	return matches, nil
}

// Replace applies the file name replacements according to the --replace
// argument.
func Replace(
	conf *config.Config,
	changes []*file.Change,
) ([]*file.Change, error) {
	var err error

	if conf.Sort != "" {
		changes, err = sortfiles.Changes(changes, conf.Sort, conf.ReverseSort)
		if err != nil {
			return nil, err
		}
	}

	changes, err = handleReplacementChain(conf, changes)
	if err != nil {
		return nil, err
	}

	return changes, nil
}
