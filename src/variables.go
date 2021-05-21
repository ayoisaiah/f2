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

var (
	filenameRegex  = regexp.MustCompile("{{f}}")
	extensionRegex = regexp.MustCompile("{{ext}}")
	parentDirRegex = regexp.MustCompile("{{p}}")
	indexRegex     = regexp.MustCompile(
		`(\d+)?(%(\d?)+d)([borh])?(\d+)?(?:<(\d+(?:-\d+)?(?:,\s*\d+(?:-\d+)?)*)>)?`,
	)
	randomRegex = regexp.MustCompile(
		`{{(\d+)?r(?:(_l|_d|_ld)|(?:<(.*)>))?}}`,
	)
	hashRegex      = regexp.MustCompile(`{{hash.(sha1|sha256|sha512|md5)}}`)
	transformRegex = regexp.MustCompile(`{{tr.(up|lw|ti|win|mac|di)}}`)
	id3Regex       *regexp.Regexp
	exifRegex      *regexp.Regexp
	dateRegex      *regexp.Regexp
	exiftoolRegex  *regexp.Regexp
)

const (
	sha1Hash   = "sha1"
	sha256Hash = "sha256"
	sha512Hash = "sha512"
	md5Hash    = "md5"
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

var lettersAndNumbers = letterBytes + numberBytes

// Exif represents exif information from an image file
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

// ID3 represents id3 data from an audio file
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
// using the specified characterSet
func randString(n int, characterSet string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = characterSet[rand.Intn(len(characterSet))]
	}
	return string(b)
}

// replaceRandomVariables reolaces `{{r}}` in the string with a generated
// random string
func replaceRandomVariables(input string, rv randomVar) string {
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
			characters = lettersAndNumbers
		}

		input = r.regex.ReplaceAllString(
			input,
			randString(r.length, characters),
		)
	}

	return input
}

// integerToRoman converts an integer to a roman numeral
func integerToRoman(number int) string {
	maxRomanNumber := 3999
	if number > maxRomanNumber {
		return strconv.Itoa(number)
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
		for number >= conversion.value {
			roman.WriteString(conversion.digit)
			number -= conversion.value
		}
	}

	return roman.String()
}

func getHash(file, hashFn string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}

	defer f.Close()

	var h hash.Hash

	switch hashFn {
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
// hash value
func replaceFileHash(input, filePath string, hv hashVar) (string, error) {
	for i := range hv.submatches {
		h := hv.values[i]

		hashValue, err := getHash(filePath, h.hashFn)
		if err != nil {
			return "", err
		}

		input = h.regex.ReplaceAllString(input, hashValue)
	}

	return input, nil
}

// replaceDateVariables replaces a date variable with the corresponding
// date value
func replaceDateVariables(input, filePath string, dv dateVar) (string, error) {
	t, err := times.Stat(filePath)
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

		input = regex.ReplaceAllString(input, timeStr)
	}

	return input, nil
}

// getID3Tags retrieves the id3 tags in an audi file (such as mp3)
// errors while reading the id3 tags are ignored since the corresponding
// variable will be replaced with an empty string
func getID3Tags(filePath string) (*ID3, error) {
	f, err := os.Open(filePath)
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

// replaceID3Variables replaces an id3 variable in the input string
// with the corresponding id3 value
func replaceID3Variables(tags *ID3, input string, id3v id3Var) string {
	submatches := id3v.submatches
	for i := range submatches {
		current := id3v.values[i]
		regex := current.regex
		submatch := current.tag

		switch submatch {
		case "format":
			input = regex.ReplaceAllString(input, tags.Format)
		case "type":
			input = regex.ReplaceAllString(input, tags.FileType)
		case "title":
			input = regex.ReplaceAllString(input, tags.Title)
		case "album":
			input = regex.ReplaceAllString(input, tags.Album)
		case "artist":
			input = regex.ReplaceAllString(input, tags.Artist)
		case "album_artist":
			input = regex.ReplaceAllString(input, tags.AlbumArtist)
		case "genre":
			input = regex.ReplaceAllString(input, tags.Genre)
		case "composer":
			input = regex.ReplaceAllString(input, tags.Composer)
		case "track":
			var track string
			if tags.Track != 0 {
				track = strconv.Itoa(tags.Track)
			}
			input = regex.ReplaceAllString(input, track)
		case "total_tracks":
			var total string
			if tags.TotalTracks != 0 {
				total = strconv.Itoa(tags.TotalTracks)
			}
			input = regex.ReplaceAllString(input, total)
		case "disc":
			var disc string
			if tags.Disc != 0 {
				disc = strconv.Itoa(tags.Disc)
			}
			input = regex.ReplaceAllString(input, disc)
		case "total_discs":
			var total string
			if tags.TotalDiscs != 0 {
				total = strconv.Itoa(tags.TotalDiscs)
			}
			input = regex.ReplaceAllString(input, total)
		case "year":
			var year string
			if tags.Year != 0 {
				year = strconv.Itoa(tags.Year)
			}
			input = regex.ReplaceAllString(input, year)
		}
	}

	return input
}

// getExifData retrieves the exif data embedded in an image file.
// Errors in decoding the exif data are ignored intentionally since
// the corresponding exif variable will be replaced by an empty
// string
func getExifData(filePath string) (*Exif, error) {
	f, err := os.Open(filePath)
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

// replaceExifVariables replaces the exif variables in an input string
func replaceExifVariables(
	exifData *Exif,
	input string,
	ev exifVar,
) (string, error) {
	for i := range ev.submatches {
		current := ev.values[i]
		regex := current.regex

		var value string
		switch current.attr {
		case "dt":
			date := exifData.DateTimeOriginal
			arr := strings.Split(date, " ")
			if len(arr) > 1 {
				var dt time.Time
				d := strings.ReplaceAll(arr[0], ":", "-")
				t := arr[1]
				var err error
				dt, err = time.Parse(time.RFC3339, d+"T"+t+"Z")
				if err != nil {
					return "", err
				}

				value = dt.Format(dateTokens[current.timeStr])
			}
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
				et := strings.Split(exifData.ExposureTime[0], "/")
				if len(et) == 1 {
					value = et[0]
					break
				}

				x, y := et[0], et[1]
				numerator, err := strconv.Atoi(x)
				if err != nil {
					value = exifData.ExposureTime[0]
					break
				}

				denominator, err := strconv.Atoi(y)
				if err != nil {
					value = exifData.ExposureTime[0]
					break
				}

				divisor := greatestCommonDivisor(numerator, denominator)
				if (numerator/divisor)%(denominator/divisor) == 0 {
					value = fmt.Sprintf(
						"%d",
						(numerator/divisor)/(denominator/divisor),
					)
				} else {
					value = fmt.Sprintf("%d_%d", numerator/divisor, denominator/divisor)
				}
			}
		case "fnum":
			value = exifDivision(exifData.FNumber)
		case "fl":
			value = exifDivision(exifData.FocalLength)
		case "fl35":
			if len(exifData.FocalLengthIn35mmFilm) > 0 {
				value = strconv.Itoa(exifData.FocalLengthIn35mmFilm[0])
			}
		case "lat":
			value = exifData.Latitude
		case "lon":
			value = exifData.Longitude
		case "wh":
			if len(exifData.ImageLength) > 0 && len(exifData.ImageWidth) > 0 {
				h, w := exifData.ImageLength[0], exifData.ImageWidth[0]
				value = strconv.Itoa(w) + "x" + strconv.Itoa(h)
				break
			}

			if len(exifData.PixelXDimension) > 0 &&
				len(exifData.PixelYDimension) > 0 {
				h, w := exifData.PixelYDimension[0], exifData.PixelXDimension[0]
				value = strconv.Itoa(w) + "x" + strconv.Itoa(h)
			}
		case "h":
			if len(exifData.ImageLength) > 0 {
				value = strconv.Itoa(exifData.ImageLength[0])
				break
			}

			if len(exifData.PixelYDimension) > 0 {
				value = strconv.Itoa(exifData.PixelYDimension[0])
			}
		case "w":
			if len(exifData.ImageWidth) > 0 {
				value = strconv.Itoa(exifData.ImageWidth[0])
				break
			}

			if len(exifData.PixelXDimension) > 0 {
				value = strconv.Itoa(exifData.PixelXDimension[0])
			}
		}
		input = regex.ReplaceAllString(input, value)
	}

	return input, nil
}

func replaceExifToolVariables(
	input, filePath string,
	ev exiftoolVar,
) (string, error) {
	et, err := exiftool.NewExiftool()
	if err != nil {
		return "", fmt.Errorf("Failed to initialise exiftool: %w", err)
	}

	defer et.Close()

	fileInfos := et.ExtractMetadata(filePath)

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

		input = regex.ReplaceAllString(input, value)
	}

	return input, nil
}

// replaceIndex deals with sequential numbering in various formats
func (op *Operation) replaceIndex(
	input string,
	count int,
	nv numberVar,
) string {
	if len(op.numberOffset) == 0 {
		for range nv.submatches {
			op.numberOffset = append(op.numberOffset, 0)
		}
	}

	for i := range nv.submatches {
		current := nv.values[i]

		op.startNumber = current.startNumber
		num := op.startNumber + (count * current.step) + op.numberOffset[i]
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
			r = strconv.FormatInt(n, 16)
		case "o":
			r = strconv.FormatInt(n, 8)
		case "b":
			r = strconv.FormatInt(n, 2)
		default:
			r = fmt.Sprintf(current.index, num)
		}

		input = current.regex.ReplaceAllString(input, r)
	}

	return input
}

// replaceTransformVariables handles string transformations like uppercase,
// lowercase, stripping characters, e.t.c
func replaceTransformVariables(
	input string,
	matches []string,
	tv transformVar,
) string {
	for i := range tv.submatches {
		current := tv.values[i]
		r := current.regex
		for _, v := range matches {
			switch current.token {
			case "up":
				input = regexReplace(r, input, strings.ToUpper(v), 1)
			case "lw":
				input = regexReplace(r, input, strings.ToLower(v), 1)
			case "ti":
				input = regexReplace(
					r,
					input,
					strings.Title(strings.ToLower(v)),
					1,
				)
			case "win":
				input = regexReplace(
					r,
					input,
					regexReplace(fullWindowsForbiddenRegex, v, "", 0),
					1,
				)
			case "mac":
				input = regexReplace(
					r,
					input,
					regexReplace(macForbiddenRegex, v, "", 0),
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

				input = regexReplace(r, input, result, 1)
			}
		}
	}

	return input
}

// handleVariables checks if any variables are present in the replacement
// string and delegates the variable replacement to the appropriate
// function
func (op *Operation) handleVariables(
	input string,
	ch Change,
	vars *replaceVars,
) (string, error) {
	fileName := ch.Source
	fileExt := filepath.Ext(fileName)
	parentDir := filepath.Base(ch.BaseDir)
	sourcePath := filepath.Join(ch.BaseDir, ch.originalSource)

	if parentDir == "." {
		// Set to base folder of current working directory
		parentDir = filepath.Base(op.workingDir)
	}

	// replace `{{f}}` in the replacement string with the original
	// filename (without the extension)
	if filenameRegex.MatchString(input) {
		input = filenameRegex.ReplaceAllString(
			input,
			filenameWithoutExtension(fileName),
		)
	}

	// replace `{{ext}}` in the replacement string with the file extension
	if extensionRegex.MatchString(input) {
		input = extensionRegex.ReplaceAllString(input, fileExt)
	}

	// replace `{{p}}` in the replacement string with the parent directory name
	if parentDirRegex.MatchString(input) {
		input = parentDirRegex.ReplaceAllString(input, parentDir)
	}

	// handle date variables (e.g {{mtime.DD}})
	if dateRegex.MatchString(input) {
		out, err := replaceDateVariables(input, sourcePath, vars.date)
		if err != nil {
			return "", err
		}
		input = out
	}

	if exiftoolRegex.MatchString(input) {
		out, err := replaceExifToolVariables(input, sourcePath, vars.exiftool)
		if err != nil {
			return "", err
		}
		input = out
	}

	if exifRegex.MatchString(input) {
		exifData, err := getExifData(sourcePath)
		if err != nil {
			return "", err
		}

		out, err := replaceExifVariables(exifData, input, vars.exif)
		if err != nil {
			return "", err
		}
		input = out
	}

	if id3Regex.MatchString(input) {
		tags, err := getID3Tags(sourcePath)
		if err != nil {
			return "", err
		}

		input = replaceID3Variables(tags, input, vars.id3)
	}

	if hashRegex.MatchString(input) {
		out, err := replaceFileHash(input, sourcePath, vars.hash)
		if err != nil {
			return "", err
		}
		input = out
	}

	if randomRegex.MatchString(input) {
		input = replaceRandomVariables(input, vars.random)
	}

	if transformRegex.MatchString(input) {
		if op.ignoreExt {
			fileName = filenameWithoutExtension(fileName)
		}

		input = replaceTransformVariables(
			input,
			op.searchRegex.FindAllString(fileName, -1),
			vars.transform,
		)
	}

	return input, nil
}
