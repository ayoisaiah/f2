package rename_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/testutil"
	"github.com/ayoisaiah/f2/rename"
)

func renameTest(t *testing.T, cases []testutil.TestCase) {
	t.Helper()

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	for i := range cases {
		tc := cases[i]

		conf := testutil.GetConfig(t, &tc, ".")

		baseDirPath, err := os.MkdirTemp(".", "f2_test")
		if err != nil {
			t.Fatal(err)
		}

		err = os.Chdir(baseDirPath)
		if err != nil {
			t.Fatal(err)
		}

		for j := range tc.Changes {
			ch := tc.Changes[j]

			cases[i].Changes[j].SourcePath = filepath.Join(
				ch.BaseDir,
				ch.Source,
			)
			cases[i].Changes[j].TargetPath = filepath.Join(
				ch.TargetDir,
				ch.Target,
			)

			f, err := os.Create(ch.Source)
			if err != nil {
				t.Fatal(err)
			}

			err = f.Close()
			if err != nil {
				t.Fatal(err)
			}
		}

		t.Run(tc.Name, func(t *testing.T) {
			err := rename.Rename(conf, tc.Changes)
			if err != nil {
				t.Fatal(err)
			}

			for j := range tc.Changes {
				ch := tc.Changes[j]

				if _, err := os.Stat(ch.TargetPath); err != nil {
					t.Fatal(err)
				}
			}
		})

		err = os.Chdir(workingDir)
		if err != nil {
			t.Fatal(err)
		}

		err = os.RemoveAll(baseDirPath)
		if err != nil {
			t.Log(err)
		}
	}
}

func TestRename(t *testing.T) {
	testCases := []testutil.TestCase{
		{
			Name: "rename a file",
			Changes: file.Changes{
				{
					Source: "File.txt",
					Target: "myFile.txt",
				},
			},
		},
		{
			Name: "rename multiple files",
			Changes: file.Changes{
				{
					Source: "File1.txt",
					Target: "myFile1.txt",
				},
				{
					Source: "File2.jpg",
					Target: "myImage2.jpg",
				},
			},
		},
		{
			Name: "rename with case change",
			Changes: file.Changes{
				{
					Source: "file.txt",
					Target: "FILE.txt",
				},
			},
		},
		{
			Name: "rename with new directory",
			Changes: file.Changes{
				{
					Source: "File.txt",
					Target: "new_folder/myFile.txt",
				},
			},
		},
		{
			Name: "rename with a different target directory",
			Changes: file.Changes{
				{
					Source:    "File.txt",
					Target:    "myFile.txt",
					TargetDir: "one/two",
				},
			},
			Args: []string{"-f", "", "--target-dir", "one/two"},
		},
	}

	renameTest(t, testCases)
}

func postRename(t *testing.T, cases []testutil.TestCase) {
	t.Helper()

	for i := range cases {
		tc := cases[i]

		testutil.UpdateFileChanges(tc.Changes)

		var stderr bytes.Buffer

		var backup bytes.Buffer

		config.Stderr = &stderr

		t.Run(tc.Name, func(t *testing.T) {
			conf := testutil.GetConfig(t, &tc, ".")

			conf.BackupLocation = &backup

			rename.PostRename(conf, tc.Changes, tc.Error)

			tc.SnapShot.Stdout = backup.Bytes()
			tc.SnapShot.Stderr = stderr.Bytes()

			testutil.CompareGoldenFile(t, &tc)
		})
	}
}

func TestPostRename(t *testing.T) {
	testCases := []testutil.TestCase{
		{
			Name: "rename a file",
			Changes: file.Changes{
				{
					Source: "File.txt",
					Target: "myFile.txt",
				},
			},
			StdoutGoldenFile: "rename_a_file_backup",
			StderrGoldenFile: "rename_a_file_backup_stderr",
			Args:             []string{"-r", "", "-V"},
		},
	}

	postRename(t, testCases)
}
