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
func handleCSV(conf *config.Config) ([]*file.Change, error) {
	processed := make(map[string]bool)

	var changes []*file.Change

	records, err := readCSVFile(conf.CSVFilename)
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

		absSourcePath, absErr := filepath.Abs(source)
		if absErr != nil {
			return nil, absErr
		}

		fileInfo, statErr := os.Stat(absSourcePath)
		if statErr != nil {
			// Skip missing source files
			if errors.Is(statErr, os.ErrNotExist) {
				continue
			}
			return nil, statErr
		}

		fileName := fileInfo.Name()

		sourceDir := filepath.Dir(absSourcePath)

		// Ensure that the file is not already processed in the case of
		// duplicate rows
		if processed[absSourcePath] {
			continue
		}

		processed[absSourcePath] = true

		match := &file.Change{
			BaseDir:        sourceDir,
			IsDir:          fileInfo.IsDir(),
			Source:         fileName,
			OriginalSource: fileName,
			RelSourcePath:  absSourcePath,
			CSVRow:         record,
		}

		changes = append(changes, match)

		var target string
		if len(record) > 1 {
			target = strings.TrimSpace(record[1])
		}

		findSlice = append(findSlice, fileName)

		replacementSlice = append(replacementSlice, target)
	}

	if len(conf.ReplacementSlice) == 0 && len(conf.FindSlice) == 0 {
		conf.FindSlice = findSlice
		conf.ReplacementSlice = replacementSlice

		err = conf.SetFindStringRegex(0)
		if err != nil {
			return nil, err
		}
	}

	return changes, nil
}
