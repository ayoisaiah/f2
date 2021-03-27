// +build windows

package f2

import (
	"path/filepath"
	"syscall"
	"testing"
)

func setHidden(path string) error {
	filenameW, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	err = syscall.SetFileAttributes(filenameW, syscall.FILE_ATTRIBUTE_HIDDEN)
	if err != nil {
		return err
	}

	return nil
}

func TestHiddenWindows(t *testing.T) {
	testDir := setupFileSystem(t)
	err := setHidden(filepath.Join(testDir, "images"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	cases := []testCase{
		{
			name: "Hidden files are ignored by default",
			want: []Change{
				{
					Source:  "abc.pdf",
					BaseDir: testDir,
					Target:  "321.pdf",
				},
				{
					Source:  "abc.epub",
					BaseDir: testDir,
					Target:  "321.epub",
				},
				{
					Source:  "abc.txt",
					BaseDir: filepath.Join(testDir, "conflicts"),
					Target:  "321.txt",
				},
			},
			args: []string{"-f", "abc", "-r", "321", "-R", testDir},
		},
		{
			name: "Hidden files are allowed",
			want: []Change{
				{
					Source:  "abc.pdf",
					BaseDir: testDir,
					Target:  "321.pdf",
				},
				{
					Source:  "abc.epub",
					BaseDir: testDir,
					Target:  "321.epub",
				},
				{
					Source:  "abc.txt",
					BaseDir: filepath.Join(testDir, "conflicts"),
					Target:  "321.txt",
				},
				{
					Source:  "abc.png",
					BaseDir: filepath.Join(testDir, "images"),
					Target:  "321.png",
				},
			},
			args: []string{"-f", "abc", "-r", "321", "-H", "-R", testDir},
		},
	}

	runFindReplace(t, cases)
}
