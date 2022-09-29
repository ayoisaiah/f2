package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/pterm/pterm"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func CleanString(str string) string {
	return nonAlphanumericRegex.ReplaceAllString(str, "_")
}

// contains checks if a string is present in
// a string slice.
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

func ContainsInt(sl []int, i int) bool {
	for _, v := range sl {
		if v == i {
			return true
		}
	}

	return false
}

func PrintTable(data [][]string, writer io.Writer) {
	d := [][]string{
		{"ORIGINAL", "RENAMED", "STATUS"},
	}

	d = append(d, data...)

	table := pterm.DefaultTable
	table.HeaderRowSeparator = "*"
	table.Boxed = true

	str, err := table.WithHasHeader().WithData(d).Srender()
	if err != nil {
		pterm.Error.Printfln("Unable to print table: %s", err.Error())
		return
	}

	fmt.Fprintln(writer, str)
}

// filenameWithoutExtension returns the input file name
// without its extension.
func FilenameWithoutExtension(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

func PrettyPrint(i interface{}) string {
	//nolint:errchkjson // no need to check error
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func GreatestCommonDivisor(a, b int) int {
	precision := 0.0001
	if float64(b) < precision {
		return a
	}

	return GreatestCommonDivisor(b, a%b)
}

func ReadCSVFile(filePath string) ([][]string, error) {
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
