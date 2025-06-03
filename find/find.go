package find

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/araddon/dateparse"
	"github.com/maja42/goval"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
	"github.com/ayoisaiah/f2/v2/internal/osutil"
	"github.com/ayoisaiah/f2/v2/internal/pathutil"
	"github.com/ayoisaiah/f2/v2/internal/sortfiles"
	"github.com/ayoisaiah/f2/v2/internal/status"
	"github.com/ayoisaiah/f2/v2/replace/variables"
	"github.com/ayoisaiah/f2/v2/report"
)

const (
	dotCharacter = 46
)

// shouldFilter decides whether a match should be included in the final
// pool of files for renaming.
func shouldFilter(conf *config.Config, match *file.Change) bool {
	if match == nil {
		return true
	}

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

	if conf.IncludeRegex != nil &&
		!conf.IncludeRegex.MatchString(match.Source) {
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

func extractCustomSort(
	conf *config.Config,
	ch *file.Change,
	vars *variables.Variables,
) error {
	// Temporarily set Target to SortVariable due to how variables.Replace() works
	ch.Target = conf.SortVariable

	err := variables.Replace(conf, ch, vars)
	if err != nil {
		return err
	}

	if conf.Sort == config.SortTimeVar {
		// if variable cannot be parsed into a valid time, default to zero value
		timeVal, _ := dateparse.ParseAny(ch.Target)

		ch.CustomSort.Time = timeVal
	}

	if conf.Sort == config.SortStringVar {
		ch.CustomSort.String = ch.Target
	}

	if conf.Sort == config.SortIntVar {
		// if variable cannot be parsed into a valid integer, default to zero
		intVal, _ := strconv.Atoi(ch.Target)

		ch.CustomSort.Int = intVal
	}

	// Reset to an empty string once custom sort variable has been extracted and
	// assigned accordingly
	ch.Target = ""

	return nil
}

func createFileChange(
	conf *config.Config,
	dirPath string,
	fileInfo fs.FileInfo,
) *file.Change {
	baseDir := filepath.Dir(dirPath)
	fileName := fileInfo.Name()

	match := &file.Change{
		BaseDir:      baseDir,
		TargetDir:    baseDir,
		IsDir:        fileInfo.IsDir(),
		Source:       fileName,
		OriginalName: fileName,
		SourcePath:   filepath.Join(baseDir, fileName),
	}

	if conf.TargetDir != "" {
		match.TargetDir = conf.TargetDir
	}

	return match
}

func evaluateSearchCondition(
	conf *config.Config,
	currentPath string,
	fileInfo os.FileInfo,
	searchVars *variables.Variables,
) (*file.Change, bool, error) {
	match := createFileChange(conf, currentPath, fileInfo)

	match.Target = conf.Search.FindCond.String()

	err := variables.Replace(conf, match, searchVars)
	if err != nil {
		return match, false, err
	}

	eval := goval.NewEvaluator()

	result, err := eval.Evaluate(match.Target, nil, nil)
	if err != nil {
		if conf.Verbose {
			report.SearchEvalFailed(currentPath, match.Target, err)
		}

		return match, false, nil
	}

	r, _ := result.(bool)
	if !r {
		return match, false, nil
	}

	match.Target = ""
	match.MatchesFindCond = true

	return match, true, nil
}

func checkIfMatch(
	conf *config.Config,
	path string,
	entry fs.DirEntry,
	sortVars,
	searchVars *variables.Variables,
) (*file.Change, bool, error) {
	var match *file.Change

	var err error

	fileInfo, err := entry.Info()
	if err != nil {
		return match, false, err
	}

	var isMatch bool

	if conf.Search.FindCond != nil {
		match, isMatch, err = evaluateSearchCondition(
			conf,
			path,
			fileInfo,
			searchVars,
		)
		if err != nil {
			return nil, false, err
		}
	} else {
		fileName := entry.Name()

		if conf.IgnoreExt && !entry.IsDir() {
			fileName = pathutil.StripExtension(fileName)
		}

		if conf.Search.Regex.MatchString(fileName) {
			match = createFileChange(conf, path, fileInfo)
			isMatch = true
		}
	}

	if shouldFilter(conf, match) {
		return match, false, nil
	}

	err = extractCustomSort(conf, match, sortVars)
	if err != nil {
		return nil, false, err
	}

	return match, isMatch, nil
}

// searchPaths walks through the filesystem and finds matches for the provided
// search pattern or variables comparison.
func searchPaths(
	conf *config.Config,
	sortVars, searchVars *variables.Variables,
) (file.Changes, error) {
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

			match, isMatch, err := checkIfMatch(
				conf,
				rootPath,
				fs.FileInfoToDirEntry(fileInfo),
				sortVars,
				searchVars,
			)
			if err != nil {
				return nil, err
			}

			if isMatch {
				matches = append(matches, match)
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

				match, isMatch, err := checkIfMatch(
					conf,
					currentPath,
					entry,
					sortVars,
					searchVars,
				)
				if err != nil {
					return err
				}

				if isMatch {
					matches = append(matches, match)
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
	backupFilePath := filepath.Join(
		os.TempDir(),
		"f2",
		"backups",
		conf.BackupFilename,
	)

	_, err := os.Stat(backupFilePath)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	fileBytes, err := os.ReadFile(backupFilePath)
	if err != nil {
		return nil, err
	}

	var backup config.Backup

	if err := json.Unmarshal(fileBytes, &backup); err != nil {
		return nil, err
	}

	changes := backup.Changes

	// Swap source and target for each change to revert the renaming
	for i := range changes {
		ch := changes[i]
		p := filepath.Join(ch.TargetDir, ch.Target)
		ch.Target = filepath.Base(p)
		ch.TargetDir = filepath.Dir(p)

		ch.Source, ch.Target = ch.Target, ch.Source
		ch.BaseDir, ch.TargetDir = ch.TargetDir, ch.BaseDir
		ch.SourcePath = filepath.Join(ch.BaseDir, ch.Source)
		ch.TargetPath = filepath.Join(ch.TargetDir, ch.Target)
		ch.Status = status.OK

		_, err := os.Stat(ch.SourcePath)
		if errors.Is(err, os.ErrNotExist) {
			ch.Status = status.SourceNotFound
		}

		changes[i] = ch
	}

	if conf.Exec {
		sortfiles.ForRenamingAndUndo(changes, conf.Revert)

		// recreate empty directories that were cleaned
		for _, v := range backup.CleanedDirs {
			_ = os.MkdirAll(v, osutil.DirPermission)
		}
	}

	return changes, nil
}

// Find returns a collection of files and directories that match the search
// pattern or explicitly included as command-line arguments.
func Find(conf *config.Config) (changes file.Changes, err error) {
	var (
		sortVars   variables.Variables
		searchVars variables.Variables
	)

	if conf.SortVariable != "" {
		sortVars, err = variables.Extract(conf.SortVariable)
		if err != nil {
			return nil, err
		}
	}

	if conf.Search.FindCond != nil {
		searchVars, err = variables.Extract(
			conf.Search.FindCond.String(),
		)
		if err != nil {
			return nil, err
		}
	}

	if conf.Revert {
		return loadFromBackup(conf)
	}

	defer func() {
		if conf.Pair && err == nil {
			sortfiles.Pairs(changes, conf.PairOrder)
		}

		if conf.Sort != config.SortDefault && err == nil {
			sortfiles.Changes(changes, conf)
		}
	}()

	if conf.CSVFilename != "" {
		return handleCSV(conf)
	}

	return searchPaths(conf, &sortVars, &searchVars)
}
