package f2

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func printColor(color, text string) string {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return text
	}

	switch color {
	case "yellow":
		return yellow.Sprint(text)
	case "green":
		return green.Sprint(text)
	case "red":
		return red.Sprint(text)
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
// a string slice
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

func filenameWithoutExtension(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

// walk is used to navigate directories recursively
// and include their contents in the pool of paths in
// which to find matches
func walk(
	paths map[string][]os.DirEntry,
	includeHidden bool,
	maxDepth int,
) (map[string][]os.DirEntry, error) {
	var iterated []string
	var n = make(map[string][]os.DirEntry)
	var counter int

loop:
	for k, v := range paths {
		if contains(iterated, k) {
			continue
		}

		if !includeHidden {
			var err error
			v, err = removeHidden(v, k)
			if err != nil {
				return nil, err
			}
		}

		for _, de := range v {
			if de.IsDir() {
				fp := filepath.Join(k, de.Name())
				dirEntry, err := os.ReadDir(fp)
				if err != nil {
					return nil, err
				}

				n[fp] = dirEntry
			}
		}

		iterated = append(iterated, k)
	}

	if len(n) > 0 {
		for k, v := range n {
			paths[k] = v
			delete(n, k)
		}

		counter++
		if !(maxDepth > 0 && counter == maxDepth) {
			goto loop
		}
	}

	return paths, nil
}

func exifDivision(slice []string) string {
	if len(slice) > 0 {
		str := slice[0]
		strSlice := strings.Split(str, "/")
		expectedLength := 2
		if len(strSlice) == expectedLength {
			numerator, err := strconv.Atoi(strSlice[0])
			if err != nil {
				return ""
			}

			denominator, err := strconv.Atoi(strSlice[1])
			if err != nil {
				return ""
			}

			v := float64(numerator) / float64(denominator)
			str, err := strconv.FormatFloat(v, 'f', -1, 64), nil
			if err != nil {
				return ""
			}

			return str
		}
	}

	return ""
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
