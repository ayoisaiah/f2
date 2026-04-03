package find

import (
	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
)

// Filter defines the interface for all file discovery filters.
type Filter interface {
	// Skip returns true if the match should be excluded from renaming.
	Skip(conf *config.Config, match *file.Change) bool
}

// Filters manages a collection of filters.
type Filters []Filter

// Skip runs the match through all registered filters.
// It returns true if any filter decides to skip the match.
func (f Filters) Skip(conf *config.Config, match *file.Change) bool {
	for _, filter := range f {
		if filter.Skip(conf, match) {
			return true
		}
	}

	return false
}

// ExcludeFilter excludes files that match the exclusion regex.
type ExcludeFilter struct{}

func (f ExcludeFilter) Skip(conf *config.Config, match *file.Change) bool {
	return conf.ExcludeRegex != nil &&
		conf.ExcludeRegex.MatchString(match.Source)
}

// IncludeDirFilter excludes directories unless explicitly included.
type IncludeDirFilter struct{}

func (f IncludeDirFilter) Skip(conf *config.Config, match *file.Change) bool {
	return !conf.IncludeDir && match.IsDir
}

// OnlyDirFilter excludes non-directory files if OnlyDir is set.
type OnlyDirFilter struct{}

func (f OnlyDirFilter) Skip(conf *config.Config, match *file.Change) bool {
	return conf.OnlyDir && !match.IsDir
}

// IncludeRegexFilter excludes files that do not match the inclusion regex.
type IncludeRegexFilter struct{}

func (f IncludeRegexFilter) Skip(conf *config.Config, match *file.Change) bool {
	return conf.IncludeRegex != nil &&
		!conf.IncludeRegex.MatchString(match.Source)
}

// NewFilters creates a new collection of filters based on the configuration.
func NewFilters(_ *config.Config) Filters {
	return Filters{
		ExcludeFilter{},
		IncludeDirFilter{},
		OnlyDirFilter{},
		IncludeRegexFilter{},
	}
}
