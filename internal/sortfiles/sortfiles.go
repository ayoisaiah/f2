// Package sort is used to sort file changes in a variety of ways
// Alphabetical order is the default
package sortfiles

import (
	"cmp"
	"io/fs"
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
// conflicts. It also ensures that child directories are renamed before their
// parents and vice versa in undo mode.
func ForRenamingAndUndo(changes file.Changes, revert bool) {
	slices.SortStableFunc(changes, func(a, b *file.Change) int {
		// sort files before directories
		if !a.IsDir && b.IsDir {
			return -1
		}

		// sort parent directories before child directories in revert mode
		if revert {
			return cmp.Compare(len(a.BaseDir), len(b.BaseDir))
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
	slices.SortStableFunc(changes, func(a, b *file.Change) int {
		sourcePathA, sourcePathB := a.SourcePath, b.SourcePath

		if a.PrimaryPair != nil {
			sourcePathA = a.PrimaryPair.SourcePath
		}

		if b.PrimaryPair != nil {
			sourcePathB = b.PrimaryPair.SourcePath
		}

		sourceA, errA := times.Stat(sourcePathA)
		sourceB, errB := times.Stat(sourcePathB)

		if errA != nil || errB != nil {
			pterm.Error.Printfln(
				"error getting file times info: %v, %v",
				errA,
				errB,
			)
			os.Exit(1)
		}

		aTime, bTime := sourceA.ModTime(), sourceB.ModTime()

		//nolint:exhaustive // considering time sorts alone
		switch conf.Sort {
		case config.SortMtime:
		case config.SortBtime:
			if sourceA.HasBirthTime() {
				aTime = sourceA.BirthTime()
			}

			if sourceB.HasBirthTime() {
				bTime = sourceB.BirthTime()
			}
		case config.SortAtime:
			aTime = sourceA.AccessTime()
			bTime = sourceB.AccessTime()
		case config.SortCtime:
			if sourceA.HasChangeTime() {
				aTime = sourceA.ChangeTime()
			}

			if sourceB.HasChangeTime() {
				bTime = sourceB.ChangeTime()
			}
		}

		a.SortCriterion.Time = aTime
		b.SortCriterion.Time = bTime

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
	slices.SortStableFunc(changes, func(a, b *file.Change) int {
		sourcePathA, sourcePathB := a.SourcePath, b.SourcePath

		if a.PrimaryPair != nil {
			sourcePathA = a.PrimaryPair.SourcePath
		}

		if b.PrimaryPair != nil {
			sourcePathB = b.PrimaryPair.SourcePath
		}

		var fileInfoA, fileInfoB fs.FileInfo
		fileInfoA, errA := os.Stat(sourcePathA)
		fileInfoB, errB := os.Stat(sourcePathB)

		if errA != nil || errB != nil {
			pterm.Error.Printfln("error getting file info: %v, %v", errA, errB)
			os.Exit(1)
		}

		fileASize := fileInfoA.Size()
		fileBSize := fileInfoB.Size()

		a.SortCriterion.Size = fileASize
		b.SortCriterion.Size = fileBSize

		// Don't sort files in different directories relative to each other
		if conf.SortPerDir && a.BaseDir != b.BaseDir {
			return 0
		}

		if conf.ReverseSort {
			return int(fileBSize - fileASize)
		}

		return int(fileASize - fileBSize)
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
			return !natural.Less(sourcePathA, sourcePathB)
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
