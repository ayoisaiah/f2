// +build linux

package f2

import (
	"path/filepath"
	"testing"
)

func TestCaseConversion(t *testing.T) {
	testDir := setupFileSystem(t)

	cases := []testCase{
		{
			name: "Convert pdf or epub to uppercase",
			want: []Change{
				{
					Source:  "abc.pdf",
					BaseDir: testDir,
					Target:  "abc.PDF",
				},
				{
					Source:  "abc.epub",
					BaseDir: testDir,
					Target:  "abc.EPUB",
				},
			},
			args: []string{"-f", "pdf|epub", "-r", `\Cu`, testDir},
		},
		{
			name: "Convert JPG to lowercase",
			want: []Change{
				{
					Source:  "123.JPG",
					BaseDir: filepath.Join(testDir, "images", "pics"),
					Target:  "123.jpg",
				},
			},
			args: []string{"-f", "JPG", "-r", `\Cl`, "-R", testDir},
		},
		{
			name: "Convert abc to title case",
			want: []Change{
				{
					Source:  "abc.epub",
					BaseDir: testDir,
					Target:  "Abc.epub",
				},
				{
					Source:  "abc.pdf",
					BaseDir: testDir,
					Target:  "Abc.pdf",
				},
			},
			args: []string{"-f", "abc", "-r", `\Ct`, testDir},
		},
	}

	runFindReplace(t, cases)
}
