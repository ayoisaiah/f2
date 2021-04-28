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

	"github.com/dhowden/tag"
	"github.com/rwcarlsen/goexif/exif"
	"gopkg.in/djherbis/times.v1"
)

var (
	filenameRegex  = regexp.MustCompile("{{f}}")
	extensionRegex = regexp.MustCompile("{{ext}}")
	parentDirRegex = regexp.MustCompile("{{p}}")
	indexRegex     = regexp.MustCompile(`(\d+)?(%(\d?)+d)([borh])?(\d+)?`)
	randomRegex    = regexp.MustCompile(
		`{{(\d+)?r(?:(_l|_d|_ld)|(?:<(.*)>))?}}`,
	)
	hashRegex = regexp.MustCompile(`{{hash.(sha1|sha256|sha512|md5)}}`)
	id3Regex  *regexp.Regexp
	exifRegex *regexp.Regexp
	dateRegex *regexp.Regexp
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
	ISOSpeedRatings  []int
	DateTimeOriginal string
	Make             string
	Model            string
	ExposureTime     []string
	FocalLength      []string
	FNumber          []string
	ImageWidth       []int
	ImageLength      []int // the image height
	LensModel        string
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

	exifRegex = regexp.MustCompile(
		"{{(?:exif|x)\\.(iso|et|fl|w|h|wh|make|model|lens|fnum)?(?:(dt)\\.(" + tokenString + "))?}}",
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
func replaceRandomVariables(str string, rv randomVar) string {
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

		str = r.regex.ReplaceAllString(str, randString(r.length, characters))
	}

	return str
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
func replaceFileHash(file, input string, hv hashVar) (string, error) {
	for i := range hv.submatches {
		h := hv.values[i]

		hashValue, err := getHash(file, h.hashFn)
		if err != nil {
			return "", err
		}

		input = h.regex.ReplaceAllString(input, hashValue)
	}

	return input, nil
}

// replaceDateVariables replaces a date variable with the corresponding
// date value
func replaceDateVariables(file, input string, dv dateVar) (string, error) {
	t, err := times.Stat(file)
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
func getID3Tags(file string) (*ID3, error) {
	f, err := os.Open(file)
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
func getExifData(file string) (*Exif, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer func() {
		ferr := f.Close()
		if ferr != nil {
			err = ferr
		}
	}()

	exifData := &Exif{}
	x, err := exif.Decode(f)
	if err == nil {
		b, err := x.MarshalJSON()
		if err == nil {
			_ = json.Unmarshal(b, exifData)
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

		switch current.attr {
		case "dt":
			date := exifData.DateTimeOriginal
			arr := strings.Split(date, " ")
			var dt time.Time
			d := strings.ReplaceAll(arr[0], ":", "-")
			t := arr[1]
			var err error
			dt, err = time.Parse(time.RFC3339, d+"T"+t+"Z")
			if err != nil {
				return "", err
			}

			timeStr := dt.Format(dateTokens[current.timeStr])
			input = regex.ReplaceAllString(input, timeStr)
		case "model":
			cmodel := exifData.Model
			cmodel = strings.ReplaceAll(cmodel, "/", "_")
			input = regex.ReplaceAllString(input, cmodel)
		case "lens":
			lens := exifData.LensModel
			lens = strings.ReplaceAll(lens, "/", "_")
			input = regex.ReplaceAllString(input, lens)
		case "make":
			cmake := exifData.Make
			input = regex.ReplaceAllString(input, cmake)
		case "iso":
			var iso string
			if len(exifData.ISOSpeedRatings) > 0 {
				iso = strconv.Itoa(exifData.ISOSpeedRatings[0])
			}
			input = regex.ReplaceAllString(input, "ISO"+iso)
		case "et":
			var et string
			if len(exifData.ExposureTime) > 0 {
				et = exifData.ExposureTime[0]
				et = strings.ReplaceAll(et, "/", "_")
			}
			input = regex.ReplaceAllString(input, et+"s")
		case "fnum":
			v := exifDivision(exifData.FNumber)
			input = regex.ReplaceAllString(input, "f"+v)
		case "fl":
			v := exifDivision(exifData.FocalLength)
			input = regex.ReplaceAllString(input, v+"mm")
		case "wh":
			var wh string
			if len(exifData.ImageLength) > 0 && len(exifData.ImageWidth) > 0 {
				h, w := exifData.ImageLength[0], exifData.ImageWidth[0]
				wh = strconv.Itoa(w) + "x" + strconv.Itoa(h)
			}
			input = regex.ReplaceAllString(input, wh)
		case "h":
			var h string
			if len(exifData.ImageLength) > 0 {
				h = strconv.Itoa(exifData.ImageLength[0])
			}
			input = regex.ReplaceAllString(input, h)
		case "w":
			var w string
			if len(exifData.ImageWidth) > 0 {
				w = strconv.Itoa(exifData.ImageWidth[0])
			}
			input = regex.ReplaceAllString(input, w)
		}
	}

	return input, nil
}

// replaceIndex deals with sequential numbering in various formats
func (op *Operation) replaceIndex(
	str string,
	count int,
	nv numberVar,
) string {
	for i := range nv.submatches {
		current := nv.values[i]
		op.startNumber = current.startNumber
		num := op.startNumber + (count * current.step)
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

		str = current.regex.ReplaceAllString(str, r)
	}

	return str
}

// handleVariables checks if any variables are present in the replacement
// string and delegates the variable replacement to the appropriate
// function
func (op *Operation) handleVariables(
	str string,
	ch Change,
	vars *replaceVars,
) (string, error) {
	fileName := filepath.Base(ch.Source)
	fileExt := filepath.Ext(fileName)
	parentDir := filepath.Base(ch.BaseDir)
	source := filepath.Join(ch.BaseDir, ch.Source)

	if parentDir == "." {
		// Set to base folder of current working directory
		parentDir = filepath.Base(op.workingDir)
	}

	// replace `{{f}}` in the replacement string with the original
	// filename (without the extension)
	if filenameRegex.MatchString(str) {
		str = filenameRegex.ReplaceAllString(
			str,
			filenameWithoutExtension(fileName),
		)
	}

	// replace `{{ext}}` in the replacement string with the file extension
	if extensionRegex.MatchString(str) {
		str = extensionRegex.ReplaceAllString(str, fileExt)
	}

	// replace `{{p}}` in the replacement string with the parent directory name
	if parentDirRegex.MatchString(str) {
		str = parentDirRegex.ReplaceAllString(str, parentDir)
	}

	// handle date variables (e.g {{mtime.DD}})
	if dateRegex.MatchString(str) {
		out, err := replaceDateVariables(source, str, vars.date)
		if err != nil {
			return "", err
		}
		str = out
	}

	if exifRegex.MatchString(str) {
		exifData, err := getExifData(source)
		if err != nil {
			return "", err
		}

		out, err := replaceExifVariables(exifData, str, vars.exif)
		if err != nil {
			return "", err
		}
		str = out
	}

	if id3Regex.MatchString(str) {
		tags, err := getID3Tags(source)
		if err != nil {
			return "", err
		}

		str = replaceID3Variables(tags, str, vars.id3)
	}

	if hashRegex.MatchString(str) {
		out, err := replaceFileHash(source, str, vars.hash)
		if err != nil {
			return "", err
		}
		str = out
	}

	if randomRegex.Match([]byte(str)) {
		str = replaceRandomVariables(str, vars.random)
	}

	return str, nil
}
