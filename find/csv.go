package find

import (
	"encoding/csv"
	"io/fs"
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

	csvReader := csv.NewReader(f)

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

// handleCSV reads the provided CSV file, and finds all the valid candidates
// for replacement.
func handleCSV(conf *config.Config) ([]*file.Change, error) {
	processed := make(map[string]bool)

	var changes []*file.Change

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

		for _, entry := range dirEntry {
			entryName := entry.Name()

			if entryName != fileInfo.Name() {
				continue
			}

			relPath := filepath.Join(sourceDir, entryName)

			// Ensure that the file is not already processed in the case of
			// duplicate rows
			if processed[relPath] {
				break
			}

			processed[relPath] = true

			fc := &file.Change{
				BaseDir:        sourceDir,
				IsDir:          entry.IsDir(),
				Source:         entryName,
				OriginalSource: entryName,
				RelSourcePath:  relPath,
				CSVRow:         record,
			}

			changes = append(changes, fc)

			break
		}

		if len(record) > 1 {
			target := strings.TrimSpace(record[1])

			replacementSlice = append(replacementSlice, target)
		}
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

	return changes, nil
}
