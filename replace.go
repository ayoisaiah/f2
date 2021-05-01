package f2

import (
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

type numberVar struct {
	submatches [][]string
	values     []struct {
		regex       *regexp.Regexp
		startNumber int
		index       string
		format      string
		step        int
		skip        []numbersToSkip
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
		hashFn string
	}
}

type randomVar struct {
	submatches [][]string
	values     []struct {
		regex      *regexp.Regexp
		length     int
		characters string
	}
}

type replaceVars struct {
	exif   exifVar
	number numberVar
	id3    id3Var
	hash   hashVar
	date   dateVar
	random randomVar
}

func getDateVar(str string) (dateVar, error) {
	var d dateVar
	if dateRegex.MatchString(str) {
		d.submatches = dateRegex.FindAllStringSubmatch(str, -1)

		for _, submatch := range d.submatches {
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

func getHashVar(str string) (hashVar, error) {
	var h hashVar
	if hashRegex.MatchString(str) {
		h.submatches = hashRegex.FindAllStringSubmatch(str, -1)

		for _, submatch := range h.submatches {
			var x struct {
				regex  *regexp.Regexp
				hashFn string
			}
			regex, err := regexp.Compile(submatch[0])
			if err != nil {
				return h, err
			}

			x.regex = regex
			x.hashFn = submatch[1]
			h.values = append(h.values, x)
		}
	}

	return h, nil
}

func getExifVar(str string) (exifVar, error) {
	var ex exifVar

	if exifRegex.MatchString(str) {
		ex.submatches = exifRegex.FindAllStringSubmatch(str, -1)
		for _, submatch := range ex.submatches {
			var x struct {
				regex   *regexp.Regexp
				attr    string
				timeStr string
			}
			regex, err := regexp.Compile(submatch[0])
			if err != nil {
				return ex, err
			}

			x.regex = regex

			if strings.Contains(submatch[0], "exif.dt") {
				submatch = append(submatch[:1], submatch[1+1:]...)
			}

			x.attr = submatch[1]
			if x.attr == "dt" {
				x.timeStr = submatch[2]
			}

			ex.values = append(ex.values, x)
		}
	}

	return ex, nil
}

func getNumberVar(str string) (numberVar, error) {
	var nv numberVar

	if indexRegex.MatchString(str) {
		nv.submatches = indexRegex.FindAllStringSubmatch(str, -1)
		for _, submatch := range nv.submatches {
			var n struct {
				regex       *regexp.Regexp
				startNumber int
				index       string
				format      string
				step        int
				skip        []numbersToSkip
			}

			regex, err := regexp.Compile(submatch[0])
			if err != nil {
				return nv, err
			}

			n.regex = regex

			if submatch[1] != "" {
				n.startNumber, err = strconv.Atoi(submatch[1])
				if err != nil {
					return nv, err
				}
			} else {
				n.startNumber = 1
			}

			n.index = submatch[2]
			n.format = submatch[4]
			n.step = 1
			if submatch[5] != "" {
				n.step, err = strconv.Atoi(submatch[5])
				if err != nil {
					return nv, err
				}
			}

			skipNumbers := submatch[6]
			if skipNumbers != "" {
				slice := strings.Split(skipNumbers, ",")
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

						n.skip = append(n.skip, numbersToSkip{
							max: int(math.Max(float64(n1), float64(n2))),
							min: int(math.Min(float64(n1), float64(n2))),
						})
						continue
					}

					num, err := strconv.Atoi(v)
					if err != nil {
						return nv, err
					}

					n.skip = append(n.skip, numbersToSkip{
						max: num,
						min: num,
					})
				}
			}

			nv.values = append(nv.values, n)
		}
	}

	return nv, nil
}

func getID3Var(str string) (id3Var, error) {
	var i id3Var
	if id3Regex.MatchString(str) {
		i.submatches = id3Regex.FindAllStringSubmatch(str, -1)
		for _, submatch := range i.submatches {
			var x struct {
				regex *regexp.Regexp
				tag   string
			}
			regex, err := regexp.Compile(submatch[0])
			if err != nil {
				return i, err
			}

			x.regex = regex
			x.tag = submatch[1]

			i.values = append(i.values, x)
		}
	}

	return i, nil
}

func getRandomVar(str string) (randomVar, error) {
	var rv randomVar

	if randomRegex.MatchString(str) {
		rv.submatches = randomRegex.FindAllStringSubmatch(str, -1)

		for _, submatch := range rv.submatches {
			var r struct {
				regex      *regexp.Regexp
				length     int
				characters string
			}
			r.length = 10
			regex, err := regexp.Compile(submatch[0])
			if err != nil {
				return rv, err
			}
			r.regex = regex

			strLen := submatch[1]
			if strLen != "" {
				r.length, err = strconv.Atoi(strLen)
				if err != nil {
					return rv, err
				}
			}

			r.characters = submatch[2]

			if submatch[3] != "" {
				r.characters = submatch[3]
			}

			rv.values = append(rv.values, r)
		}
	}

	return rv, nil
}

func getAllVariables(str string) (replaceVars, error) {
	var v replaceVars
	var err error
	v.exif, err = getExifVar(str)
	if err != nil {
		return v, err
	}

	v.number, err = getNumberVar(str)
	if err != nil {
		return v, err
	}

	v.id3, err = getID3Var(str)
	if err != nil {
		return v, err
	}

	v.hash, err = getHashVar(str)
	if err != nil {
		return v, err
	}

	v.date, err = getDateVar(str)
	if err != nil {
		return v, err
	}

	v.random, err = getRandomVar(str)
	if err != nil {
		return v, err
	}

	return v, nil
}

func (op *Operation) replaceString(fileName string) (str string) {
	findString := op.findString
	if findString == "" {
		findString = fileName
	}
	replacement := op.replacement

	if strings.HasPrefix(replacement, `\T`) {
		matches := op.searchRegex.FindAllString(fileName, -1)
		str = fileName
		for _, v := range matches {
			switch replacement {
			case `\Tcu`:
				str = strings.ReplaceAll(str, v, strings.ToUpper(v))
			case `\Tcl`:
				str = strings.ReplaceAll(str, v, strings.ToLower(v))
			case `\Tct`:
				str = strings.ReplaceAll(
					str,
					v,
					strings.Title(strings.ToLower(v)),
				)
			case `\Twin`:
				str = fullWindowsForbiddenRegx.ReplaceAllString(str, "")
			case `\Tmac`:
				str = strings.ReplaceAll(str, ":", "")
			}
		}
		return
	}

	if op.stringMode {
		if op.ignoreCase {
			str = op.searchRegex.ReplaceAllString(fileName, replacement)
		} else {
			str = strings.ReplaceAll(fileName, findString, replacement)
		}
	} else {
		str = op.searchRegex.ReplaceAllString(fileName, replacement)
	}

	return str
}

// replace replaces the matched text in each path with the
// replacement string
func (op *Operation) replace() (err error) {
	vars, err := getAllVariables(op.replacement)
	if err != nil {
		return err
	}

	for i, v := range op.matches {
		fileName, dir := filepath.Base(v.Source), filepath.Dir(v.Source)
		fileExt := filepath.Ext(fileName)
		if op.ignoreExt {
			fileName = filenameWithoutExtension(fileName)
		}

		str := op.replaceString(fileName)

		// handle variables
		str, err = op.handleVariables(str, v, &vars)
		if err != nil {
			return err
		}

		// If numbering scheme is present
		if indexRegex.MatchString(str) {
			str = op.replaceIndex(str, i, vars.number)
		}

		if op.ignoreExt {
			str += fileExt
		}

		v.Target = filepath.Join(dir, str)
		op.matches[i] = v
	}

	return nil
}
