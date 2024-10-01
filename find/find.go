package find

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"

	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/pathutil"
	"github.com/ayoisaiah/f2/internal/sortfiles"
	"github.com/ayoisaiah/f2/internal/status"
)

const (
	dotCharacter = 46
)

// shouldFilter decides whether a match should be included in the final
// pool of files for renaming.
func shouldFilter(conf *config.Config, match *file.Change) bool {
	if conf.ExcludeRegex != nil &&
		conf.ExcludeRegex.MatchString(match.Source) {
		return true
	}

	if !conf.IncludeDir && match.IsDir {
		return true
	}

	if conf.OnlyDir && !match.IsDir {
		return true
	}

	return false
}

// skipFileIfHidden checks if a file is hidden, and if so, returns a boolean
// confirming whether it should be skipped or not.
func skipFileIfHidden(
	path string,
	filesAndDirPaths []string,
	includeHidden bool,
) (bool, error) {
	if includeHidden {
		return false, nil // No need to check if we're including hidden files
	}

	isHidden, err := checkIfHidden(filepath.Base(path), filepath.Dir(path))
	if err != nil {
		return false, err
	}

	if !isHidden {
		return false, nil // No need to check further if the file isn't hidden
	}

	entryAbsPath, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}

	// Ensure that file path arguments are included regardless of hidden status
	for _, pathArg := range filesAndDirPaths {
		argAbsPath, err := filepath.Abs(pathArg)
		if err != nil {
			return false, err
		}

		if strings.EqualFold(entryAbsPath, argAbsPath) {
			return false, nil
		}
	}

	return true, nil // Skip the hidden file
}

// isMaxDepth reports whether the configured max depth has been reached.
func isMaxDepth(rootPath, currentPath string, maxDepth int) bool {
	if rootPath == filepath.Dir(currentPath) || maxDepth == 0 {
		return false
	}

	if maxDepth == -1 {
		return true
	}

	relativePath := strings.TrimPrefix(
		currentPath,
		rootPath+string(os.PathSeparator),
	)
	depthCount := strings.Count(relativePath, string(os.PathSeparator))

	return depthCount > maxDepth
}

func createFileChange(dirPath string, fileInfo fs.FileInfo) *file.Change {
	baseDir := filepath.Dir(dirPath)
	fileName := fileInfo.Name()

	match := &file.Change{
		BaseDir:      baseDir,
		IsDir:        fileInfo.IsDir(),
		Source:       fileName,
		OriginalName: fileName,
		SourcePath:   filepath.Join(baseDir, fileName),
	}

	return match
}

// searchPaths walks through the filesystem and finds matches for the provided
// search pattern.
func searchPaths(conf *config.Config) (file.Changes, error) {
	processedPaths := make(map[string]bool)

	var matches file.Changes

	for _, rootPath := range conf.FilesAndDirPaths {
		rootPath = filepath.Clean(rootPath)

		fileInfo, err := os.Stat(rootPath)
		if err != nil {
			return nil, err
		}

		if !fileInfo.IsDir() {
			if processedPaths[rootPath] {
				continue
			}

			if conf.Search.Regex.MatchString(fileInfo.Name()) {
				match := createFileChange(rootPath, fileInfo)

				if !shouldFilter(conf, match) {
					matches = append(matches, match)
				}
			}

			processedPaths[rootPath] = true

			continue
		}

		maxDepth := -1 // default value for non-recursive iterations
		if conf.Recursive {
			maxDepth = conf.MaxDepth
		}

		err = filepath.WalkDir(
			rootPath,
			func(currentPath string, entry fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				// skip the root path and already processed paths
				if rootPath == currentPath || processedPaths[currentPath] {
					return nil
				}

				if skipHidden, hiddenErr := skipFileIfHidden(
					currentPath,
					conf.FilesAndDirPaths,
					conf.IncludeHidden,
				); hiddenErr != nil {
					return hiddenErr
				} else if skipHidden {
					if entry.IsDir() {
						return fs.SkipDir
					}

					return nil
				}

				if entry.IsDir() && conf.Recursive &&
					conf.ExcludeDirRegex != nil {
					if conf.ExcludeDirRegex.MatchString(entry.Name()) {
						return fs.SkipDir
					}
				}

				if isMaxDepth(rootPath, currentPath, maxDepth) {
					return fs.SkipDir
				}

				fileName := entry.Name()

				entryIsDir := entry.IsDir()

				if conf.IgnoreExt && !entryIsDir {
					fileName = pathutil.StripExtension(fileName)
				}

				if conf.Search.Regex.MatchString(fileName) {
					fileInfo, infoErr := entry.Info()
					if infoErr != nil {
						return infoErr
					}

					match := createFileChange(currentPath, fileInfo)

					if !shouldFilter(conf, match) {
						matches = append(matches, match)
					}
				}

				processedPaths[currentPath] = true

				return nil
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return matches, nil
}

// loadFromBackup loads the details of the previous renaming operation
// from the backup file. It returns the changes or an error if the backup file
// cannot be found or parsed.
func loadFromBackup(conf *config.Config) (file.Changes, error) {
	backupFilePath, err := xdg.SearchDataFile(
		filepath.Join("f2", "backups", conf.BackupFilename),
	)
	if err != nil {
		//nolint:nilerr // The file does not exist, but it's not an error in this context
		return nil, nil
	}

	fileBytes, err := os.ReadFile(backupFilePath)
	if err != nil {
		return nil, err
	}

	var changes file.Changes

	if err := json.Unmarshal(fileBytes, &changes); err != nil {
		return nil, err
	}

	// Swap source and target for each change to revert the renaming
	for i := range changes {
		ch := changes[i]
		ch.Source, ch.Target = ch.Target, ch.Source
		ch.SourcePath = filepath.Join(ch.BaseDir, ch.Source)
		ch.TargetPath = filepath.Join(ch.BaseDir, ch.Target)
		ch.Status = status.OK

		_, err := os.Stat(ch.SourcePath)
		if errors.Is(err, os.ErrNotExist) {
			ch.Status = status.SourceNotFound
		}

		changes[i] = ch
	}

	if conf.Exec {
		sortfiles.ForRenamingAndUndo(changes, conf.Revert)
	}

	return changes, nil
}

// Find returns a collection of files and directories that match the search
// pattern or explicitly included as command-line arguments.
func Find(conf *config.Config) (changes file.Changes, err error) {
	if conf.Revert {
		return loadFromBackup(conf)
	}

	defer func() {
		if conf.Pair && err == nil {
			sortfiles.Pairs(changes, conf.PairOrder)
			return
		}

		if conf.Sort != config.SortDefault && err == nil {
			sortfiles.Changes(
				changes,
				conf.Sort,
				conf.ReverseSort,
				conf.SortPerDir,
			)
		}
	}()

	if conf.CSVFilename != "" {
		return handleCSV(conf)
	}

	return searchPaths(conf)
}
