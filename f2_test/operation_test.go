package f2_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	stdpath "path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/adrg/xdg"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/sebdah/goldie/v2"

	"github.com/ayoisaiah/f2"
	"github.com/ayoisaiah/f2/internal/utils"
)

func init() {
	//nolint:dogsled // necessary for testing setup
	_, filename, _, _ := runtime.Caller(0)

	dir := stdpath.Join(stdpath.Dir(filename), "..")

	err := os.Chdir(dir)
	if err != nil {
		log.Fatal(err)
	}

	workingDir, err := filepath.Abs(".")
	if err != nil {
		log.Fatalf("Unable to retrieve working directory: %v", err)
	}

	workingDir = strings.ReplaceAll(workingDir, "/", "_")
	if runtime.GOOS == f2.Windows {
		workingDir = strings.ReplaceAll(workingDir, `\`, "_")
		workingDir = strings.ReplaceAll(workingDir, ":", "_")
	}

	backupFilePath, err = xdg.DataFile(
		filepath.Join("f2", "backups", workingDir+".json"),
	)
	if err != nil {
		log.Fatalf("Unable to retrieve xdg data file directory: %v", err)
	}

	rand.Seed(time.Now().UnixNano())
}

var testFixtures = "testdata"

var (
	backupFilePath string
)

var fileSystem = []string{
	"docs/éèêëçñåēčŭ.xlsx",
	"dev/index.js",
	"dev/index.ts",
	"docu.ments/job-contract.docx",
	"special/$-(+)_file.txt",
	"images/dsc-001.arw",
	"images/dsc-002.arw",
	"images/sony/dsc-003.arw",
	"images/canon/startrails1.jpg",
	"images/canon/startrails2.jpg",
	"movies/No Pressure (2021) S1.E1.1080p.mkv",
	"movies/No Pressure (2021) S1.E2.1080p.mkv",
	"movies/No Pressure (2021) S1.E3.1080p.mkv",
	"movies/green-mile_1999.mp4",
	"music/Overgrown (2013)/01 Overgrown.flac",
	"music/Overgrown (2013)/02 I Am Sold.flac",
	"music/Overgrown (2013)/Cover.jpg",
	"ebooks/atomic-habits.pdf",
	"ebooks/1984.pdf",
	"ebooks/animal-farm.epub",
	"ebooks/fear-of-life.EPUB",
	"ebooks/green-mile_1996.mobi",
	"ebooks/.banned/.mein-kampf.pdf",
	"ebooks/.banned/lolita.epub",
	".golang.pdf",
}

// setupFileSystem creates all required files and folders for
// the tests and returns a function that is used as
// a teardown function when the tests are done.
func setupFileSystem(tb testing.TB) string {
	tb.Helper()

	testDir := tb.TempDir()

	absPath, err := filepath.Abs(testDir)
	if err != nil {
		tb.Fatalf("Unable to get absolute path to test directory: %v", err)
	}

	tb.Cleanup(func() {
		if err = os.RemoveAll(absPath); err != nil {
			tb.Fatalf(
				"Failure occurred while cleaning up the filesystem: %v",
				err,
			)
		}
	})

	for _, v := range fileSystem {
		dir := filepath.Dir(v)

		filePath := filepath.Join(testDir, dir)

		err = os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			tb.Fatalf(
				"Unable to create directories in path: '%s', due to err: %v",
				filePath,
				err,
			)
		}
	}

	for _, f := range fileSystem {
		pathToFile := filepath.Join(absPath, f)

		file, err := os.Create(pathToFile)
		if err != nil {
			tb.Fatalf(
				"Unable to write to file: '%s', due to err: %v",
				pathToFile,
				err,
			)
		}

		file.Close()
	}

	return absPath
}

func newTestRun(args []string) ([]byte, error) {
	var buf bytes.Buffer

	app := f2.GetApp(os.Stdin, &buf)

	err := app.Run(args)
	if err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

func sortChanges(s []f2.Change) {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Source < s[j].Source
	})
}

func parseArgs(t *testing.T, name, args string) []string {
	t.Helper()

	result := make([]string, len(os.Args))

	copy(result, os.Args)

	if runtime.GOOS == f2.Windows {
		args = strings.ReplaceAll(args, `\`, `₦`)
	}

	argsSlice, err := shellquote.Split(args)
	if err != nil {
		t.Fatalf(
			"Test (%s) -> shellquote.Split(%s) yielded error: %v",
			name,
			args,
			err,
		)
	}

	if runtime.GOOS == f2.Windows {
		for i, v := range argsSlice {
			argsSlice[i] = strings.ReplaceAll(v, `₦`, `\`)
		}
	}

	result = append(result[:1], argsSlice...)

	return result
}

type TestCase struct {
	Name        string                            `json:"name"`
	Changes     []f2.Change                       `json:"changes"`
	Want        []string                          `json:"want"`
	Args        string                            `json:"args"`
	PathArgs    []string                          `json:"path_args"`
	Conflicts   map[f2.ConflictType][]f2.Conflict `json:"conflicts"`
	DefaultOpts string                            `json:"default_opts"`
	GoldenFile  string                            `json:"golden_file"`
	Setup       []string                          `json:"setup"`
}

func retrieveTestCases(t *testing.T, filename string) []TestCase {
	t.Helper()

	var cases []TestCase

	b, err := os.ReadFile(filepath.Join(testFixtures, filename))
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(b, &cases)
	if err != nil {
		t.Fatal(err)
	}

	for i := range cases {
		tc := cases[i]

		for _, v := range tc.Want {
			var ch f2.Change

			tokens := strings.Split(v, "|")

			for j, token := range tokens {
				if j == 0 {
					ch.Source = token
					continue
				}

				if j == 1 {
					ch.Target = token
					continue
				}

				if j == 2 {
					if token != "" {
						ch.BaseDir = token
					}

					continue
				}

				r, err := strconv.ParseBool(token)
				if err != nil {
					t.Fatal(err)
				}

				if j == 3 {
					ch.IsDir = r
					continue
				}

				if j == 4 {
					ch.WillOverwrite = r
					continue
				}
			}

			tc.Changes = append(tc.Changes, ch)
		}
	}

	return cases
}

func preTestSetup(
	t *testing.T,
	testDir string,
	setup []string,
) (string, error) {
	t.Helper()

	if utils.Contains(setup, "testdata") || utils.Contains(setup, "golden") {
		testDir = testFixtures
	}

	if utils.Contains(setup, "windows_hidden") {
		err := setHidden(filepath.Join(testDir, "images"))
		if err != nil {
			t.Fatal(err)
		}
	}

	if utils.Contains(setup, "exiftool") {
		_, err := exec.LookPath("exiftool")
		if err != nil {
			t.SkipNow()
		}
	}

	if utils.Contains(setup, "date variables") {
		mtime := time.Date(2022, time.April, 10, 13, 0, 0, 0, time.UTC)
		atime := time.Date(2023, time.July, 11, 13, 0, 0, 0, time.UTC)

		for _, file := range fileSystem {
			path := filepath.Join(testDir, file)

			err := os.Chtimes(path, atime, mtime)
			if err != nil {
				return "", err
			}
		}
	}

	return testDir, nil
}

func runTestCases(t *testing.T, cases []TestCase) {
	t.Helper()

	for i := range cases {
		tc := cases[i]

		t.Run(tc.Name, func(t *testing.T) {
			testDir := setupFileSystem(t)

			if len(tc.Setup) > 0 {
				v, err := preTestSetup(t, testDir, tc.Setup)
				if err != nil {
					t.Fatal(err)
				}

				testDir = v
			}

			if tc.DefaultOpts != "" {
				os.Setenv(f2.EnvDefaultOpts, tc.DefaultOpts)
			}

			for j := range tc.Changes {
				ch := tc.Changes[j]
				if ch.BaseDir == "" {
					tc.Changes[j].BaseDir = testDir
				} else {
					tc.Changes[j].BaseDir = filepath.Join(testDir, ch.BaseDir)
				}
			}

			pathArgs := testDir
			if len(tc.PathArgs) != 0 {
				var res []string
				for _, v := range tc.PathArgs {
					res = append(
						res,
						fmt.Sprintf("'%s'", filepath.Join(testDir, v)),
					)
				}

				pathArgs = strings.Join(res, " ")
			}

			csvTestFile := filepath.Join(testFixtures, "input.csv")

			if strings.Contains(tc.Args, "<csv>") {
				tc.Args = strings.ReplaceAll(tc.Args, "<csv>", csvTestFile)
			}

			var cargs string

			//nolint:gocritic // if-else more appropriate
			if utils.Contains(tc.Setup, "golden") {
				cargs = tc.Args + " --no-color " + pathArgs
			} else if strings.Contains(tc.Args, "-") {
				cargs = tc.Args + " --json " + pathArgs
			} else {
				cargs = tc.Args + " " + pathArgs
			}

			args := parseArgs(t, tc.Name, cargs)

			result, err := newTestRun(args)
			if err != nil {
				if len(tc.Conflicts) == 0 &&
					!utils.Contains(tc.Setup, "golden") {
					t.Log(string(result))
					t.Fatal(err)
				}
			}

			if tc.DefaultOpts != "" {
				os.Setenv(f2.EnvDefaultOpts, "")
			}

			for k, v := range tc.Conflicts {
				for j, v2 := range v {
					tc.Conflicts[k][j].Target = filepath.Join(
						testDir,
						v2.Target,
					)
					for l, v3 := range v2.Sources {
						v3 = filepath.Join(testDir, v3)
						tc.Conflicts[k][j].Sources[l] = v3
					}
				}
			}

			if utils.Contains(tc.Setup, "golden") {
				g := goldie.New(
					t,
					goldie.WithFixtureDir(testFixtures),
				)

				g.Assert(t, tc.GoldenFile, result)
			} else {
				jsonTest(t, &tc, result)
			}
		})
	}
}

func jsonTest(t *testing.T, tc *TestCase, result []byte) {
	t.Helper()

	var output f2.JSONOutput

	err := json.Unmarshal(result, &output)
	if err != nil {
		t.Fatal(err)
	}

	if len(tc.Conflicts) > 0 {
		if !cmp.Equal(
			tc.Conflicts,
			output.Conflicts,
		) {
			t.Fatalf(
				"Test (%s) — Expected: %+v, got: %+v\n",
				tc.Name,
				tc.Conflicts,
				output.Conflicts,
			)
		}

		return
	}

	sortChanges(tc.Changes)
	sortChanges(output.Changes)

	if !cmp.Equal(
		tc.Changes,
		output.Changes,
		cmpopts.IgnoreUnexported(f2.Change{}),
	) &&
		len(tc.Changes) != 0 {
		t.Fatalf(
			"Test (%s) -> Expected results to be: %s, but got: %s\n",
			tc.Name,
			utils.PrettyPrint(tc.Changes),
			utils.PrettyPrint(output.Changes),
		)
	}
}

func TestAllOSes(t *testing.T) {
	cases := retrieveTestCases(t, "all.json")
	runTestCases(t, cases)
}

func TestShortHelp(t *testing.T) {
	help := f2.ShortHelp(f2.NewApp())

	g := goldie.New(t, goldie.WithFixtureDir(testFixtures))
	g.Assert(t, "help", []byte(help))
}
