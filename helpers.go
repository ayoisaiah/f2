package f2

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/djherbis/times.v1"
)

// getTimeInfo retrieves the atime, mtime, ctime and btime for files.
func getTimeInfo(filename string) (times.Timespec, error) {
	t, err := times.Stat(filename)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// getNewPath returns a filename based on the target
// which is not available due to it existing on the filesystem
// or when another renamed file shares the same path.
// It appends an increasing number to the target path until it finds one
// that does not conflict with the filesystem or with another renamed
// file
func getNewPath(target, baseDir string, m map[string][]struct {
	source string
	index  int
}) string {
	f := filenameWithoutExtension(filepath.Base(target))
	re := regexp.MustCompile("\\(\\d+\\)$")
	// Extract the numbered index at the end of the filename (if any)
	match := re.FindStringSubmatch(f)
	if len(match) == 0 {
		match = []string{"(2)"}
		f += " (2)"
	}
	num := 2
	// ignoring error from Sscanf. num will be set to 2 regardless
	fmt.Sscanf(match[0], "(%d)", &num)
	for {
		newPath := re.ReplaceAllString(f, "("+strconv.Itoa(num)+")")
		newPath = newPath + filepath.Ext(target)
		fullPath := filepath.Join(baseDir, newPath)

		// Ensure the new path does not exist on the filesystem
		if _, err := os.Stat(fullPath); err != nil && errors.Is(err, os.ErrNotExist) {
			if m != nil {
				// Check if newPath conflicts with another renamed file
				for k := range m {
					if k == fullPath {
						goto out
					}
				}
			}
			return newPath
		}
	out:
		num += 1
	}
}

func removeDotfiles(de []os.DirEntry) (ret []os.DirEntry) {
	for _, e := range de {
		if e.Name()[0] != 46 {
			ret = append(ret, e)
		}
	}
	return
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

func walk(paths map[string][]os.DirEntry, includeHidden bool) (map[string][]os.DirEntry, error) {
	iterated := []string{}
	var n = make(map[string][]os.DirEntry)

loop:
	for k, v := range paths {
		if contains(iterated, k) {
			continue
		}

		if !includeHidden {
			v = removeDotfiles(v)
		}

		iterated = append(iterated, k)
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
	}

	if len(n) > 0 {
		for k, v := range n {
			paths[k] = v
			delete(n, k)
		}

		goto loop
	}

	return paths, nil
}

func exifDivision(slice []string) string {
	if len(slice) > 0 {
		str := slice[0]
		strSlice := strings.Split(str, "/")
		if len(strSlice) == 2 {
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
