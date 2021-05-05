package f2

import (
	"path/filepath"
	"regexp"
	"testing"
)

func TestFindReplace(t *testing.T) {
	testDir := setupFileSystem(t)

	cases := []testCase{
		{
			name: "Replace the last 2 matches",
			want: []Change{
				{
					Source:  "No Pressure (2021) S1.E1.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure (2021) S1.E5.5080p.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E2.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure (2021) S5.E2.5080p.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E3.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure (2021) S5.E3.5080p.mkv",
				},
			},
			args: []string{
				"-f",
				"1",
				"-r",
				"5",
				"-l",
				"-2",
				testDir,
			},
		},
		{
			name: "Replace the last match",
			want: []Change{
				{
					Source:  "No Pressure (2021) S1.E1.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure (2021) S1.E1.5080p.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E2.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure (2021) S1.E2.5080p.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E3.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure (2021) S1.E3.5080p.mkv",
				},
			},
			args: []string{
				"-f",
				"1",
				"-r",
				"5",
				"-l",
				"-1",
				testDir,
			},
		},
		{
			name: "Replace the first 10 matches",
			want: []Change{
				{
					Source:  "No Pressure (2021) S1.E1.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure (2025) S5.E5.5080p.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E2.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure (2025) S5.E2.5080p.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E3.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure (2025) S5.E3.5080p.mkv",
				},
			},
			args: []string{
				"-f",
				"1",
				"-r",
				"5",
				"-l",
				"10",
				testDir,
			},
		},
		{
			want: []Change{
				{
					Source:  "No Pressure (2021) S1.E1.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure (2025) S5.E1.1080p.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E2.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure (2025) S5.E2.1080p.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E3.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure (2025) S5.E3.1080p.mkv",
				},
			},
			args: []string{
				"-f",
				"1",
				"-r",
				"5",
				"-l",
				"2",
				testDir,
			},
		},
		{
			want: []Change{
				{
					Source:  "No Pressure (2021) S1.E1.1080p.mkv",
					BaseDir: testDir,
					Target:  "1.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E2.1080p.mkv",
					BaseDir: testDir,
					Target:  "2.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E3.1080p.mkv",
					BaseDir: testDir,
					Target:  "3.mkv",
				},
			},
			args: []string{
				"-f",
				".*E(\\d+).*",
				"-r",
				"$1.mkv",
				testDir,
			},
		},
		{
			want: []Change{
				{
					Source:  "No Pressure (2021) S1.E1.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure 98.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E2.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure 99.mkv",
				},
				{
					Source:  "No Pressure (2021) S1.E3.1080p.mkv",
					BaseDir: testDir,
					Target:  "No Pressure 100.mkv",
				},
			},
			args: []string{
				"-f",
				"(No Pressure).*",
				"-r",
				"$1 98%d.mkv",
				testDir,
			},
		},
		{
			want: []Change{
				{
					Source:  "index.js",
					BaseDir: filepath.Join(testDir, "scripts"),
					Target:  "index.ts",
				},
				{
					Source:  "main.js",
					BaseDir: filepath.Join(testDir, "scripts"),
					Target:  "main.ts",
				},
			},
			args: []string{
				"-f",
				"js",
				"-r",
				"ts",
				filepath.Join(testDir, "scripts"),
			},
		},
		{
			want: []Change{
				{
					Source:  "index.js",
					BaseDir: filepath.Join(testDir, "scripts"),
					Target:  "i n d e x .js",
				},
				{
					Source:  "main.js",
					BaseDir: filepath.Join(testDir, "scripts"),
					Target:  "m a i n .js",
				},
			},
			args: []string{
				"-f",
				"(.)",
				"-r",
				"$1 ",
				"-e",
				filepath.Join(testDir, "scripts"),
			},
		},
		{
			want: []Change{
				{
					Source:  "a.jpg",
					BaseDir: filepath.Join(testDir, "images"),
					Target:  "a.jpeg",
				},
				{
					Source:  "b.jPg",
					BaseDir: filepath.Join(testDir, "images"),
					Target:  "b.jpeg",
				},
				{
					Source:  "123.JPG",
					BaseDir: filepath.Join(testDir, "images", "pics"),
					Target:  "123.jpeg",
				},
				{
					Source:  "free.jpg",
					BaseDir: filepath.Join(testDir, "images", "pics"),
					Target:  "free.jpeg",
				},
				{
					Source:  "img.jpg",
					BaseDir: filepath.Join(testDir, "morepics", "nested"),
					Target:  "img.jpeg",
				},
			},
			args: []string{
				"-f",
				"jpg",
				"-r",
				"jpeg",
				"-R",
				"-i",
				testDir,
			},
		},
		{
			want: []Change{
				{
					Source:  "pics",
					IsDir:   true,
					BaseDir: filepath.Join(testDir, "images"),
					Target:  "images",
				},
				{
					Source:  "morepics",
					IsDir:   true,
					BaseDir: testDir,
					Target:  "moreimages",
				},
				{
					Source:  "pic-1.avif",
					BaseDir: filepath.Join(testDir, "morepics"),
					Target:  "image-1.avif",
				},
				{
					Source:  "pic-2.avif",
					BaseDir: filepath.Join(testDir, "morepics"),
					Target:  "image-2.avif",
				},
			},
			args: []string{"-f", "pic", "-r", "image", "-d", "-R", testDir},
		},
		{
			want: []Change{
				{
					Source:  "pics",
					IsDir:   true,
					BaseDir: filepath.Join(testDir, "images"),
					Target:  "images",
				},
				{
					Source:  "morepics",
					IsDir:   true,
					BaseDir: testDir,
					Target:  "moreimages",
				},
			},
			args: []string{"-f", "pic", "-r", "image", "-D", "-R", testDir},
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
		{
			input:     "žůžo.txt",
			transform: `\Td`,
			find:      "žůžo",
			output:    "zuzo.txt",
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
