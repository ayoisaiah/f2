// Package sort is used to sort file changes in a variety of ways
// Alphabetical order is the default
package sortfiles

import (
	"cmp"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"gopkg.in/djherbis/times.v1"

	"github.com/maruel/natural"
	"github.com/pterm/pterm"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
	"github.com/ayoisaiah/f2/v2/internal/pathutil"
)

func isPair(prev, curr *file.Change) bool {
	return pathutil.StripExtension(
		prev.SourcePath,
	) == pathutil.StripExtension(
		curr.SourcePath,
	)
}

// Pairs sorts the given file changes based on a custom pairing order.
// Files with extensions matching earlier entries in pairOrder are sorted
// before those matching later entries.
func Pairs(changes file.Changes, pairOrder []string) {
	slices.SortStableFunc(changes, func(a, b *file.Change) int {
		// Compare stripped paths
		if result := strings.Compare(
			pathutil.StripExtension(a.SourcePath),
			pathutil.StripExtension(b.SourcePath),
		); result != 0 {
			return result
		}

		// Compare extensions based on pairOrder
		aExt, bExt := filepath.Ext(a.Source), filepath.Ext(b.Source)

		for _, v := range pairOrder {
			v = "." + v

			switch {
			case strings.EqualFold(aExt, v):
				return -1
			case strings.EqualFold(bExt, v):
				return 1
			}
		}

		return 0
	})

	for i, v := range changes {
		if i > 0 && i < len(changes) {
			prev := changes[i-1]

			if isPair(prev, v) {
				if prev.PrimaryPair != nil {
					v.PrimaryPair = prev.PrimaryPair
				} else {
					v.PrimaryPair = prev
				}
			}
		}
	}
}

// ForRenamingAndUndo is used to sort files before directories to avoid renaming
// conflicts, and child directories before their parents. Undo mode reverses
// this: a directory is restored before anything inside it.
func ForRenamingAndUndo(changes file.Changes, revert bool) {
	slices.SortStableFunc(changes, func(a, b *file.Change) int {
		// in revert mode, restore parents before what is inside them, so that
		// the paths of the files and directories inside exist again
		if revert {
			if c := cmp.Compare(len(a.BaseDir), len(b.BaseDir)); c != 0 {
				return c
			}

			// at the same depth, files before directories as when renaming
			return cmp.Compare(boolToInt(a.IsDir), boolToInt(b.IsDir))
		}

		// sort files before directories
		if !a.IsDir && b.IsDir {
			return -1
		}

		// sort child directories before parent directories
		return cmp.Compare(len(b.BaseDir), len(a.BaseDir))
	})
}

// Hierarchically ensures all files in the same directory are sorted
// before children directories.
func Hierarchically(changes file.Changes) {
	slices.SortStableFunc(changes, func(a, b *file.Change) int {
		lenA, lenB := len(a.BaseDir), len(b.BaseDir)
		if lenA == lenB {
			return 0
		}

		return cmp.Compare(lenA, lenB)
	})
}

// ByTime sorts the changes by the specified file timing attribute
// (modified time, access time, change time, or birth time).
func ByTime(
	changes file.Changes,
	conf *config.Config,
) {
	for _, change := range changes {
		sourcePath := change.SourcePath
		if change.PrimaryPair != nil {
			sourcePath = change.PrimaryPair.SourcePath
		}

		source, err := times.Stat(sourcePath)
		if err != nil {
			pterm.Error.Printfln("error getting file times info: %v", err)
			os.Exit(1)
		}

		fileTime := source.ModTime()

		//nolint:exhaustive // considering time sorts alone
		switch conf.Sort {
		case config.SortMtime:
		case config.SortBtime:
			if source.HasBirthTime() {
				fileTime = source.BirthTime()
			}
		case config.SortAtime:
			fileTime = source.AccessTime()
		case config.SortCtime:
			if source.HasChangeTime() {
				fileTime = source.ChangeTime()
			}
		}

		change.SortCriterion.Time = fileTime
	}

	slices.SortStableFunc(changes, func(a, b *file.Change) int {
		aTime := a.SortCriterion.Time
		bTime := b.SortCriterion.Time

		if conf.SortPerDir && a.BaseDir != b.BaseDir {
			return 0
		}

		if conf.ReverseSort {
			return -cmp.Compare(aTime.UnixNano(), bTime.UnixNano())
		}

		return cmp.Compare(aTime.UnixNano(), bTime.UnixNano())
	})
}

// BySize sorts the file changes in place based on their file size, either in
// ascending or descending order depending on the `reverseSort` flag.
func BySize(changes file.Changes, conf *config.Config) {
	for _, change := range changes {
		sourcePath := change.SourcePath
		if change.PrimaryPair != nil {
			sourcePath = change.PrimaryPair.SourcePath
		}

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			pterm.Error.Printfln("error getting file info: %v", err)
			os.Exit(1)
		}

		change.SortCriterion.Size = fileInfo.Size()
	}

	slices.SortStableFunc(changes, func(a, b *file.Change) int {
		// Don't sort files in different directories relative to each other
		if conf.SortPerDir && a.BaseDir != b.BaseDir {
			return 0
		}

		if conf.ReverseSort {
			return cmp.Compare(b.SortCriterion.Size, a.SortCriterion.Size)
		}

		return cmp.Compare(a.SortCriterion.Size, b.SortCriterion.Size)
	})
}

// Natural sorts the changes according to natural order (meaning numbers are
// interpreted naturally). However, non-numeric characters are remain sorted in
// ASCII order.
func Natural(changes file.Changes, reverseSort bool) {
	sort.SliceStable(changes, func(i, j int) bool {
		sourcePathA := changes[i].SourcePath
		sourcePathB := changes[j].SourcePath

		if changes[i].PrimaryPair != nil {
			sourcePathA = changes[i].PrimaryPair.SourcePath
		}

		if changes[j].PrimaryPair != nil {
			sourcePathB = changes[j].PrimaryPair.SourcePath
		}

		if reverseSort {
			return natural.Less(sourcePathB, sourcePathA)
		}

		return natural.Less(sourcePathA, sourcePathB)
	})
}

// Changes is used to sort changes according to the configured sort value.
func Changes(
	changes file.Changes,
	conf *config.Config,
) {
	if conf.SortPerDir {
		Hierarchically(changes)
	}

	//nolint:exhaustive // default sort not needed
	switch conf.Sort {
	case config.SortNatural:
		Natural(changes, conf.ReverseSort)
	case config.SortSize:
		BySize(changes, conf)
	case config.SortMtime,
		config.SortAtime,
		config.SortBtime,
		config.SortCtime:
		ByTime(changes, conf)
	case config.SortTimeVar:
		ByTimeVar(changes, conf)
	case config.SortStringVar:
		ByStringVar(changes, conf)
	case config.SortIntVar:
		ByIntVar(changes, conf)
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}

	return 0
}
