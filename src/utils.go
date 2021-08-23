package f2

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/olekukonko/tablewriter"
)

func removeHidden(
	de []os.DirEntry,
	baseDir string,
) (ret []os.DirEntry, err error) {
	for _, e := range de {
		r, err := isHidden(e.Name(), baseDir)
		if err != nil {
			return nil, err
		}

		if !r {
			ret = append(ret, e)
		}
	}

	return ret, nil
}

// contains checks if a string is present in
// a string slice.
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

func printTable(data [][]string, w io.Writer) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Input", "Output", "Status"})
	table.SetAutoWrapText(false)

	for _, v := range data {
		table.Append(v)
	}

	table.Render()
}

// filenameWithoutExtension returns the input file name
// without its extension.
func filenameWithoutExtension(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func greatestCommonDivisor(a, b int) int {
	precision := 0.0001
	if float64(b) < precision {
		return a
	}

	return greatestCommonDivisor(b, a%b)
}

func readCSVFile(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
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
