package variables

import (
	"context"
	"log/slog"
	"strings"

	"github.com/barasher/go-exiftool"

	"github.com/ayoisaiah/f2/v2/internal/apperr"
	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
	"github.com/ayoisaiah/f2/v2/internal/localize"
)

var errExiftoolInit = &apperr.Error{
	Message: localize.T("error.exiftool_init"),
}

// Replace replaces the all exiftool variables in the target.
func (xtVars exiftoolVars) Replace(
	_ context.Context,
	conf *config.Config,
	change *file.Change,
) error {
	if len(xtVars.matches) == 0 {
		return nil
	}

	target, err := replaceExifToolVars(conf, change, xtVars)
	if err != nil {
		return err
	}

	change.Target = target

	return nil
}

// replaceExifToolVars replaces the all exiftool variables in the target.
func replaceExifToolVars(
	conf *config.Config,
	change *file.Change,
	xtVars exiftoolVars,
) (string, error) {
	target := change.Target

	var fileMeta []exiftool.FileMetadata

	if change.ExiftoolData == nil {
		var err error

		fileMeta, err = ExtractExiftoolMetadata(conf, change.SourcePath)
		if err != nil {
			return "", err
		}
	} else {
		fileMeta = append(fileMeta, *change.ExiftoolData)
	}

	for i := range xtVars.matches {
		current := xtVars.matches[i]

		var value string

		for _, meta := range fileMeta {
			if meta.Err != nil {
				continue
			}

			if change.ExiftoolData == nil {
				change.ExiftoolData = &meta
			}

			v, err := meta.GetString(current.attr)
			if err != nil {
				continue
			}

			if current.attr == "DateTimeOriginal" &&
				strings.HasPrefix(current.transformToken, "dt") {
				o, err := meta.GetString("OffsetTimeOriginal")
				if err != nil {
					slog.Debug(
						"could not retrieve OffsetTimeOriginal, assuming UTC time",
						slog.Any("match", change),
					)
				}

				v += o
			}

			value = replaceSlashes(v)
		}

		value = transformString(conf, value, current.transformToken)

		target = RegexReplace(current.regex, target, value, 0, nil)
	}

	return target, nil
}

func ExtractExiftoolMetadata(
	conf *config.Config,
	fileNames ...string,
) ([]exiftool.FileMetadata, error) {
	var opts []func(*exiftool.Exiftool) error

	if conf.ExiftoolOpts.API != "" {
		opts = append(opts, exiftool.Api(conf.ExiftoolOpts.API))
	}

	if conf.ExiftoolOpts.Charset != "" {
		opts = append(opts, exiftool.Charset(conf.ExiftoolOpts.Charset))
	}

	if conf.ExiftoolOpts.CoordFormat != "" {
		opts = append(
			opts,
			exiftool.CoordFormant(conf.ExiftoolOpts.CoordFormat),
		)
	}

	if conf.ExiftoolOpts.DateFormat != "" {
		opts = append(
			opts,
			exiftool.DateFormant(conf.ExiftoolOpts.DateFormat),
		)
	}

	if conf.ExiftoolOpts.ExtractEmbedded {
		opts = append(opts, exiftool.ExtractEmbedded())
	}

	et, err := exiftool.NewExiftool(opts...)
	if err != nil {
		return nil, errExiftoolInit.Wrap(err)
	}

	defer et.Close()

	return et.ExtractMetadata(fileNames...), nil
}
