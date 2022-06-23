package f2

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/djherbis/times.v1"
)

func randate() time.Time {
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min

	return time.Unix(sec, 0)
}

func TestReplaceDateVariables(t *testing.T) {
	testDir := setupFileSystem(t)

	for _, file := range fileSystem {
		path := filepath.Join(testDir, file)

		// change the atime and mtime to a random value
		mtime, atime := randate(), randate()

		err := os.Chtimes(path, atime, mtime)
		if err != nil {
			t.Fatalf("Expected no errors, but got one: %v\n", err)
		}

		timeInfo, err := times.Stat(path)
		if err != nil {
			t.Fatalf("Expected no errors, but got one: %v\n", err)
		}

		want := make(map[string]string)
		got := make(map[string]string)

		accessTime := timeInfo.AccessTime()
		modTime := timeInfo.ModTime()

		fileTimes := []string{"mtime", "atime", "ctime", "btime"}

		for _, v := range fileTimes {
			var timeValue time.Time

			switch v {
			case "mtime":
				timeValue = modTime
			case "atime":
				timeValue = accessTime
			case "ctime":
				timeValue = modTime
				if timeInfo.HasChangeTime() {
					timeValue = timeInfo.ChangeTime()
				}
			case "btime":
				timeValue = modTime
				if timeInfo.HasBirthTime() {
					timeValue = timeInfo.BirthTime()
				}
			}

			for key, token := range dateTokens {
				want[v+"."+key] = timeValue.Format(token)

				dv, err := getDateVar("{{" + v + "." + key + "}}")
				if err != nil {
					t.Fatalf("Test (%s) — Unexpected error: %v", v, err)
				}

				out, err := replaceDateVariables("{{"+v+"."+key+"}}", path, dv)
				if err != nil {
					t.Fatalf("Expected no errors, but got one: %v\n", err)
				}

				got[v+"."+key] = out
			}
		}

		if !cmp.Equal(want, got) {
			t.Fatalf(
				"Expected %v, but got %v\n",
				prettyPrint(want),
				prettyPrint(got),
			)
		}
	}
}

func TestReplaceRandomVariable(t *testing.T) {
	slice := []string{
		`{{10r_l}}`,
		`{{8r_d}}`,
		`{{9r_l}}`,
		`{{5r_ld}}`,
		`{{15r<12345>}}`,
		`{{r}}`,
	}

	for _, v := range slice {
		submatches := randomRegex.FindAllStringSubmatch(v, -1)
		strLen := submatches[0][1]
		length := 10

		var err error

		if strLen != "" {
			length, err = strconv.Atoi(strLen)
			if err != nil {
				t.Fatalf("Test (%s) — Unexpected error: %v", v, err)
			}
		}

		rv, err := getRandomVar(v)
		if err != nil {
			t.Fatalf("Test (%s) — Unexpected error: %v", v, err)
		}

		str := replaceRandomVariables(v, rv)
		if len(str) != length {
			t.Fatalf(
				"Test (%s) — Expected length of random string to be %d, got: %d",
				v,
				length,
				len(str),
			)
		}
	}
}

func TestIntegerToRoman(t *testing.T) {
	testCases := []struct {
		input  int
		output string
	}{
		{463, "CDLXIII"},
		{464, "CDLXIV"},
		{1386, "MCCCLXXXVI"},
		{1838, "MDCCCXXXVIII"},
		{4000, "4000"},
		{7070, "7070"},
	}
	for _, v := range testCases {
		str := integerToRoman(v.input)
		if str != v.output {
			t.Fatalf("Roman(%v) = %v, want %v.", v.input, str, v.output)
		}
	}
}

func TestReplaceTransformVariables(t *testing.T) {
	testDir := setupFileSystem(t)

	cases := []testCase{
		{
			name: "transform directory name to uppercase",
			want: []Change{
				{
					Source:  "docs.03.05.period",
					Target:  "DOCS.03.05.PERIOD",
					BaseDir: testDir,
					IsDir:   true,
				},
			},
			args: "-f '^docs.*' -r {{tr.up}} -D " + testDir,
		},
		{
			name: "transform file name to uppercase",
			want: []Change{
				{
					Source:  "abc.pdf",
					Target:  "ABC.PDF",
					BaseDir: testDir,
				},
				{
					Source:  "abc.epub",
					Target:  "ABC.EPUB",
					BaseDir: testDir,
				},
			},
			args: "-f abc.* -r {{tr.up}} " + testDir,
		},
		{
			name: "transform file extension to title case",
			want: []Change{
				{
					Source:  "abc.pdf",
					Target:  "abc.Pdf",
					BaseDir: testDir,
				},
				{
					Source:  "abc.epub",
					Target:  "abc.Epub",
					BaseDir: testDir,
				},
			},
			args: "-f pdf|epub -r {{tr.ti}} " + testDir,
		},
		{
			name: "transform file name to title case",
			want: []Change{
				{
					Source:  "abc.pdf",
					Target:  "abc_abc_ABC_abc_abc.pdf",
					BaseDir: testDir,
				},
				{
					Source:  "abc.epub",
					Target:  "abc_abc_ABC_abc_abc.epub",
					BaseDir: testDir,
				},
			},
			args: "-f abc.* -r {{tr.di}}_{{tr.lw}}_{{tr.up}}_{{tr.win}}_{{tr.mac}} -e " + testDir,
		},
	}

	runFindReplaceHelper(t, cases)
}

func TestReplaceExifToolVariables(t *testing.T) {
	_, err := exec.LookPath("exiftool")
	if err != nil {
		return
	}

	rootDir := filepath.Join("..", "testdata", "images")

	cases := []testCase{
		{
			name: "Use exiftool data to rename DNG file",
			want: []Change{
				{
					Source:  "proraw.dng",
					BaseDir: rootDir,
				},
			},
			args: "-f proraw.dng -r {{xt.FOV}}_{{xt.ISO}}_{{xt.ImageWidth}} " + rootDir,
		},
	}

	for _, c := range cases {
		f := filenameWithoutExtension(c.want[0].Source)

		jsonFile, err := os.ReadFile(filepath.Join(rootDir, f+"_exiftool.json"))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		var m = make(map[string]interface{})

		err = json.Unmarshal(jsonFile, &m)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		target := fmt.Sprintf(
			"%v_%v_%v",
			m["FOV"],
			m["ISO"],
			m["ImageWidth"],
		)

		c.want[0].Target = target
	}

	runFindReplaceHelper(t, cases)
}
