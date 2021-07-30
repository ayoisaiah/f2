package f2

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gookit/color"
	"github.com/olekukonko/tablewriter"
)

type colorString string

const (
	red    colorString = "red"
	green  colorString = "green"
	yellow colorString = "yellow"
)

func printColor(c colorString, text string) string {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return text
	}

	switch c {
	case yellow:
		return color.HEX("#FFAB00").Sprint(text)
	case green:
		return color.HEX("#23D160").Sprint(text)
	case red:
		return color.HEX("#FF2F2F").Sprint(text)
	}

	return text
}

func printError(silent bool, err error) {
	if !silent {
		fmt.Fprintln(os.Stderr, err)
	}
}

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

func printTable(data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
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
