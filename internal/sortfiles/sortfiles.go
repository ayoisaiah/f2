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

	"gopkg.in/djherbis/times.v1"

	"github.com/MagicalTux/natsort"
	"github.com/pterm/pterm"

	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/timeutil"
)

// ForRenamingAndUndo is used to sort files before directories to avoid renaming
// conflicts. It also ensures that child directories are renamed before their
// parents and vice versa in undo mode.
func ForRenamingAndUndo(changes []*file.Change, revert bool) {
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

// EnforceHierarchicalOrder ensures all files in the same directory are sorted
// before children directories.
func EnforceHierarchicalOrder(changes []*file.Change) {
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
	changes []*file.Change,
	sortName string,
	reverseSort bool,
	sortPerDir bool,
) {
	slices.SortStableFunc(changes, func(a, b *file.Change) int {
		sourceA, errA := times.Stat(a.RelSourcePath)
		sourceB, errB := times.Stat(b.RelSourcePath)

		if errA != nil || errB != nil {
			pterm.Error.Printfln(
				"error getting file times info: %v, %v",
				errA,
				errB,
			)
			os.Exit(1)
		}

		aTime, bTime := sourceA.ModTime(), sourceB.ModTime()

		switch sortName {
		case timeutil.Mod:
		case timeutil.Birth:
			if sourceA.HasBirthTime() {
				aTime = sourceA.BirthTime()
			}

			if sourceB.HasBirthTime() {
				bTime = sourceB.BirthTime()
			}
		case timeutil.Access:
			aTime = sourceA.AccessTime()
			bTime = sourceB.AccessTime()
		case timeutil.Change:
			if sourceA.HasChangeTime() {
				aTime = sourceA.ChangeTime()
			}

			if sourceB.HasChangeTime() {
				bTime = sourceB.ChangeTime()
			}
		}

		if sortPerDir &&
			filepath.Dir(a.RelSourcePath) != filepath.Dir(b.RelSourcePath) {
			return 0
		}

		if reverseSort {
			return -cmp.Compare(aTime.UnixNano(), bTime.UnixNano())
		}

		return cmp.Compare(aTime.UnixNano(), bTime.UnixNano())
	})
}

// BySize sorts the file changes in place based on their file size, either in
// ascending or descending order depending on the `reverseSort` flag.
func BySize(changes []*file.Change, reverseSort, sortPerDir bool) {
	slices.SortStableFunc(changes, func(a, b *file.Change) int {
		var fileInfoA, fileInfoB fs.FileInfo
		fileInfoA, errA := os.Stat(a.RelSourcePath)
		fileInfoB, errB := os.Stat(b.RelSourcePath)

		if errA != nil || errB != nil {
			pterm.Error.Printfln("error getting file info: %v, %v", errA, errB)
			os.Exit(1)
		}

		fileASize := fileInfoA.Size()
		fileBSize := fileInfoB.Size()

		// Don't sort files in different directories relative to each other
		if sortPerDir &&
			filepath.Dir(a.RelSourcePath) != filepath.Dir(b.RelSourcePath) {
			return 0
		}

		if reverseSort {
			return int(fileBSize - fileASize)
		}

		return int(fileASize - fileBSize)
	})
}

// Natural sorts the changes according to natural order (meaning numbers are
// interpreted naturally). However, non-numeric characters are remain sorted in
// ASCII order.
func Natural(changes []*file.Change, reverseSort bool) {
	sort.SliceStable(changes, func(i, j int) bool {
		sourceA := changes[i].RelSourcePath
		sourceB := changes[j].RelSourcePath

		if reverseSort {
			return !natsort.Compare(sourceA, sourceB)
		}

		return natsort.Compare(sourceA, sourceB)
	})
}

// Changes is used to sort changes according to the configured sort value.
func Changes(
	changes []*file.Change,
	sortName string,
	reverseSort bool,
	sortPerDir bool,
) {
	// TODO: EnforceHierarchicalOrder should be the default sort
	if sortPerDir {
		EnforceHierarchicalOrder(changes)
	}

	switch sortName {
	case "natural":
		Natural(changes, reverseSort)
	case "size":
		BySize(changes, reverseSort, sortPerDir)
	case timeutil.Mod,
		timeutil.Access,
		timeutil.Birth,
		timeutil.Change:
		ByTime(changes, sortName, reverseSort, sortPerDir)
	}
}
