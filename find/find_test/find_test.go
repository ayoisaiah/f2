package find_test

import (
	"errors"
	"os"
	"testing"

	"github.com/ayoisaiah/f2/find"
	"github.com/ayoisaiah/f2/internal/testutil"
)

var findFileSystem = []string{
	"backup/archive.zip",
	"backup/documents/.hidden_resume.txt",
	"backup/documents/old_cover_letter.docx",
	"backup/documents/old_resume.docx",
	"backup/important_data/file1.txt",
	"backup/important_data/file2.txt",
	"backup/photos/family/old_photo1.jpg",
	"backup/photos/family/old_photo2.jpg",
	"documents/.hidden_file.txt",
	"documents/UPPERCASE_FILE.txt",
	"documents/cover_letter.docx",
	"documents/resume.docx",
	".hidden_file",
	"LICENSE.txt",
	"Makefile",
	"README.md",
	"main.go",
	"photos/family/Photo1.jpg",
	"photos/family/photo2.PNG",
	"photos/family/photo3.gif",
	"photos/vacation/beach.jpg",
	"photos/vacation/mountains/.hidden_photo.jpg",
	"photos/vacation/mountains/OLDPHOTO3.JPG",
	"photos/vacation/mountains/OLD_PHOTO5.JPG",
	"photos/vacation/mountains/photo1.jpg",
	"photos/vacation/mountains/old_photo2.jpg",
	"photos/vacation/mountains/photo4.webp",
	"photos/vacation/mountains/Öffnen.txt",
	"projects/project1/README.md",
	"projects/project1/index.html",
	"projects/project1/styles/main.css",
	"projects/project2/CHANGELOG.txt",
	"projects/project2/assets/logo (1).png",
	"projects/project2/index.html",
	"projects/project3/src/main.java",
	"videos/funny_cats (3).mp4",
	"videos/tutorials/GoLang.mp4",
	"videos/tutorials/JavaScript.mp4",
}

var testCases = []testutil.TestCase{
	{
		Name: "include directories in search",
		Want: []string{"projects"},
		Args: []string{"-f", "projects", "-d"},
	},

	{
		Name: "include directories and infinitely recurse",
		Want: []string{
			"projects",
			"projects/project1",
			"projects/project2",
			"projects/project3",
		},
		Args: []string{"-f", "project", "-dR"},
	},

	{
		Name: "match recursively up to a maximum depth",
		Want: []string{
			"photos/family/photo2.PNG",
			"photos/family/photo3.gif",
		},
		Args: []string{"-f", "photo", "-R", "-m", "2"},
	},

	{
		Name: "match recursively but exclude certain patterns",
		Want: []string{
			"photos/family/photo2.PNG",
			"photos/family/photo3.gif",
		},
		Args: []string{
			"-f",
			"photo",
			"-R",
			"--exclude-dir",
			"backup",
			"-exclude-dir",
			"mountains",
		},
	},

	{
		Name: "match recursively but exclude certain directories",
		Want: []string{
			"photos/family/photo2.PNG",
			"photos/family/photo3.gif",
			"photos/vacation/mountains/photo1.jpg",
		},
		Args:      []string{"-f", "photo", "-R", "-E", "^old", "-E", "webp$"},
		SetupFunc: setupWindowsHidden,
	},

	{
		Name: "match only directories",
		Want: []string{"backup/photos", "photos"},
		Args: []string{"-f", "photo", "-R", "-D"},
	},

	{
		Name: "ignore the file extension",
		Want: []string{
			"backup/archive.zip",
			"backup/documents/old_cover_letter.docx",
			"documents/cover_letter.docx",
			"photos/vacation/beach.jpg",
			"videos/funny_cats (3).mp4",
			"videos/tutorials/JavaScript.mp4",
		},
		Args: []string{"-f", "c", "-Re"},
	},

	{
		Name: "match all the directories at the top level",
		Want: []string{"backup", "documents", "photos", "projects", "videos"},
		Args: []string{"-f", ".*", "-D"},
	},

	{
		Name:      "match files at the top level",
		Want:      []string{"LICENSE.txt", "Makefile", "README.md", "main.go"},
		Args:      []string{"-f", ".*"},
		SetupFunc: setupWindowsHidden,
	},

	{
		Name: "ignore text casing in search",
		Want: []string{
			"backup/photos/family/old_photo1.jpg",
			"backup/photos/family/old_photo2.jpg",
			"photos/vacation/mountains/OLDPHOTO3.JPG",
			"photos/vacation/mountains/OLD_PHOTO5.JPG",
			"photos/vacation/mountains/old_photo2.jpg",
		},
		Args: []string{"-f", "old", "-Ri", "-E", "docx"},
	},

	{
		Name: "match regex special characters with escaping",
		Want: []string{
			"projects/project2/assets/logo (1).png",
			"videos/funny_cats (3).mp4",
		},
		Args: []string{"-f", "\\(\\d+\\)", "-R"},
	},

	{
		Name: "match regex special characters without escaping",
		Want: []string{
			"projects/project2/assets/logo (1).png",
			"videos/funny_cats (3).mp4",
		},
		Args: []string{"-f", "(", "-Rs"},
	},

	{
		Name: "match any all uppercase filenames",
		Want: []string{
			"LICENSE.txt",
			"README.md",
			"projects/project1/README.md",
			"projects/project2/CHANGELOG.txt",
		},
		Args: []string{"-f", "^[A-Z]+$", "-Re"},
	},

	{
		Name: "match files not containing a dot",
		Want: []string{
			"Makefile",
		},
		Args: []string{"-f", "^[^.]+$", "-R"},
	},

	{
		Name: "match files containing an umulat",
		Want: []string{
			"photos/vacation/mountains/Öffnen.txt",
		},
		Args: []string{"-f", "[äöüÄÖÜ]", "-R"},
	},

	{
		Name: "max depth should have no effect without recursing",
		Want: []string{},
		Args: []string{"-f", "jpg", "-m", "4"},
	},

	{
		Name: "find matches in specific directory argument",
		Want: []string{
			"documents/cover_letter.docx",
			"documents/resume.docx",
		},
		PathArgs: []string{"documents"},
		Args:     []string{"-f", "\\.docx$"},
	},

	{
		Name: "find matches in only specific file paths",
		Want: []string{
			"photos/vacation/mountains/photo1.jpg",
			"photos/vacation/beach.jpg",
		},
		PathArgs: []string{
			"photos/vacation/mountains/photo1.jpg",
			"photos/vacation/beach.jpg",
		},
		Args: []string{"-f", "jpg"},
	},

	{
		Name:  "expect error when non-existent file path is provided",
		Error: os.ErrNotExist,
		Want: []string{
			"photos/vacation/mountains/photo1.jpg",
			"photos/vacation/beach.jpg",
		},
		PathArgs: []string{
			"photos/vacation/mountains/photo1.jpg",
			"nonexistent.jpg",
		},
		Args: []string{"-f", "jpg"},
	},
}

func findTest(t *testing.T, cases []testutil.TestCase) {
	t.Helper()

	testDir := testutil.SetupFileSystem(t, "find", findFileSystem)

	for i := range cases {
		tc := cases[i]

		t.Run(tc.Name, func(t *testing.T) {
			testutil.UpdateBaseDir(tc.Want, testDir)

			if tc.SetupFunc != nil {
				t.Cleanup(tc.SetupFunc(t, testDir))
			}

			// TODO: Make it possible to test without explicitly providing
			// directory argument
			config := testutil.GetConfig(t, &tc, testDir)

			changes, err := find.Find(config)
			if err == nil {
				testutil.CompareSourcePath(t, tc.Want, changes)
				return
			}

			if !errors.Is(err, tc.Error) {
				t.Fatal(err)
			}
		})
	}
}

// TestFind tests how different flags affect how files are matched including
// the following:
// exclude, hidden, include-dir, only-dir, ignore-case, ignore-ext, max-depth,
// recursive, string-mode.
func TestFind(t *testing.T) {
	findTest(t, testCases)
}

// TODO: Test reverting from a backup file.
func TestLoadFromBackup(t *testing.T) {
	t.Skip("not implemented")
}
