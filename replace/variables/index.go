package variables

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
)

// Replace replaces indexing variables in the target with their
// corresponding values.
func (iv *indexVars) Replace(
	_ context.Context,
	conf *config.Config,
	change *file.Change,
) error {
	if conf.ResetIndexPerDir {
		// Detect when a new directory is entered
		if change.BaseDir != iv.currentBaseDir {
			// track the position at which the base directory changed
			iv.newDirIndex = change.Position
		}
	}

	iv.currentBaseDir = change.BaseDir

	// This has the effect of resetting the index for a new directory when the
	// `ResetIndexPerDir` option is set
	changeIndex := change.Position - iv.newDirIndex

	if !indexVarRegex.MatchString(change.Target) {
		return nil
	}

	if len(iv.capturVarIndex) > 0 {
		// The captureVariable has been replaced with the real value at this point
		// so retriveing the indexing vars will now provide the correct `startNumber`
		// value
		numVar, err := getIndexingVars(change.Target)
		if err != nil {
			return err
		}

		iv.matches = numVar.matches
		iv.offset = numVar.offset
	}

	change.Target = replaceIndex(change.Target, changeIndex, iv)

	return nil
}

// integerToRoman converts an integer to a roman numeral
// For integers above 3999, it returns the stringified integer.
func integerToRoman(integer int) string {
	maxRomanNumber := 3999
	if integer > maxRomanNumber {
		return strconv.Itoa(integer)
	}

	conversions := []struct {
		digit string
		value int
	}{
		{"M", 1000},
		{"CM", 900},
		{"D", 500},
		{"CD", 400},
		{"C", 100},
		{"XC", 90},
		{"L", 50},
		{"XL", 40},
		{"X", 10},
		{"IX", 9},
		{"V", 5},
		{"IV", 4},
		{"I", 1},
	}

	var roman strings.Builder

	for _, conversion := range conversions {
		for integer >= conversion.value {
			roman.WriteString(conversion.digit)
			integer -= conversion.value
		}
	}

	return roman.String()
}

// replaceIndex replaces indexing variables in the target with their
// corresponding values. The `changeIndex` argument is used in conjunction with
// other values to increment the current index.
func replaceIndex(
	target string,
	changeIndex int, // position of change in the entire renaming operation
	indexing *indexVars,
) string {
	for i := range indexing.matches {
		current := indexing.matches[i]

		if !current.step.isSet && !current.isCaptureVar {
			current.step.value = 1
		}

		startNumber := current.startNumber
		currentIndex := startNumber + (changeIndex * current.step.value) + indexing.offset[i]

		if current.isCaptureVar {
			currentIndex = startNumber + current.step.value + indexing.offset[i]
		}

		if len(current.skip) != 0 {
		outer:
			for {
				for _, v := range current.skip {
					//nolint:gocritic // nesting is manageable
					if currentIndex >= v.min && currentIndex <= v.max {
						// Prevent infinite loops when skipping a captured variable
						step := current.step.value
						if step == 0 {
							step = 1
						}

						currentIndex += step

						if !current.isCaptureVar {
							indexing.offset[i] += step
						}

						continue outer
					}
				}

				break
			}
		}

		numInt64 := int64(currentIndex)

		var formattedNum string

		switch current.numberSystem {
		case "r":
			formattedNum = integerToRoman(currentIndex)
		case "h":
			base16 := 16
			formattedNum = strconv.FormatInt(numInt64, base16)
		case "o":
			base8 := 8
			formattedNum = strconv.FormatInt(numInt64, base8)
		case "b":
			base2 := 2
			formattedNum = strconv.FormatInt(numInt64, base2)
		default:
			if currentIndex < 0 {
				currentIndex *= -1
				formattedNum = "-" + fmt.Sprintf(
					current.indexFormat,
					currentIndex,
				)
			} else {
				formattedNum = fmt.Sprintf(current.indexFormat, currentIndex)
			}
		}

		target = current.regex.ReplaceAllString(target, formattedNum)
	}

	return target
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
