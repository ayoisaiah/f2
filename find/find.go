package find

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/file"
	internalpath "github.com/ayoisaiah/f2/internal/path"
)

const (
	dotCharacter = 46
)

// shouldFilter decides whether a match should be included in the final
// pool of files for renaming.
func shouldFilter(conf *config.Config, match *file.Change) (bool, error) {
	if conf.ExcludeRegex != nil &&
		conf.ExcludeRegex.MatchString(match.Source) {
		return true, nil
	}

	if !conf.IncludeDir && match.IsDir {
		return true, nil
	}

	if conf.OnlyDir && !match.IsDir {
		return true, nil
	}

	if !conf.IncludeHidden {
		isHidden, err := checkIfHidden(match.Source, match.BaseDir)
		if err != nil {
			return false, err
		}

		if isHidden {
			// Ensure that explicitly included file arguments are not affected
			entryAbsPath, err := filepath.Abs(
				filepath.Join(match.BaseDir, match.Source),
			)
			if err != nil {
				return false, err
			}

			shouldSkip := true

			for _, pathArg := range conf.FilesAndDirPaths {
				argAbsPath, err := filepath.Abs(pathArg)
				if err != nil {
					return false, err
				}

				if strings.EqualFold(entryAbsPath, argAbsPath) {
					shouldSkip = false
				}
			}

			return shouldSkip, nil
		}
	}

	return false, nil
}

// isMaxDepth reports whether the configured max depth has been reached.
func isMaxDepth(rootPath, currentPath string, maxDept int) bool {
	if rootPath == filepath.Dir(currentPath) {
		return false
	}

	if maxDept == -1 {
		return true
	}

	p := strings.Replace(currentPath, rootPath+string(os.PathSeparator), "", 1)

	if strings.Count(p, string(os.PathSeparator)) > maxDept && maxDept != 0 {
		return true
	}

	return false
}

// searchPaths walks through the filesystem and finds matches for the provided
// search pattern.
func searchPaths(conf *config.Config) ([]*file.Change, error) {
	processedPaths := make(map[string]bool)

	var matches []*file.Change

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

			match := &file.Change{
				BaseDir:        filepath.Dir(rootPath),
				IsDir:          fileInfo.IsDir(),
				Source:         fileInfo.Name(),
				OriginalSource: fileInfo.Name(),
			}

			excludeMatch, ferr := shouldFilter(conf, match)
			if ferr != nil {
				return nil, ferr
			}

			if !excludeMatch {
				matches = append(matches, match)

				processedPaths[rootPath] = true
			}

			continue
		}

		maxDepth := -1
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

				if isMaxDepth(rootPath, currentPath, maxDepth) {
					return fs.SkipDir
				}

				fileName := entry.Name()

				entryIsDir := entry.IsDir()

				if conf.IgnoreExt && !entryIsDir {
					fileName = internalpath.StripExtension(fileName)
				}

				matched := conf.SearchRegex.MatchString(fileName)
				if !matched {
					return nil
				}

				fileName = entry.Name()

				match := &file.Change{
					BaseDir:        filepath.Dir(currentPath),
					IsDir:          entryIsDir,
					Source:         fileName,
					OriginalSource: fileName,
				}

				excludeMatch, ferr := shouldFilter(conf, match)
				if ferr != nil {
					return ferr
				}

				if !excludeMatch {
					matches = append(matches, match)

					processedPaths[currentPath] = true
				}

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
