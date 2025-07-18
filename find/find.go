package find

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/araddon/dateparse"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/eval"
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

		ch.SortCriterion.TimeVar = timeVal
	}

	if conf.Sort == config.SortStringVar {
		ch.SortCriterion.StringVar = ch.Target
	}

	if conf.Sort == config.SortIntVar {
		// if variable cannot be parsed into a valid integer, default to zero
		intVal, _ := strconv.Atoi(ch.Target)

		ch.SortCriterion.IntVar = intVal
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
	match *file.Change,
	searchVars *variables.Variables,
) (removeMatch bool, err error) {
	if match.PrimaryPair != nil && match.PrimaryPair.MatchesFindCond {
		match.MatchesFindCond = true
		return false, nil
	}

	err = variables.Replace(conf, match, searchVars)
	if err != nil {
		return true, err
	}

	result, err := eval.Evaluate(match.Target)
	if err != nil {
		return true, err
	}

	match.Target = ""

	if !result {
		return true, nil
	}

	match.MatchesFindCond = true

	return false, nil
}

func checkIfMatch(
	conf *config.Config,
	path string,
	entry fs.DirEntry,
	sortVars *variables.Variables,
) (*file.Change, bool, error) {
	var err error

	fileInfo, err := entry.Info()
	if err != nil {
		return nil, false, err
	}

	var isMatch bool

	if conf.Search.FindCond != nil {
		match := createFileChange(conf, path, fileInfo)
		match.Target = conf.Search.FindCond.String()

		return match, true, err
	}

	fileName := entry.Name()

	if conf.IgnoreExt && !entry.IsDir() && !conf.Pair {
		fileName = pathutil.StripExtension(fileName)
	}

	if !conf.Search.Regex.MatchString(fileName) {
		return nil, false, nil
	}

	match := createFileChange(conf, path, fileInfo)

	isMatch = true

	slog.Debug(
		"found file matching search pattern",
		slog.Any("match", match),
	)

	if shouldFilter(conf, match) {
		slog.Debug(
			"excluding file based on filter criteria",
			slog.Any("match", match),
		)

		return match, false, nil
	}

	if conf.SortVariable != "" {
		err = extractCustomSort(conf, match, sortVars)
		if err != nil {
			return nil, false, err
		}
	}

	return match, isMatch, nil
}

func walkDirectory(
	conf *config.Config,
	rootPath string,
	processedPaths map[string]bool,
	sortVars *variables.Variables,
) (file.Changes, error) {
	var matches file.Changes

	maxDepth := -1
	if conf.Recursive {
		maxDepth = conf.MaxDepth
	}

	err := filepath.WalkDir(
		rootPath,
		func(currentPath string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if rootPath == currentPath || processedPaths[currentPath] {
				return nil
			}

			if shouldSkipHidden, skipErr := skipFileIfHidden(
				currentPath,
				conf.FilesAndDirPaths,
				conf.IncludeHidden,
			); skipErr != nil {
				return skipErr
			} else if shouldSkipHidden {
				slog.Debug(
					"skipping hidden path",
					slog.String("path", currentPath),
				)

				if entry.IsDir() {
					return fs.SkipDir
				}

				return nil
			}

			if shouldSkipExcludedDir(conf, entry) {
				slog.Debug(
					"skipping excluded directory",
					slog.String("path", currentPath),
				)

				return fs.SkipDir
			}

			if isMaxDepth(rootPath, currentPath, maxDepth) {
				slog.Debug(
					"max depth reached, skipping directory traversal",
					slog.String("root_path", rootPath),
					slog.String("path", currentPath),
					slog.Int("max_depth", maxDepth),
				)

				return fs.SkipDir
			}

			match, isMatch, err := checkIfMatch(
				conf,
				currentPath,
				entry,
				sortVars,
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

	return matches, err
}

func shouldSkipExcludedDir(conf *config.Config, entry fs.DirEntry) bool {
	return entry.IsDir() && conf.Recursive &&
		conf.ExcludeDirRegex != nil &&
		conf.ExcludeDirRegex.MatchString(entry.Name())
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

			match, isMatch, matchErr := checkIfMatch(
				conf,
				rootPath,
				fs.FileInfoToDirEntry(fileInfo),
				sortVars,
			)
			if matchErr != nil {
				return nil, matchErr
			}

			if isMatch {
				matches = append(matches, match)
			}

			processedPaths[rootPath] = true

			continue
		}

		dirMatches, err := walkDirectory(
			conf,
			rootPath,
			processedPaths,
			sortVars,
		)
		if err != nil {
			return nil, err
		}

		matches = append(matches, dirMatches...)
	}

	if conf.Search.FindCond != nil {
		var err error

		matches, err = processFindExpression(conf, matches, searchVars)
		if err != nil {
			return nil, err
		}
	}

	return matches, nil
}

func processFindExpression(
	conf *config.Config,
	matches file.Changes,
	searchVars *variables.Variables,
) (file.Changes, error) {
	if conf.Pair {
		sortfiles.Pairs(matches, conf.PairOrder)
		slog.Debug(
			"finished sorting file pairings",
			slog.Any("matches", matches),
		)
	}

	if conf.ExifToolVarPresent {
		names, indices := matches.SourceNamesWithIndices(conf.Pair)

		slog.Debug("extracting exif variables", slog.Any("paths", names))

		fileMeta, err := variables.ExtractExiftoolMetadata(
			conf,
			names...)
		if err != nil {
			return nil, err
		}

		for i := range fileMeta {
			index := indices[i]
			matches[index].ExiftoolData = &fileMeta[i]
			slog.Debug(
				"attaching exif data to file",
				slog.String("match", matches[index].SourcePath),
				slog.String("file", fileMeta[i].File),
				slog.Bool(
					"is_match",
					fileMeta[i].File == matches[index].SourcePath,
				),
			)
		}
	}

	matches = slices.DeleteFunc(matches, func(match *file.Change) bool {
		removeMatch, err := evaluateSearchCondition(conf, match, searchVars)
		if err != nil {
			if conf.Verbose {
				report.SearchEvalFailed(match.SourcePath, match.Target, err)
			}
		}

		if removeMatch {
			slog.Debug(
				"excluding file: find condition evaluated to false",
				slog.String("path", match.SourcePath),
				slog.String("evaluated", match.Target),
			)
		}

		return removeMatch
	})

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

	var backup file.Backup

	err = json.Unmarshal(fileBytes, &backup)
	if err != nil {
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

		_, err = os.Stat(ch.SourcePath)
		if errors.Is(err, os.ErrNotExist) {
			ch.Status = status.SourceNotFound
		}

		changes[i] = ch
	}

	if conf.Exec {
		sortfiles.ForRenamingAndUndo(changes, conf.Revert)

		// recreate empty directories that were cleaned
		for _, v := range backup.CleanedDirs {
			err = os.MkdirAll(v, osutil.DirPermission)
			if err != nil {
				return nil, err
			}
		}
	}

	return changes, nil
}

// Find returns a collection of files and directories that match the search
// pattern or explicitly included as command-line arguments.
func Find(conf *config.Config) (matches file.Changes, err error) {
	if conf.Revert {
		return loadFromBackup(conf)
	}

	sortVars, err := variables.Extract(conf.SortVariable)
	if err != nil {
		return nil, err
	}

	var searchVars variables.Variables
	if conf.Search.FindCond != nil {
		searchVars, err = variables.Extract(
			conf.Search.FindCond.String(),
		)
		if err != nil {
			return nil, err
		}
	}

	if conf.CSVFilename != "" {
		matches, err = handleCSV(conf)
	} else {
		matches, err = searchPaths(conf, &sortVars, &searchVars)
		if err != nil {
			return nil, err
		}
	}

	slog.Debug(
		"search complete",
		slog.Any("matches", matches),
	)

	if conf.Search.FindCond == nil && conf.Pair {
		sortfiles.Pairs(matches, conf.PairOrder)
		slog.Debug(
			"finished sorting file pairings",
			slog.Any("matches", matches),
		)
	}

	if conf.Sort != config.SortDefault {
		sortfiles.Changes(matches, conf)
		slog.Debug(
			fmt.Sprintf("finished sorting matches by %s", conf.Sort),
			slog.Any("matches", matches),
		)
	}

	// If using indices without an explicit sort, ensure that the files
	// are arranged hierarchically
	if conf.IndexPresent && conf.Sort == config.SortDefault {
		sortfiles.Hierarchically(matches)
	}

	return
}
