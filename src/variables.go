package f2

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	exiftool "github.com/barasher/go-exiftool"
	"github.com/dhowden/tag"
	"github.com/rwcarlsen/goexif/exif"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"gopkg.in/djherbis/times.v1"
)

type hashAlgorithm string

const (
	sha1Hash   hashAlgorithm = "sha1"
	sha256Hash hashAlgorithm = "sha256"
	sha512Hash hashAlgorithm = "sha512"
	md5Hash    hashAlgorithm = "md5"
)

const (
	modTime     = "mtime"
	accessTime  = "atime"
	birthTime   = "btime"
	changeTime  = "ctime"
	currentTime = "now"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyz"
	numberBytes = "0123456789"
)

// Exif represents exif information from an image file.
type Exif struct {
	ISOSpeedRatings       []int
	DateTimeOriginal      string
	Make                  string
	Model                 string
	ExposureTime          []string
	FocalLength           []string
	FNumber               []string
	ImageWidth            []int
	ImageLength           []int // the image height
	LensModel             string
	Software              string
	FocalLengthIn35mmFilm []int
	PixelYDimension       []int
	PixelXDimension       []int
	Longitude             string
	Latitude              string
}

// ID3 represents id3 data from an audio file.
type ID3 struct {
	Format      string
	FileType    string
	Title       string
	Album       string
	Artist      string
	AlbumArtist string
	Genre       string
	Composer    string
	Year        int
	Track       int
	TotalTracks int
	Disc        int
	TotalDiscs  int
}

var (
	filenameRegex  = regexp.MustCompile("{{f}}")
	extensionRegex = regexp.MustCompile("{{ext}}")
	parentDirRegex = regexp.MustCompile("{{p}}")
	indexRegex     = regexp.MustCompile(
		`(\$\d+)?(\d+)?(%(\d?)+d)([borh])?(\d+)?(?:<(\d+(?:-\d+)?(?:,\s*\d+(?:-\d+)?)*)>)?`,
	)
	randomRegex = regexp.MustCompile(
		`{{(\d+)?r(?:(_l|_d|_ld)|(?:<(.*)>))?}}`,
	)
	hashRegex      = regexp.MustCompile(`{{hash.(sha1|sha256|sha512|md5)}}`)
	transformRegex = regexp.MustCompile(`{{tr.(up|lw|ti|win|mac|di)}}`)
	csvRegex       = regexp.MustCompile(`{{csv.(\d+)}}`)
	id3Regex       *regexp.Regexp
	exifRegex      *regexp.Regexp
	dateRegex      *regexp.Regexp
	exiftoolRegex  *regexp.Regexp
)

var dateTokens = map[string]string{
	"YYYY": "2006",
	"YY":   "06",
	"MMMM": "January",
	"MMM":  "Jan",
	"MM":   "01",
	"M":    "1",
	"DDDD": "Monday",
	"DDD":  "Mon",
	"DD":   "02",
	"D":    "2",
	"H":    "15",
	"hh":   "03",
	"h":    "3",
	"mm":   "04",
	"m":    "4",
	"ss":   "05",
	"s":    "5",
	"A":    "PM",
	"a":    "pm",
}

func init() {
	tokens := make([]string, 0, len(dateTokens))
	for key := range dateTokens {
		tokens = append(tokens, key)
	}

	tokenString := strings.Join(tokens, "|")
	dateRegex = regexp.MustCompile(
		"{{(" + modTime + "|" + changeTime + "|" + birthTime + "|" + accessTime + "|" + currentTime + ")\\.(" + tokenString + ")}}",
	)

	exiftoolRegex = regexp.MustCompile(`{{xt\.([0-9a-zA-Z]+)}}`)

	exifRegex = regexp.MustCompile(
		"{{(?:exif|x)\\.(iso|et|fl|w|h|wh|make|model|lens|fnum|fl35|lat|lon|soft)?(?:(dt)\\.(" + tokenString + "))?}}",
	)

	id3Regex = regexp.MustCompile(
		`{{id3\.(format|type|title|album|album_artist|artist|genre|year|composer|track|disc|total_tracks|total_discs)}}`,
	)

	rand.Seed(time.Now().UnixNano())
}

// randString returns a random string of the specified length
// using the specified characterSet.
func randString(n int, characterSet string) string {
	b := make([]byte, n)

	for i := range b {
		b[i] = characterSet[rand.Intn(len(characterSet))] //nolint:gosec // appropriate use of math.rand
	}

	return string(b)
}

// replaceRandomVariables replaces all random string variables
// in the target filename with a generated random string that matches
// the specifications.
func replaceRandomVariables(target string, rv randomVar) string {
	for i := range rv.submatches {
		r := rv.values[i]
		characters := r.characters

		switch characters {
		case "":
			characters = letterBytes
		case `_d`:
			characters = numberBytes
		case `_l`:
			characters = letterBytes
		case `_ld`:
			characters = letterBytes + numberBytes
		}

		target = r.regex.ReplaceAllString(
			target,
			randString(r.length, characters),
		)
	}

	return target
}

// integerToRoman converts an integer to a roman numeral
// For integers above 3999, it returns the stringified integer.
func integerToRoman(integer int) string {
	maxRomanNumber := 3999
	if integer > maxRomanNumber {
		return strconv.Itoa(integer)
	}

	conversions := []struct {
		value int
		digit string
	}{
		{1000, "M"},
		{900, "CM"},
		{500, "D"},
		{400, "CD"},
		{100, "C"},
		{90, "XC"},
		{50, "L"},
		{40, "XL"},
		{10, "X"},
		{9, "IX"},
		{5, "V"},
		{4, "IV"},
		{1, "I"},
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

// getHash retrieves the appropriate hash value for the specified file.
func getHash(file string, hashValue hashAlgorithm) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}

	defer f.Close()

	var h hash.Hash

	switch hashValue {
	case sha1Hash:
		h = sha1.New()
	case sha256Hash:
		h = sha256.New()
	case sha512Hash:
		h = sha512.New()
	case md5Hash:
		h = md5.New()
	default:
		return "", nil
	}

	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// replaceFileHash replaces a hash variable with the corresponding
// hash value.
func replaceFileHash(target, sourcePath string, hv hashVar) (string, error) {
	for i := range hv.submatches {
		h := hv.values[i]

		hashValue, err := getHash(sourcePath, h.hashFn)
		if err != nil {
			return "", err
		}

		target = h.regex.ReplaceAllString(target, hashValue)
	}

	return target, nil
}

// replaceDateVariables replaces any date variables in the target
// with the corresponding date value.
func replaceDateVariables(
	target, sourcePath string,
	dv dateVar,
) (string, error) {
	t, err := times.Stat(sourcePath)
	if err != nil {
		return "", err
	}

	for i := range dv.submatches {
		current := dv.values[i]
		regex := current.regex
		token := current.token

		var timeStr string

		switch current.attr {
		case modTime:
			modTime := t.ModTime()
			timeStr = modTime.Format(dateTokens[token])
		case birthTime:
			birthTime := t.ModTime()
			if t.HasBirthTime() {
				birthTime = t.BirthTime()
			}

			timeStr = birthTime.Format(dateTokens[token])
		case accessTime:
			accessTime := t.AccessTime()
			timeStr = accessTime.Format(dateTokens[token])
		case changeTime:
			changeTime := t.ModTime()
			if t.HasChangeTime() {
				changeTime = t.ChangeTime()
			}

			timeStr = changeTime.Format(dateTokens[token])
		case currentTime:
			currentTime := time.Now()
			timeStr = currentTime.Format(dateTokens[token])
		}

		target = regex.ReplaceAllString(target, timeStr)
	}

	return target, nil
}

// getID3Tags retrieves the id3 tags in an audi file (such as mp3)
// errors while reading the id3 tags are ignored since the corresponding
// variable will be replaced with an empty string.
func getID3Tags(sourcePath string) (*ID3, error) {
	f, err := os.Open(sourcePath)
	if err != nil {
		return nil, err
	}

	m, err := tag.ReadFrom(f)
	if err != nil {
		return &ID3{}, nil
	}

	trackNum, totalTracks := m.Track()
	discNum, totalDiscs := m.Disc()

	return &ID3{
		Format:      string(m.Format()),
		FileType:    string(m.FileType()),
		Title:       m.Title(),
		Album:       m.Album(),
		Artist:      m.Artist(),
		AlbumArtist: m.AlbumArtist(),
		Track:       trackNum,
		TotalTracks: totalTracks,
		Disc:        discNum,
		TotalDiscs:  totalDiscs,
		Composer:    m.Composer(),
		Year:        m.Year(),
		Genre:       m.Genre(),
	}, nil
}

// replaceID3Variables replaces all id3 variables in the target file name
// with the corresponding id3 tag value.
func replaceID3Variables(
	target, sourcePath string,
	id3v id3Var,
) (string, error) {
	tags, err := getID3Tags(sourcePath)
	if err != nil {
		return target, err
	}

	submatches := id3v.submatches
	for i := range submatches {
		current := id3v.values[i]
		regex := current.regex
		submatch := current.tag

		switch submatch {
		case "format":
			target = regex.ReplaceAllString(target, tags.Format)
		case "type":
			target = regex.ReplaceAllString(target, tags.FileType)
		case "title":
			target = regex.ReplaceAllString(target, tags.Title)
		case "album":
			target = regex.ReplaceAllString(target, tags.Album)
		case "artist":
			target = regex.ReplaceAllString(target, tags.Artist)
		case "album_artist":
			target = regex.ReplaceAllString(target, tags.AlbumArtist)
		case "genre":
			target = regex.ReplaceAllString(target, tags.Genre)
		case "composer":
			target = regex.ReplaceAllString(target, tags.Composer)
		case "track":
			var track string
			if tags.Track != 0 {
				track = strconv.Itoa(tags.Track)
			}

			target = regex.ReplaceAllString(target, track)
		case "total_tracks":
			var total string
			if tags.TotalTracks != 0 {
				total = strconv.Itoa(tags.TotalTracks)
			}

			target = regex.ReplaceAllString(target, total)
		case "disc":
			var disc string
			if tags.Disc != 0 {
				disc = strconv.Itoa(tags.Disc)
			}

			target = regex.ReplaceAllString(target, disc)
		case "total_discs":
			var total string
			if tags.TotalDiscs != 0 {
				total = strconv.Itoa(tags.TotalDiscs)
			}

			target = regex.ReplaceAllString(target, total)
		case "year":
			var year string
			if tags.Year != 0 {
				year = strconv.Itoa(tags.Year)
			}

			target = regex.ReplaceAllString(target, year)
		}
	}

	return target, nil
}

// getExifData retrieves the exif data embedded in an image file.
// Errors in decoding the exif data are ignored intentionally since
// the corresponding exif variable will be replaced by an empty
// string.
func getExifData(sourcePath string) (*Exif, error) {
	f, err := os.Open(sourcePath)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	exifData := &Exif{}

	x, err := exif.Decode(f)
	if err == nil {
		var b []byte

		b, err = x.MarshalJSON()
		if err == nil {
			_ = json.Unmarshal(b, exifData)
		}

		lat, lon, err := x.LatLong()
		if err == nil {
			exifData.Latitude = fmt.Sprintf("%.5f", lat)
			exifData.Longitude = fmt.Sprintf("%.5f", lon)
		}
	}

	return exifData, nil
}

// getExifExposureTime retrieves the exposure time from
// exif data. This exposure time may be a fraction
// so it is reduced to its simplest form and the
// forward slash is replaced with an underscore since
// it is forbidden in file names.
func getExifExposureTime(exifData *Exif) string {
	et := strings.Split(exifData.ExposureTime[0], "/")
	if len(et) == 1 {
		return et[0]
	}

	x, y := et[0], et[1]

	numerator, err := strconv.Atoi(x)
	if err != nil {
		return ""
	}

	denominator, err := strconv.Atoi(y)
	if err != nil {
		return ""
	}

	divisor := greatestCommonDivisor(numerator, denominator)
	if (numerator/divisor)%(denominator/divisor) == 0 {
		return fmt.Sprintf(
			"%d",
			(numerator/divisor)/(denominator/divisor),
		)
	}

	return fmt.Sprintf("%d_%d", numerator/divisor, denominator/divisor)
}

// getExifDate parses the exif original date and returns it
// in the specified format.
func getExifDate(exifData *Exif, format string) string {
	dateTimeString := exifData.DateTimeOriginal
	dateTimeSlice := strings.Split(dateTimeString, " ")

	// must include date and time components
	expectedLength := 2
	if len(dateTimeSlice) < expectedLength {
		return ""
	}

	dateString := strings.ReplaceAll(dateTimeSlice[0], ":", "-")
	timeString := dateTimeSlice[1]

	dateTime, err := time.Parse(time.RFC3339, dateString+"T"+timeString+"Z")
	if err != nil {
		return ""
	}

	return dateTime.Format(dateTokens[format])
}

// getDecimalFromFraction converts a value in the following format: [8/5]
// to its equivalent decimal value -> 1.6.
func getDecimalFromFraction(slice []string) string {
	if len(slice) == 0 {
		return ""
	}

	fractionSlice := strings.Split(slice[0], "/")

	expectedLength := 2
	if len(fractionSlice) != expectedLength {
		return ""
	}

	numerator, err := strconv.Atoi(fractionSlice[0])
	if err != nil {
		return ""
	}

	denominator, err := strconv.Atoi(fractionSlice[1])
	if err != nil {
		return ""
	}

	v := float64(numerator) / float64(denominator)

	bitSize := 64

	return strconv.FormatFloat(v, 'f', -1, bitSize)
}

// getExifDimensions retrieves the specified dimension
// w -> width, h -> height, wh -> width x height.
func getExifDimensions(exifData *Exif, dimension string) string {
	var w, h string
	if len(exifData.ImageWidth) > 0 {
		w = strconv.Itoa(exifData.ImageWidth[0])
	} else if len(exifData.PixelXDimension) > 0 {
		w = strconv.Itoa(exifData.PixelXDimension[0])
	}

	if len(exifData.ImageLength) > 0 {
		h = strconv.Itoa(exifData.ImageLength[0])
	} else if len(exifData.PixelYDimension) > 0 {
		h = strconv.Itoa(exifData.PixelYDimension[0])
	}

	switch dimension {
	case "w":
		return w
	case "h":
		return h
	case "wh":
		return w + "x" + h
	}

	return ""
}

// replaceExifVariables replaces the exif variables in an input string
// if an error occurs while attempting to get the value represented
// by the variables, it is replaced with an empty string.
func replaceExifVariables(
	target, sourcePath string,
	ev exifVar,
) (string, error) {
	exifData, err := getExifData(sourcePath)
	if err != nil {
		return target, err
	}

	for i := range ev.submatches {
		current := ev.values[i]
		regex := current.regex

		var value string

		switch current.attr {
		case "dt":
			value = getExifDate(exifData, current.timeStr)
		case "soft":
			value = exifData.Software
		case "model":
			value = strings.ReplaceAll(exifData.Model, "/", "_")
		case "lens":
			value = strings.ReplaceAll(exifData.LensModel, "/", "_")
		case "make":
			value = exifData.Make
		case "iso":
			if len(exifData.ISOSpeedRatings) > 0 {
				value = strconv.Itoa(exifData.ISOSpeedRatings[0])
			}
		case "et":
			if len(exifData.ExposureTime) > 0 {
				value = getExifExposureTime(exifData)
			}
		case "fnum":
			if len(exifData.FNumber) > 0 {
				value = getDecimalFromFraction(exifData.FNumber)
			}
		case "fl":
			if len(exifData.FocalLength) > 0 {
				value = getDecimalFromFraction(exifData.FocalLength)
			}
		case "fl35":
			if len(exifData.FocalLengthIn35mmFilm) > 0 {
				value = strconv.Itoa(exifData.FocalLengthIn35mmFilm[0])
			}
		case "lat":
			value = exifData.Latitude
		case "lon":
			value = exifData.Longitude
		case "wh", "h", "w":
			value = getExifDimensions(exifData, current.attr)
		}

		target = regex.ReplaceAllString(target, value)
	}

	return target, nil
}

// replaceExifToolVariables replaces the all exiftool
// variables in the target.
func replaceExifToolVariables(
	target, sourcePath string,
	ev exiftoolVar,
) (string, error) {
	et, err := exiftool.NewExiftool()
	if err != nil {
		return "", fmt.Errorf("Failed to initialise exiftool: %w", err)
	}

	defer et.Close()

	fileInfos := et.ExtractMetadata(sourcePath)

	for i := range ev.submatches {
		current := ev.values[i]
		regex := current.regex

		var value string

		for _, fileInfo := range fileInfos {
			if fileInfo.Err != nil {
				continue
			}

			for k, v := range fileInfo.Fields {
				if current.attr == k {
					value = fmt.Sprintf("%v", v)
					// replace forward and backward slashes with underscore
					value = strings.ReplaceAll(value, `/`, "_")
					value = strings.ReplaceAll(value, `\`, "_")

					break
				}
			}
		}

		target = regex.ReplaceAllString(target, value)
	}

	return target, nil
}

// replaceIndex replaces indexing variables in the target with their
// corresponding values. The `index` argument is used in conjunction with
// other values to increment the current index.
func (op *Operation) replaceIndex(
	target string,
	index int,
	nv numberVar,
) string {
	if len(op.numberOffset) == 0 {
		for range nv.submatches {
			op.numberOffset = append(op.numberOffset, 0)
		}
	}

	for i := range nv.submatches {
		current := nv.values[i]

		if current.step == 0 && !containsInt(nv.capturVarIndex, i) {
			current.step = 1
		}

		op.startNumber = current.startNumber
		num := op.startNumber + (index * current.step) + op.numberOffset[i]

		if len(current.skip) != 0 {
		outer:
			for {
				for _, v := range current.skip {
					if num >= v.min && num <= v.max {
						num += current.step
						op.numberOffset[i] += current.step
						continue outer
					}
				}
				break
			}
		}

		n := int64(num)

		var r string

		switch current.format {
		case "r":
			r = integerToRoman(num)
		case "h":
			base16 := 16
			r = strconv.FormatInt(n, base16)
		case "o":
			base8 := 8
			r = strconv.FormatInt(n, base8)
		case "b":
			base2 := 2
			r = strconv.FormatInt(n, base2)
		default:
			r = fmt.Sprintf(current.index, num)
		}

		target = current.regex.ReplaceAllString(target, r)
	}

	return target
}

// replaceTransformVariables handles string transformations like uppercase,
// lowercase, stripping characters, e.t.c.
func replaceTransformVariables(
	target string,
	matches []string,
	tv transformVar,
) string {
	for i := range tv.submatches {
		current := tv.values[i]
		r := current.regex

		for _, v := range matches {
			switch current.token {
			case "up":
				target = regexReplace(r, target, strings.ToUpper(v), 1)
			case "lw":
				target = regexReplace(r, target, strings.ToLower(v), 1)
			case "ti":
				target = regexReplace(
					r,
					target,
					strings.Title(strings.ToLower(v)),
					1,
				)
			case "win":
				target = regexReplace(
					r,
					target,
					regexReplace(fullWindowsForbiddenCharRegex, v, "", 0),
					1,
				)
			case "mac":
				target = regexReplace(
					r,
					target,
					regexReplace(macForbiddenCharRegex, v, "", 0),
					1,
				)
			case "di":
				t := transform.Chain(
					norm.NFD,
					runes.Remove(runes.In(unicode.Mn)),
					norm.NFC,
				)

				result, _, err := transform.String(t, v)
				if err != nil {
					return v
				}

				target = regexReplace(r, target, result, 1)
			}
		}
	}

	return target
}

// replaceCsvVariables inserts the appropriate CSV column
// in the replacement target or an empty string if the column
// is not present in the row.
func replaceCsvVariables(target string, csvRow []string, cv csvVar) string {
	for i := range cv.submatches {
		current := cv.values[i]
		column := current.column - 1
		r := current.regex

		var value string

		if len(csvRow) > column && column >= 0 {
			value = csvRow[column]
		}

		target = r.ReplaceAllString(target, value)
	}

	return target
}

// replaceVariables checks if any variables are present in the target filename
// and delegates the variable replacement to the appropriate function.
func (op *Operation) replaceVariables(
	ch *Change,
	vars *variables,
) error {
	sourceName := ch.Source
	fileExt := filepath.Ext(sourceName)
	parentDir := filepath.Base(ch.BaseDir)
	sourcePath := filepath.Join(ch.BaseDir, ch.originalSource)

	if parentDir == "." {
		// Set to base folder of current working directory
		parentDir = filepath.Base(op.workingDir)
	}

	// replace `{{f}}` in the target with the original filename
	// (excluding the extension)
	if filenameRegex.MatchString(ch.Target) {
		ch.Target = regexReplace(
			filenameRegex,
			ch.Target,
			filenameWithoutExtension(sourceName),
			0,
		)
	}

	// replace `{{ext}}` in the target with the file extension
	if extensionRegex.MatchString(ch.Target) {
		ch.Target = regexReplace(extensionRegex, ch.Target, fileExt, 0)
	}

	// replace `{{p}}` in the target with the parent directory name
	if parentDirRegex.MatchString(ch.Target) {
		ch.Target = regexReplace(parentDirRegex, ch.Target, parentDir, 0)
	}

	// handle date variables (e.g {{mtime.DD}})
	if dateRegex.MatchString(ch.Target) {
		out, err := replaceDateVariables(ch.Target, sourcePath, vars.date)
		if err != nil {
			return err
		}

		ch.Target = out
	}

	if exiftoolRegex.MatchString(ch.Target) {
		out, err := replaceExifToolVariables(
			ch.Target,
			sourcePath,
			vars.exiftool,
		)
		if err != nil {
			return err
		}

		ch.Target = out
	}

	if exifRegex.MatchString(ch.Target) {
		out, err := replaceExifVariables(ch.Target, sourcePath, vars.exif)
		if err != nil {
			return err
		}

		ch.Target = out
	}

	if id3Regex.MatchString(ch.Target) {
		out, err := replaceID3Variables(ch.Target, sourcePath, vars.id3)
		if err != nil {
			return err
		}

		ch.Target = out
	}

	if csvRegex.MatchString(ch.Target) {
		out := replaceCsvVariables(ch.Target, ch.csvRow, vars.csv)

		ch.Target = out
	}

	if hashRegex.MatchString(ch.Target) {
		out, err := replaceFileHash(ch.Target, sourcePath, vars.hash)
		if err != nil {
			return err
		}

		ch.Target = out
	}

	if randomRegex.MatchString(ch.Target) {
		ch.Target = replaceRandomVariables(ch.Target, vars.random)
	}

	if transformRegex.MatchString(ch.Target) {
		if op.ignoreExt && !ch.IsDir {
			sourceName = filenameWithoutExtension(sourceName)
		}

		ch.Target = replaceTransformVariables(
			ch.Target,
			op.searchRegex.FindAllString(sourceName, -1),
			vars.transform,
		)
	}

	// Replace indexing scheme like %03d in the target
	if indexRegex.MatchString(ch.Target) {
		if len(vars.number.capturVarIndex) > 0 {
			indices := make([]int, len(vars.number.capturVarIndex))

			copy(indices, vars.number.capturVarIndex)

			numVar, err := getNumberVar(ch.Target)
			if err != nil {
				return err
			}

			vars.number = numVar
			vars.number.capturVarIndex = indices
		}

		ch.Target = op.replaceIndex(ch.Target, ch.index, vars.number)
	}

	return nil
}
