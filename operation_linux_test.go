// +build linux

package f2

import (
	"path/filepath"
	"regexp"
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
			args: []string{"-f", "pdf|epub", "-r", `\Tcu`, testDir},
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
			args: []string{"-f", "JPG", "-r", `\Tcl`, "-R", testDir},
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
			args: []string{"-f", "abc", "-r", `\Tct`, testDir},
		},
	}

	runFindReplace(t, cases)
}

func TestTransformation(t *testing.T) {
	cases := []struct {
		input     string
		transform string
		find      string
		output    string
	}{
		{
			input:     `abc<>_{}*?\/\.epub`,
			transform: `\Twin`,
			find:      `abc.*`,
			output:    "abc_{}.epub",
		},
		{
			input:     `abc<>_{}*:?\/\.epub`,
			transform: `\Tmac`,
			find:      `abc.*`,
			output:    `abc<>_{}*?\/\.epub`,
		},
	}

	for _, v := range cases {
		op := &Operation{}
		op.replacement = v.transform
		regex, err := regexp.Compile(v.find)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		op.searchRegex = regex
		out := op.replaceString(v.input)

		if out != v.output {
			t.Fatalf("Expected %s, but got: %s", v.output, out)
		}
	}
}
