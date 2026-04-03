package variables

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
)

// Exif represents exif information from an image file.
type Exif struct {
	Latitude              string
	DateTimeOriginal      string
	Make                  string
	Model                 string
	Longitude             string
	Software              string
	LensModel             string
	ImageLength           []int
	ImageWidth            []int
	FNumber               []string
	FocalLength           []string
	FocalLengthIn35mmFilm []int
	PixelYDimension       []int
	PixelXDimension       []int
	ExposureTime          []string
	ISOSpeedRatings       []int
}

// Replace replaces the exif variables in an input string
// if an error occurs while attempting to get the value represented
// by the variables, it is replaced with an empty string.
func (ev exifVars) Replace(
	_ context.Context,
	conf *config.Config,
	change *file.Change,
) error {
	if len(ev.matches) == 0 {
		return nil
	}

	target, err := replaceExifVars(conf, change.Target, change.SourcePath, ev)
	if err != nil {
		return err
	}

	change.Target = target

	return nil
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
func getExifDate(exifData *Exif, token string) string {
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

	return replaceDateToken(dateTime, token)
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
	var width, height string
	if len(exifData.ImageWidth) > 0 {
		width = strconv.Itoa(exifData.ImageWidth[0])
	} else if len(exifData.PixelXDimension) > 0 {
		width = strconv.Itoa(exifData.PixelXDimension[0])
	}

	if len(exifData.ImageLength) > 0 {
		height = strconv.Itoa(exifData.ImageLength[0])
	} else if len(exifData.PixelYDimension) > 0 {
		height = strconv.Itoa(exifData.PixelYDimension[0])
	}

	switch dimension {
	case "w":
		return width
	case "h":
		return height
	case "wh":
		return width + "x" + height
	}

	return ""
}

// replaceExifVars replaces the exif variables in an input string
// if an error occurs while attempting to get the value represented
// by the variables, it is replaced with an empty string.
func replaceExifVars(
	conf *config.Config,
	target, sourcePath string,
	ev exifVars,
) (string, error) {
	exifData, err := getExifData(sourcePath)
	if err != nil {
		return target, err
	}

	for i := range ev.matches {
		current := ev.matches[i]
		regex := current.regex

		var exifTag string

		switch current.attr {
		case "cdt":
			exifTag = getExifDate(exifData, current.timeStr)
		case "soft":
			exifTag = exifData.Software
		case "model":
			exifTag = exifData.Model
		case "lens":
			exifTag = exifData.LensModel
		case "make":
			exifTag = exifData.Make
		case "iso":
			if len(exifData.ISOSpeedRatings) > 0 {
				exifTag = strconv.Itoa(exifData.ISOSpeedRatings[0])
			}
		case "et":
			if len(exifData.ExposureTime) > 0 {
				exifTag = getExifExposureTime(exifData)
			}
		case "fnum":
			if len(exifData.FNumber) > 0 {
				exifTag = getDecimalFromFraction(exifData.FNumber)
			}
		case "fl":
			if len(exifData.FocalLength) > 0 {
				exifTag = getDecimalFromFraction(exifData.FocalLength)
			}
		case "fl35":
			if len(exifData.FocalLengthIn35mmFilm) > 0 {
				exifTag = strconv.Itoa(exifData.FocalLengthIn35mmFilm[0])
			}
		case "lat":
			exifTag = exifData.Latitude
		case "lon":
			exifTag = exifData.Longitude
		case "wh", "h", "w":
			exifTag = getExifDimensions(exifData, current.attr)
		}

		exifTag = transformString(
			conf,
			replaceSlashes(exifTag),
			current.transformToken,
		)

		target = RegexReplace(regex, target, exifTag, 0, nil)
	}

	return target, nil
}
