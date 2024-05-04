package find

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/pathutil"
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
			slog.Debug(
				"hidden file is explicitly included, not skipping",
				slog.String("path", path),
			)

			return false, nil
		}
	}

	return true, nil // Skip the hidden file
}

// isMaxDepth reports whether the configured max depth has been reached.
func isMaxDepth(rootPath, currentPath string, maxDepth int) bool {
	if rootPath == filepath.Dir(currentPath) {
		return false
	}

	if maxDepth == -1 {
		return true
	}

	p := strings.Replace(currentPath, rootPath+string(os.PathSeparator), "", 1)

	if strings.Count(p, string(os.PathSeparator)) > maxDepth && maxDepth != 0 {
		return true
	}

	return false
}

func createFileChange(dirPath string, fileInfo fs.FileInfo) *file.Change {
	baseDir := filepath.Dir(dirPath)
	fileName := fileInfo.Name()

	match := &file.Change{
		BaseDir:        baseDir,
		IsDir:          fileInfo.IsDir(),
		Source:         fileName,
		OriginalSource: fileName,
		RelSourcePath:  filepath.Join(baseDir, fileName),
	}

	return match
}

// searchPaths walks through the filesystem and finds matches for the provided
// search pattern.
func searchPaths(conf *config.Config) ([]*file.Change, error) {
	slog.Debug(
		"searching path arguments for matches",
		slog.Any("paths", conf.FilesAndDirPaths),
	)

	processedPaths := make(map[string]bool)

	var matches []*file.Change

	for _, rootPath := range conf.FilesAndDirPaths {
		rootPath = filepath.Clean(rootPath)

		fileInfo, err := os.Stat(rootPath)
		if err != nil {
			return nil, err
		}

		if !fileInfo.IsDir() {
			slog.Debug(
				"processing root file argument",
				slog.Any("path", rootPath),
			)

			if processedPaths[rootPath] {
				slog.Debug(
					"skipping processed file",
					slog.String("path", rootPath),
				)

				continue
			}

			if conf.SearchRegex.MatchString(fileInfo.Name()) {
				match := createFileChange(rootPath, fileInfo)

				if !shouldFilter(conf, match) {
					slog.Debug(
						"match found and passed filters",
						slog.String("path", rootPath),
					)

					matches = append(matches, match)
				} else {
					slog.Debug("match found but excluded", slog.String("path", rootPath))
				}
			} else {
				slog.Debug("file not matched for renaming", slog.String("path", rootPath))
			}

			processedPaths[rootPath] = true

			continue
		}

		maxDepth := -1 // default value for non-recursive iterations
		if conf.Recursive {
			maxDepth = conf.MaxDepth

			slog.Debug(
				"recursively traversing directories to search for matches",
				slog.Int("max_depth", maxDepth),
			)
		}

		slog.Debug(
			"processing root directory argument",
			slog.Any("path", rootPath),
		)

		err = filepath.WalkDir(
			rootPath,
			func(currentPath string, entry fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				// skip the root path and already processed paths
				if rootPath == currentPath || processedPaths[currentPath] {
					slog.Debug(
						"skipping processed path",
						slog.String("path", currentPath),
						slog.Bool("is_root", rootPath == currentPath),
					)

					return nil
				}

				if skipHidden, hiddenErr := skipFileIfHidden(
					currentPath,
					conf.FilesAndDirPaths,
					conf.IncludeHidden,
				); hiddenErr != nil {
					return hiddenErr
				} else if skipHidden {
					slog.Debug("skipping hidden path", slog.String("path", currentPath))

					if entry.IsDir() {
						return fs.SkipDir
					}

					return nil
				}

				if isMaxDepth(rootPath, currentPath, maxDepth) {
					slog.Debug(
						"skipping entire directory: max depth reached",
						slog.String("path", currentPath),
						slog.String("parent_dir", filepath.Dir(currentPath)),
					)

					return fs.SkipDir
				}

				fileName := entry.Name()

				entryIsDir := entry.IsDir()

				if conf.IgnoreExt && !entryIsDir {
					fileName = pathutil.StripExtension(fileName)

					slog.Debug(
						"extension stripped",
						slog.String("old_filename", entry.Name()),
						slog.String("new_filename", fileName),
						slog.Bool("is_dir", entryIsDir),
					)
				}

				if conf.SearchRegex.MatchString(fileName) {
					fileInfo, infoErr := entry.Info()
					if infoErr != nil {
						return infoErr
					}

					match := createFileChange(currentPath, fileInfo)

					if !shouldFilter(conf, match) {
						slog.Debug(
							"match found and passed filters",
							slog.Any("path", currentPath),
						)

						matches = append(matches, match)
					} else {
						slog.Debug("match found but excluded", slog.String("path", currentPath))
					}
				} else {
					slog.Debug("file not matched for renaming", slog.String("path", currentPath))
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

// Find returns a collection of files and directories that match the search
// pattern or explicitly included as command-line arguments.
func Find(conf *config.Config) ([]*file.Change, error) {
	if conf.CSVFilename != "" {
		return handleCSV(conf)
	}

	return searchPaths(conf)
}
