package variables

import (
	"context"
	"os"
	"strconv"

	"github.com/dhowden/tag"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
)

// Replace replaces all id3 variables in the target file name with the
// corresponding id3 tag value.
func (id3v id3Vars) Replace(
	_ context.Context,
	conf *config.Config,
	change *file.Change,
) error {
	if len(id3v.matches) == 0 {
		return nil
	}

	target, err := replaceID3Variables(conf, change, id3v)
	if err != nil {
		return err
	}

	change.Target = target

	return nil
}

// getID3Tags retrieves the id3 tags in an audi file (such as mp3)
// errors while reading the id3 tags are ignored since the corresponding
// variable will be replaced with an empty string.
func getID3Tags(sourcePath string) (*file.ID3, error) {
	f, err := os.Open(sourcePath)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	metadata, err := tag.ReadFrom(f)
	if err != nil {
		// empty ID3 instance which means the variables are replaced with empty strings
		return &file.ID3{}, nil
	}

	trackNum, totalTracks := metadata.Track()
	discNum, totalDiscs := metadata.Disc()

	return &file.ID3{
		Format:      string(metadata.Format()),
		FileType:    string(metadata.FileType()),
		Title:       metadata.Title(),
		Album:       metadata.Album(),
		Artist:      metadata.Artist(),
		AlbumArtist: metadata.AlbumArtist(),
		Track:       trackNum,
		TotalTracks: totalTracks,
		Disc:        discNum,
		TotalDiscs:  totalDiscs,
		Composer:    metadata.Composer(),
		Year:        metadata.Year(),
		Genre:       metadata.Genre(),
	}, nil
}

// replaceID3Variables replaces all id3 variables in the target file name
// with the corresponding id3 tag value.
func replaceID3Variables(
	conf *config.Config,
	change *file.Change,
	id3v id3Vars,
) (string, error) {
	target := change.Target

	tags := change.ID3Data
	if tags == nil {
		var err error

		tags, err = getID3Tags(change.SourcePath)
		if err != nil {
			return target, err
		}

		change.ID3Data = tags
	}

	for i := range id3v.matches {
		current := id3v.matches[i]
		submatch := current.tag

		var id3Tag string

		switch submatch {
		case "format":
			id3Tag = tags.Format
		case "type":
			id3Tag = tags.FileType
		case "title":
			id3Tag = tags.Title
		case "album":
			id3Tag = tags.Album
		case "artist":
			id3Tag = tags.Artist
		case "album_artist":
			id3Tag = tags.AlbumArtist
		case "genre":
			id3Tag = tags.Genre
		case "composer":
			id3Tag = tags.Composer
		case "track":
			if tags.Track != 0 {
				id3Tag = strconv.Itoa(tags.Track)
			}

		case "total_tracks":
			if tags.TotalTracks != 0 {
				id3Tag = strconv.Itoa(tags.TotalTracks)
			}

		case "disc":
			if tags.Disc != 0 {
				id3Tag = strconv.Itoa(tags.Disc)
			}
		case "total_discs":
			if tags.TotalDiscs != 0 {
				id3Tag = strconv.Itoa(tags.TotalDiscs)
			}

		case "year":
			if tags.Year != 0 {
				id3Tag = strconv.Itoa(tags.Year)
			}
		}

		id3Tag = transformString(
			conf,
			replaceSlashes(id3Tag),
			current.transformToken,
		)

		target = RegexReplace(current.regex, target, id3Tag, 0, nil)
	}

	return target, nil
}
