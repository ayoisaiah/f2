package f2

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/djherbis/times.v1"
)

// sortMatches is used to sort files to avoid renaming conflicts.
func (op *Operation) sortMatches() {
	sort.SliceStable(op.matches, func(i, j int) bool {
		// sort parent directories before child directories in revert mode
		if op.revert {
			return len(op.matches[i].BaseDir) < len(op.matches[j].BaseDir)
		}

		// sort files before directories
		if !op.matches[i].IsDir {
			return true
		}

		// sort child directories before parent directories
		return len(op.matches[i].BaseDir) > len(op.matches[j].BaseDir)
	})
}

// sortBySize sorts the matches according to their file size.
func (op *Operation) sortBySize() (err error) {
	sort.SliceStable(op.matches, func(i, j int) bool {
		ipath := filepath.Join(op.matches[i].BaseDir, op.matches[i].Source)
		jpath := filepath.Join(op.matches[j].BaseDir, op.matches[j].Source)

		var ifile, jfile fs.FileInfo
		ifile, err = os.Stat(ipath)
		jfile, err = os.Stat(jpath)

		isize := ifile.Size()
		jsize := jfile.Size()

		if op.reverseSort {
			return isize > jsize
		}

		return isize < jsize
	})

	return err
}

// sortByTime sorts the matches by the specified file attribute
// (mtime, atime, btime or ctime).
func (op *Operation) sortByTime() (err error) {
	sort.SliceStable(op.matches, func(i, j int) bool {
		ipath := filepath.Join(op.matches[i].BaseDir, op.matches[i].Source)
		jpath := filepath.Join(op.matches[j].BaseDir, op.matches[j].Source)

		var ifile, jfile times.Timespec
		ifile, err = times.Stat(ipath)
		jfile, err = times.Stat(jpath)

		var itime, jtime time.Time
		switch op.sort {
		case modTime:
			itime = ifile.ModTime()
			jtime = jfile.ModTime()
		case birthTime:
			itime = ifile.ModTime()
			jtime = jfile.ModTime()
			if ifile.HasBirthTime() {
				itime = ifile.BirthTime()
			}
			if jfile.HasBirthTime() {
				jtime = jfile.BirthTime()
			}
		case accessTime:
			itime = ifile.AccessTime()
			jtime = jfile.AccessTime()
		case changeTime:
			itime = ifile.ModTime()
			jtime = jfile.ModTime()
			if ifile.HasChangeTime() {
				itime = ifile.ChangeTime()
			}
			if jfile.HasChangeTime() {
				jtime = jfile.ChangeTime()
			}
		}

		it, jt := itime.UnixNano(), jtime.UnixNano()

		if op.reverseSort {
			return it < jt
		}

		return it > jt
	})

	return err
}

// TODO: Needs to be refactored.
func (op *Operation) sortPaths(
	paths map[string][]os.DirEntry,
	sorted bool,
) []Change {
	var sortedPaths []Change

	if sorted {
		type KeyValue struct {
			Key   string
			Value []os.DirEntry
		}

		// create an empty slice of key-value pairs
		dirPaths := make([]KeyValue, 0, len(paths))
		// append all map keys-value pairs to the slice
		for k, v := range paths {
			dirPaths = append(dirPaths, KeyValue{k, v})
		}

		// sort map keys
		sort.SliceStable(dirPaths, func(i, j int) bool {
			if op.reverseSort {
				return strings.ToLower(
					dirPaths[i].Key,
				) > strings.ToLower(
					dirPaths[j].Key,
				)
			}

			return strings.ToLower(
				dirPaths[i].Key,
			) < strings.ToLower(
				dirPaths[j].Key,
			)
		})

		for _, v := range dirPaths {
			dirPath := v.Key
			val := paths[v.Key]

			// sort directory entries
			sort.SliceStable(val, func(i, j int) bool {
				if op.reverseSort {
					return strings.ToLower(
						val[i].Name(),
					) > strings.ToLower(
						val[j].Name(),
					)
				}

				return strings.ToLower(
					val[i].Name(),
				) < strings.ToLower(
					val[j].Name(),
				)
			})

			for _, f := range val {
				var change = Change{
					BaseDir:        dirPath,
					IsDir:          f.IsDir(),
					Source:         filepath.Clean(f.Name()),
					originalSource: filepath.Clean(f.Name()),
				}

				sortedPaths = append(sortedPaths, change)
			}
		}

		return sortedPaths
	}

	for k, v := range paths {
		for _, f := range v {
			var change = Change{
				BaseDir:        k,
				IsDir:          f.IsDir(),
				Source:         filepath.Clean(f.Name()),
				originalSource: filepath.Clean(f.Name()),
			}

			sortedPaths = append(sortedPaths, change)
		}
	}

	return sortedPaths
}

// sortBy delegates the sorting of matches to the appropriate method.
func (op *Operation) sortBy() (err error) {
	switch op.sort {
	case "size":
		return op.sortBySize()
	case accessTime, modTime, birthTime, changeTime:
		return op.sortByTime()
	}

	return nil
}
