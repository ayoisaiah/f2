package find

import (
	"encoding/csv"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/samber/lo"

	"github.com/ayoisaiah/f2/internal/config"
	internalpath "github.com/ayoisaiah/f2/internal/path"
)

const (
	dotCharacter = 46
)

// csvRows keeps track of each row in a CSV file so that it can be associated
// with a file renaming change. The key is the absolute path of the source file
// and the value is the correspoding row in the CSV file.
var csvRows = make(map[string][]string)

// readCSVFile reads all the records contained in a CSV file specified by
// `pathToCSV`.
func readCSVFile(pathToCSV string) ([][]string, error) {
	f, err := os.Open(pathToCSV)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	csvReader := csv.NewReader(f)

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

// handleCSV reads the provided CSV file, and finds all the valid candidates
// for replacement.
func handleCSV(conf *config.Config) (internalpath.Collection, error) {
	paths := make(internalpath.Collection)

	records, err := readCSVFile(conf.CSVFilename)
	if err != nil {
		return nil, err
	}

	csvAbsPath, err := filepath.Abs(conf.CSVFilename)
	if err != nil {
		return nil, err
	}

	findSlice := make([]string, 0, len(records))

	replacementSlice := make([]string, 0, len(records))

	for _, record := range records {
		if len(record) == 0 {
			continue
		}

		source := strings.TrimSpace(record[0])

		absSourcePath := filepath.Join(filepath.Dir(csvAbsPath), source)

		fileInfo, err2 := os.Stat(absSourcePath)
		if err2 != nil {
			return nil, err2
		}

		findSlice = append(findSlice, fileInfo.Name())

		sourceDir := filepath.Dir(absSourcePath)

		var dirEntry []fs.DirEntry

		dirEntry, err2 = os.ReadDir(sourceDir)
		if err2 != nil {
			return nil, err2
		}

	entryLoop:
		for _, entry := range dirEntry {
			if entry.Name() == fileInfo.Name() {
				// Ensure that the file is not already
				// present in the directory entry
				for _, e := range paths[sourceDir] {
					if e.Name() == fileInfo.Name() {
						break entryLoop
					}
				}

				paths[sourceDir] = append(paths[sourceDir], entry)

				break
			}
		}

		if len(record) > 1 {
			target := strings.TrimSpace(record[1])

			replacementSlice = append(replacementSlice, target)
		}

		csvRows[absSourcePath] = record
	}

	if len(conf.ReplacementSlice) == 0 {
		if len(conf.FindSlice) == 0 {
			config.SetFindSlice(findSlice)
			config.SetReplacementSlice(replacementSlice)

			err = config.SetFindStringRegex(0)
			if err != nil {
				return nil, err
			}
		}
	}

	return paths, nil
}

// filterMatches filters out files that do not match the search pattern or
// excluded using other arguments.
func filterMatches(
	conf *config.Config,
	pathsToFilter internalpath.Collection,
) (internalpath.Collection, error) {
	matches := make(internalpath.Collection)

	for baseDir := range pathsToFilter {
		dirEntry := pathsToFilter[baseDir]
		filteredDirEntry := dirEntry[:0]

		for i := range dirEntry {
			entry := dirEntry[i]

			fileName := entry.Name()

			entryIsDir := entry.IsDir()

			if conf.IgnoreExt && !entryIsDir {
				fileName = internalpath.StripExtension(fileName)
			}

			matched := conf.SearchRegex.MatchString(fileName)
			if !matched {
				continue
			}

			// Ensure full name is used for the other checks
			fileName = entry.Name()

			if conf.ExcludeRegex != nil &&
				conf.ExcludeRegex.MatchString(fileName) {
				continue
			}

			if entryIsDir && !conf.IncludeDir {
				continue
			}

			if conf.OnlyDir && !entryIsDir {
				continue
			}

			if !conf.IncludeHidden {
				isHidden, err := checkIfHidden(fileName, baseDir)
				if err != nil {
					return nil, err
				}

				// Ensure that explicitly included file arguments are not affected
				if isHidden {
					entryAbsPath, err := filepath.Abs(
						filepath.Join(baseDir, fileName),
					)
					if err != nil {
						return nil, err
					}

					shouldSkip := true

					for _, pathArg := range conf.FilesAndDirPaths {
						argAbsPath, err := filepath.Abs(pathArg)
						if err != nil {
							return nil, err
						}

						if strings.EqualFold(entryAbsPath, argAbsPath) {
							shouldSkip = false
						}
					}

					if shouldSkip {
						continue
					}
				}
			}

			filteredDirEntry = append(filteredDirEntry, entry)

			matches[baseDir] = filteredDirEntry
		}
	}

	return matches, nil
}

// traverseDirs walks through the specified directories and collects their
// contents in a map until the specified max depth is reached or there are no
// more directories to recurse into. Hidden directories are ignored if specified.
func traverseDirs(
	dirAndContents internalpath.Collection,
	maxDepth int,
	includeHidden bool,
) (internalpath.Collection, error) {
	all := []map[string][]os.DirEntry{
		dirAndContents,
	}

	currentDepth := 0

	for {
		nextLevel := make(internalpath.Collection)

		for baseDir, dirContents := range all[currentDepth] {
			for i := range dirContents {
				entry := dirContents[i]

				if !entry.IsDir() {
					continue
				}

				dirName := entry.Name()

				if !includeHidden {
					dirIsHidden, err := checkIfHidden(dirName, baseDir)
					if err != nil {
						return nil, err
					}

					if dirIsHidden {
						continue
					}
				}

				entryPath := filepath.Join(baseDir, dirName)

				dirEntry, err := os.ReadDir(entryPath)
				if err != nil {
					return nil, err
				}

				nextLevel[entryPath] = dirEntry
			}
		}

		if len(nextLevel) != 0 {
			currentDepth++

			all = append(all, nextLevel)

			if maxDepth == 0 || maxDepth != currentDepth {
				continue
			}
		}

		break
	}

	return lo.Assign(all...), nil
}

// searchPaths groups the directories that will be searched for matches and their
// contents.
func searchPaths(conf *config.Config) (internalpath.Collection, error) {
	dirAndContents := make(internalpath.Collection)

	for _, path := range conf.FilesAndDirPaths {
		var fileInfo os.FileInfo

		path = filepath.Clean(path)

		// Skip paths that have already been processed
		if _, ok := dirAndContents[path]; ok {
			continue
		}

		fileInfo, err := os.Stat(path)
		if err != nil {
			return nil, err
		}

		if fileInfo.IsDir() {
			dirAndContents[path], err = os.ReadDir(path)
			if err != nil {
				return nil, err
			}

			continue
		}

		// Add the specified file path to its parent directory's contents only if
		// it does not exist already to avoid duplicates
		dir := filepath.Dir(path)

		dirEntry := fs.FileInfoToDirEntry(fileInfo)

		dirContainsFile := slices.ContainsFunc(
			dirAndContents[dir],
			func(d fs.DirEntry) bool {
				return d.Name() == dirEntry.Name()
			},
		)

		if !dirContainsFile {
			dirAndContents[dir] = append(dirAndContents[dir], dirEntry)
		}
	}

	if conf.Recursive {
		d, err := traverseDirs(
			dirAndContents,
			conf.MaxDepth,
			conf.IncludeHidden,
		)
		if err != nil {
			return nil, err
		}

		dirAndContents = d
	}

	return dirAndContents, nil
}

// Find returns a collection of files and directories that match the search
// pattern or explicitly included as command-line arguments.
func Find(conf *config.Config) (internalpath.Collection, error) {
	if conf.CSVFilename != "" {
		return handleCSV(conf)
	}

	paths, err := searchPaths(conf)
	if err != nil {
		return nil, err
	}

	matches, err := filterMatches(conf, paths)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func GetCSVRows() map[string][]string {
	return csvRows
}
