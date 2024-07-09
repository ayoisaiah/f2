// Package sort is used to sort file changes in a variety of ways
// Alphabetical order is the default
package sortfiles

import (
	"io/fs"
	"os"
	"sort"
	"time"

	"gopkg.in/djherbis/times.v1"

	"github.com/MagicalTux/natsort"

	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/timeutil"
)

// FilesBeforeDirs is used to sort files before directories to avoid renaming
// conflicts. It also ensures that child directories are renamed before their
// parents and vice versa in undo mode.
func FilesBeforeDirs(changes []*file.Change, revert bool) []*file.Change {
	sort.SliceStable(changes, func(i, j int) bool {
		compareElement1 := changes[i]
		compareElement2 := changes[j]

		// sort parent directories before child directories in revert mode
		if revert {
			return len(compareElement1.BaseDir) < len(compareElement2.BaseDir)
		}

		// sort files before directories
		if !compareElement1.IsDir {
			return true
		}

		// sort child directories before parent directories
		return len(compareElement1.BaseDir) > len(compareElement2.BaseDir)
	})

	return changes
}

// EnforceHierarchicalOrder ensures all files in the same directory are sorted before
// children directories.
func EnforceHierarchicalOrder(changes []*file.Change) []*file.Change {
	sort.SliceStable(changes, func(i, j int) bool {
		compareElement1 := changes[i]
		compareElement2 := changes[j]

		return len(compareElement1.BaseDir) < len(compareElement2.BaseDir)
	})

	return changes
}

// ByTime sorts the changes by the specified file timing attribute
// (modified time, access time, change time, or birth time).
func ByTime(
	changes []*file.Change,
	sortName string,
	reverseSort bool,
) ([]*file.Change, error) {
	var err error

	sort.SliceStable(changes, func(i, j int) bool {
		compareElement1Path := changes[i].RelSourcePath
		compareElement2Path := changes[j].RelSourcePath

		var compareElement1, compareElement2 times.Timespec
		compareElement1, err = times.Stat(compareElement1Path)
		compareElement2, err = times.Stat(compareElement2Path)

		var itime, jtime time.Time

		switch sortName {
		case timeutil.Mod:
			itime = compareElement1.ModTime()
			jtime = compareElement2.ModTime()
		case timeutil.Birth:
			itime = compareElement1.ModTime()
			jtime = compareElement2.ModTime()

			if compareElement1.HasBirthTime() {
				itime = compareElement1.BirthTime()
			}

			if compareElement2.HasBirthTime() {
				jtime = compareElement2.BirthTime()
			}
		case timeutil.Access:
			itime = compareElement1.AccessTime()
			jtime = compareElement2.AccessTime()
		case timeutil.Change:
			itime = compareElement1.ModTime()
			jtime = compareElement2.ModTime()

			if compareElement1.HasChangeTime() {
				itime = compareElement1.ChangeTime()
			}

			if compareElement2.HasChangeTime() {
				jtime = compareElement2.ChangeTime()
			}
		}

		it, jt := itime.UnixNano(), jtime.UnixNano()

		if reverseSort {
			return it < jt
		}

		return it > jt
	})

	return changes, err
}

// BySize sorts the changes according to their file size.
func BySize(changes []*file.Change, reverseSort bool) ([]*file.Change, error) {
	var err error

	sort.SliceStable(changes, func(i, j int) bool {
		compareElement1Path := changes[i].RelSourcePath
		compareElement2Path := changes[j].RelSourcePath

		var compareElement1, compareElement2 fs.FileInfo
		compareElement1, err = os.Stat(compareElement1Path)
		compareElement2, err = os.Stat(compareElement2Path)

		isize := compareElement1.Size()
		jsize := compareElement2.Size()

		if reverseSort {
			return isize > jsize
		}

		return isize < jsize
	})

	return changes, err
}

// Natural sorts the changes according to natural order (meaning numbers are
// interpreted naturally).
func Natural(changes []*file.Change, reverseSort bool) ([]*file.Change, error) {
	sort.SliceStable(changes, func(i, j int) bool {
		compareElement1 := changes[i].RelSourcePath
		compareElement2 := changes[j].RelSourcePath

		if reverseSort {
			return !natsort.Compare(compareElement1, compareElement2)
		}

		return natsort.Compare(compareElement1, compareElement2)
	})

	return changes, nil
}

// Changes is used to sort changes according to the configured sort value.
func Changes(
	changes []*file.Change,
	sortName string,
	reverseSort bool,
) ([]*file.Change, error) {
	switch sortName {
	case "natural":
		return Natural(changes, reverseSort)
	case "size":
		return BySize(changes, reverseSort)
	case timeutil.Mod,
		timeutil.Access,
		timeutil.Birth,
		timeutil.Change:
		return ByTime(changes, sortName, reverseSort)
	}

	return changes, nil
}
