package find

import (
	"bufio"
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/report"
)

// readCSVFile reads all the records contained in a CSV file specified by
// `pathToCSV`.
func readCSVFile(pathToCSV string) ([][]string, error) {
	f, err := os.Open(pathToCSV)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	// Use bufio for potential performance gains with large CSV files
	csvReader := csv.NewReader(bufio.NewReader(f))

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

// handleCSV reads the provided CSV file, and finds all the valid candidates
// for renaming.
func handleCSV(conf *config.Config) (file.Changes, error) {
	processed := make(map[string]bool)

	var changes file.Changes

	records, err := readCSVFile(conf.CSVFilename)
	if err != nil {
		return nil, err
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Change to the directory of the CSV file
	err = os.Chdir(conf.WorkingDir)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = os.Chdir(currentDir)
	}()

	for i, record := range records {
		if len(record) == 0 {
			continue
		}

		source := strings.TrimSpace(record[0])

		fileInfo, statErr := os.Stat(source)
		if statErr != nil {
			// Skip missing source files
			if errors.Is(statErr, os.ErrNotExist) {
				if conf.Verbose {
					report.NonExistentFile(source, i+1)
				}

				continue
			}

			return nil, statErr
		}

		fileName := fileInfo.Name()

		sourceDir := filepath.Dir(source)

		// Ensure that the file is not already processed in the case of
		// duplicate rows
		if processed[source] {
			continue
		}

		processed[source] = true

		match := &file.Change{
			BaseDir:      sourceDir,
			TargetDir:    sourceDir,
			IsDir:        fileInfo.IsDir(),
			Source:       fileName,
			Target:       fileName,
			OriginalName: fileName,
			SourcePath:   filepath.Join(sourceDir, fileName),
			CSVRow:       record,
			Position:     i,
		}

		if conf.TargetDir != "" {
			match.TargetDir = conf.TargetDir
		}

		if len(record) > 1 {
			match.Target = strings.TrimSpace(record[1])

			if filepath.IsAbs(match.Target) {
				match.TargetDir = ""
				continue
			}
		}

		changes = append(changes, match)
	}

	return changes, nil
}
