// Package sort is used to sort file changes in a variety of ways
// Alphabetical order is the default
package sort

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/djherbis/times.v1"

	"github.com/ayoisaiah/f2/internal/file"
	internaltime "github.com/ayoisaiah/f2/internal/time"
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

// ByTime sorts the changes by the specified file timing attribute
// (modified time, access time, change time, or birth time).
func ByTime(
	changes []*file.Change,
	sortName string,
	reverseSort bool,
) ([]*file.Change, error) {
	var err error

	sort.SliceStable(changes, func(i, j int) bool {
		compareElement1Path := filepath.Join(
			changes[i].BaseDir,
			changes[i].Source,
		)
		compareElement2Path := filepath.Join(
			changes[j].BaseDir,
			changes[j].Source,
		)

		var compareElement1, compareElement2 times.Timespec
		compareElement1, err = times.Stat(compareElement1Path)
		compareElement2, err = times.Stat(compareElement2Path)

		var itime, jtime time.Time
		switch sortName {
		case internaltime.Mod:
			itime = compareElement1.ModTime()
			jtime = compareElement2.ModTime()
		case internaltime.Birth:
			itime = compareElement1.ModTime()
			jtime = compareElement2.ModTime()
			if compareElement1.HasBirthTime() {
				itime = compareElement1.BirthTime()
			}
			if compareElement2.HasBirthTime() {
				jtime = compareElement2.BirthTime()
			}
		case internaltime.Access:
			itime = compareElement1.AccessTime()
			jtime = compareElement2.AccessTime()
		case internaltime.Change:
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
		compareElement1Path := filepath.Join(
			changes[i].BaseDir,
			changes[i].Source,
		)
		compareElement2Path := filepath.Join(
			changes[j].BaseDir,
			changes[j].Source,
		)

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

// Alphabetically sorts the changes in alphabetical order.
func Alphabetically(changes []*file.Change, reverseSort bool) []*file.Change {
	sort.SliceStable(changes, func(i, j int) bool {
		compareElement1 := strings.ToLower(changes[i].Source)
		compareElement2 := strings.ToLower(changes[j].Source)
		if reverseSort {
			return compareElement1 > compareElement2
		}

		return compareElement1 < compareElement2
	})

	return changes
}

// Changes is used to sort changes according to the configured sort value.
func Changes(
	changes []*file.Change,
	sortName string,
	reverseSort bool,
) ([]*file.Change, error) {
	switch sortName {
	case "size":
		return BySize(changes, reverseSort)
	case internaltime.Mod,
		internaltime.Access,
		internaltime.Birth,
		internaltime.Change:
		return ByTime(changes, sortName, reverseSort)
	}

	return Alphabetically(changes, reverseSort), nil
}
