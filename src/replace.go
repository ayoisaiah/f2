package f2

import (
	"errors"
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type numbersToSkip struct {
	min int
	max int
}

type numberVarVal struct {
	regex  *regexp.Regexp
	index  string
	format string
	skip   []numbersToSkip
	step   struct {
		isSet bool
		value int
	}
	startNumber int
}

type numberVar struct {
	capturVarIndex []int
	submatches     [][]string
	values         []numberVarVal
}

type transformVar struct {
	submatches [][]string
	values     []struct {
		regex *regexp.Regexp
		token string
	}
}

type exiftoolVar struct {
	submatches [][]string
	values     []struct {
		regex *regexp.Regexp
		attr  string
	}
}

type exifVar struct {
	submatches [][]string
	values     []struct {
		regex   *regexp.Regexp
		attr    string
		timeStr string
	}
}

type id3Var struct {
	submatches [][]string
	values     []struct {
		regex *regexp.Regexp
		tag   string
	}
}

type dateVar struct {
	submatches [][]string
	values     []struct {
		regex *regexp.Regexp
		attr  string
		token string
	}
}

type hashVar struct {
	submatches [][]string
	values     []struct {
		regex  *regexp.Regexp
		hashFn hashAlgorithm
	}
}

type randomVar struct {
	submatches [][]string
	values     []struct {
		regex      *regexp.Regexp
		characters string
		length     int
	}
}

type csvVar struct {
	submatches [][]string
	values     []struct {
		regex  *regexp.Regexp
		column int
	}
}

type variables struct {
	exif      exifVar
	exiftool  exiftoolVar
	number    numberVar
	id3       id3Var
	hash      hashVar
	date      dateVar
	random    randomVar
	transform transformVar
	csv       csvVar
}

var (
	errInvalidSubmatches = errors.New("Invalid number of submatches")
)

// getCsvVar retrieves all the csv variables in the replacement
// string if any.
func getCsvVar(replacementInput string) (csvVar, error) {
	var c csvVar
	if csvRegex.MatchString(replacementInput) {
		c.submatches = csvRegex.FindAllStringSubmatch(replacementInput, -1)
		expectedLength := 2

		for _, submatch := range c.submatches {
			if len(submatch) < expectedLength {
				return c, errInvalidSubmatches
			}

			var x struct {
				regex  *regexp.Regexp
				column int
			}

			regex, err := regexp.Compile(submatch[0])
			if err != nil {
				return c, err
			}

			x.regex = regex

			n, err := strconv.Atoi(submatch[1])
			if err != nil {
				return c, err
			}

			x.column = n
			c.values = append(c.values, x)
		}
	}

	return c, nil
}

// getDateVar retrieves all the date variables in the replacement
// string if any.
func getDateVar(replacementInput string) (dateVar, error) {
	var d dateVar
	if dateRegex.MatchString(replacementInput) {
		d.submatches = dateRegex.FindAllStringSubmatch(replacementInput, -1)
		expectedLength := 3

		for _, submatch := range d.submatches {
			if len(submatch) < expectedLength {
				return d, errInvalidSubmatches
			}

			var x struct {
				regex *regexp.Regexp
				attr  string
				token string
			}

			regex, err := regexp.Compile(submatch[0])
			if err != nil {
				return d, err
			}

			x.regex = regex
			x.attr = submatch[1]
			x.token = submatch[2]

			d.values = append(d.values, x)
		}
	}

	return d, nil
}

// getHashVar retrieves all the hash variables in the replacement
// string if any.
func getHashVar(replacementInput string) (hashVar, error) {
	var h hashVar
	if hashRegex.MatchString(replacementInput) {
		h.submatches = hashRegex.FindAllStringSubmatch(replacementInput, -1)
		expectedLength := 2

		for _, submatch := range h.submatches {
			if len(submatch) < expectedLength {
				return h, errInvalidSubmatches
			}

			var x struct {
				regex  *regexp.Regexp
				hashFn hashAlgorithm
			}

			regex, err := regexp.Compile(submatch[0])
			if err != nil {
				return h, err
			}

			x.regex = regex
			x.hashFn = hashAlgorithm(submatch[1])
			h.values = append(h.values, x)
		}
	}

	return h, nil
}

// getTransformVar retrieves all the string transformation variables
// in the replacement string if any.
func getTransformVar(replacementInput string) (transformVar, error) {
	var t transformVar
	if transformRegex.MatchString(replacementInput) {
		t.submatches = transformRegex.FindAllStringSubmatch(
			replacementInput,
			-1,
		)
		expectedLength := 2

		for _, submatch := range t.submatches {
			if len(submatch) < expectedLength {
				return t, errInvalidSubmatches
			}

			var x struct {
				regex *regexp.Regexp
				token string
			}

			regex, err := regexp.Compile(submatch[0])
			if err != nil {
				return t, err
			}

			x.regex = regex
			x.token = submatch[1]
			t.values = append(t.values, x)
		}
	}

	return t, nil
}

// getExifVar retrieves all the exif variables in the replacement
// string if any.
func getExifVar(replacementInput string) (exifVar, error) {
	var ex exifVar

	if exifRegex.MatchString(replacementInput) {
		ex.submatches = exifRegex.FindAllStringSubmatch(replacementInput, -1)
		expectedLength := 3

		for _, submatch := range ex.submatches {
			if len(submatch) < expectedLength {
				return ex, errInvalidSubmatches
			}

			var val struct {
				regex   *regexp.Regexp
				attr    string
				timeStr string
			}

			regex, err := regexp.Compile(submatch[0])
			if err != nil {
				return ex, err
			}

			val.regex = regex

			if strings.Contains(submatch[0], "exif.dt") ||
				strings.Contains(submatch[0], "x.dt") {
				submatch = append(submatch[:1], submatch[1+1:]...)
			}

			val.attr = submatch[1]
			if val.attr == "dt" {
				val.timeStr = submatch[2]
			}

			ex.values = append(ex.values, val)
		}
	}

	return ex, nil
}

// getNumberVar retrieves all the index variables in the replacement string
// if any.
func getNumberVar(replacementInput string) (numberVar, error) {
	var nv numberVar

	if indexRegex.MatchString(replacementInput) {
		nv.submatches = indexRegex.FindAllStringSubmatch(replacementInput, -1)
		expectedLength := 8

		for i, submatch := range nv.submatches {
			if len(submatch) < expectedLength {
				return nv, errInvalidSubmatches
			}

			var val numberVarVal

			regex, err := regexp.Compile(submatch[0])
			if err != nil {
				return nv, err
			}

			val.regex = regex

			if submatch[1] != "" {
				nv.capturVarIndex = append(nv.capturVarIndex, i)
			}

			if submatch[2] != "" {
				val.startNumber, err = strconv.Atoi(submatch[2])
				if err != nil {
					return nv, err
				}
			} else {
				val.startNumber = 1
			}

			val.index = submatch[3]
			val.format = submatch[5]

			if submatch[6] != "" {
				val.step.isSet = true

				val.step.value, err = strconv.Atoi(submatch[6])
				if err != nil {
					return nv, err
				}
			}

			skipNumbers := submatch[7]
			if skipNumbers != "" {
				slice := strings.Split(skipNumbers, ";")
				for _, v := range slice {
					if strings.Contains(v, "-") {
						sl := strings.Split(v, "-")

						n1, err := strconv.Atoi(sl[0])
						if err != nil {
							return nv, err
						}

						n2, err := strconv.Atoi(sl[1])
						if err != nil {
							return nv, err
						}

						val.skip = append(val.skip, numbersToSkip{
							max: int(math.Max(float64(n1), float64(n2))),
							min: int(math.Min(float64(n1), float64(n2))),
						})

						continue
					}

					num, err := strconv.Atoi(v)
					if err != nil {
						return nv, err
					}

					val.skip = append(val.skip, numbersToSkip{
						max: num,
						min: num,
					})
				}
			}

			nv.values = append(nv.values, val)
		}
	}

	return nv, nil
}

// getExifToolVar retrieves all the exiftool variables in the
// replacement string if any.
func getExifToolVar(replacementInput string) (exiftoolVar, error) {
	var et exiftoolVar
	if exiftoolRegex.MatchString(replacementInput) {
		et.submatches = exiftoolRegex.FindAllStringSubmatch(
			replacementInput,
			-1,
		)
		expectedLength := 2

		for _, submatch := range et.submatches {
			if len(submatch) < expectedLength {
				return et, errInvalidSubmatches
			}

			var x struct {
				regex *regexp.Regexp
				attr  string
			}

			regex, err := regexp.Compile(submatch[0])
			if err != nil {
				return et, err
			}

			x.regex = regex
			x.attr = submatch[1]

			et.values = append(et.values, x)
		}
	}

	return et, nil
}

// getID3Var retrieves all the id3 variables in the
// replacement string if any.
func getID3Var(replacementInput string) (id3Var, error) {
	var iv id3Var
	if id3Regex.MatchString(replacementInput) {
		iv.submatches = id3Regex.FindAllStringSubmatch(replacementInput, -1)
		expectedLength := 2

		for _, submatch := range iv.submatches {
			if len(submatch) < expectedLength {
				return iv, errInvalidSubmatches
			}

			var x struct {
				regex *regexp.Regexp
				tag   string
			}

			regex, err := regexp.Compile(submatch[0])
			if err != nil {
				return iv, err
			}

			x.regex = regex
			x.tag = submatch[1]

			iv.values = append(iv.values, x)
		}
	}

	return iv, nil
}

// getRandomVar retrieves all the random variables in the
// replacement string if any.
func getRandomVar(replacementInput string) (randomVar, error) {
	var rv randomVar

	if randomRegex.MatchString(replacementInput) {
		rv.submatches = randomRegex.FindAllStringSubmatch(replacementInput, -1)
		expectedLength := 4

		for _, submatch := range rv.submatches {
			if len(submatch) < expectedLength {
				return rv, errInvalidSubmatches
			}

			var val struct {
				regex      *regexp.Regexp
				characters string
				length     int
			}

			val.length = 10

			regex, err := regexp.Compile(submatch[0])
			if err != nil {
				return rv, err
			}

			val.regex = regex

			strLen := submatch[1]
			if strLen != "" {
				val.length, err = strconv.Atoi(strLen)
				if err != nil {
					return rv, err
				}
			}

			val.characters = submatch[2]

			if submatch[3] != "" {
				val.characters = submatch[3]
			}

			rv.values = append(rv.values, val)
		}
	}

	return rv, nil
}

// extractVariables retrieves all the variables present in the replacement
// string.
func extractVariables(replacementInput string) (variables, error) {
	var v variables

	var err error

	v.exif, err = getExifVar(replacementInput)
	if err != nil {
		return v, err
	}

	v.number, err = getNumberVar(replacementInput)
	if err != nil {
		return v, err
	}

	v.id3, err = getID3Var(replacementInput)
	if err != nil {
		return v, err
	}

	v.hash, err = getHashVar(replacementInput)
	if err != nil {
		return v, err
	}

	v.date, err = getDateVar(replacementInput)
	if err != nil {
		return v, err
	}

	v.random, err = getRandomVar(replacementInput)
	if err != nil {
		return v, err
	}

	v.exiftool, err = getExifToolVar(replacementInput)
	if err != nil {
		return v, err
	}

	v.transform, err = getTransformVar(replacementInput)
	if err != nil {
		return v, err
	}

	v.csv, err = getCsvVar(replacementInput)
	if err != nil {
		return v, err
	}

	return v, nil
}

// regexReplace replaces matched substrings in the input with the replacement.
// It respects the specified replacement limit. A negative limit indicates that
// replacement should start from the end of the fileName.
func regexReplace(
	r *regexp.Regexp,
	input, replacement string,
	replaceLimit int,
) string {
	var output string

	switch limit := replaceLimit; {
	case limit > 0:
		counter := 0
		output = r.ReplaceAllStringFunc(
			input,
			func(val string) string {
				if counter == replaceLimit {
					return val
				}

				counter++
				return r.ReplaceAllString(val, replacement)
			},
		)
	case limit < 0:
		matches := r.FindAllString(input, -1)

		l := len(matches) + limit
		counter := 0
		output = r.ReplaceAllStringFunc(
			input,
			func(val string) string {
				if counter >= l {
					return r.ReplaceAllString(val, replacement)
				}

				counter++
				return val
			},
		)
	default:
		output = r.ReplaceAllString(input, replacement)
	}

	return output
}

// replaceString replaces all matches in the filename
// with the replacement string.
func (op *Operation) replaceString(originalName string) string {
	return regexReplace(
		op.searchRegex,
		originalName,
		op.replacement,
		op.replaceLimit,
	)
}

// replace handles the replacement of matches in each file with the
// replacement string.
func (op *Operation) replace() (err error) {
	vars, err := extractVariables(op.replacement)
	if err != nil {
		return err
	}

	for i := range op.matches {
		ch := op.matches[i]
		ch.index = i
		originalName := ch.Source
		fileExt := filepath.Ext(originalName)

		if op.ignoreExt && !ch.IsDir {
			originalName = filenameWithoutExtension(originalName)
		}

		ch.Target = op.replaceString(originalName)

		// Replace any variables present with their corresponding values
		err = op.replaceVariables(&ch, &vars)
		if err != nil {
			return err
		}

		// Reattach the original extension to the new file name
		if op.ignoreExt && !ch.IsDir {
			ch.Target += fileExt
		}

		ch.Target = strings.TrimSpace(filepath.Clean(ch.Target))
		ch.status = statusOK
		op.matches[i] = ch
	}

	return nil
}
