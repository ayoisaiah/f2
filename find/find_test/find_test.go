package find_test

import (
	"testing"

	"github.com/ayoisaiah/f2/find"
	"github.com/ayoisaiah/f2/internal/testutil"
)

var findFileSystem = []string{
	"backup/archive.zip",
	"backup/documents/.hidden_old_resume.txt",
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
	"photos/vacation/mountains/.hidden_old_photo.jpg",
	"photos/vacation/mountains/photo1.jpg",
	"photos/vacation/mountains/old_photo2.jpg",
	"photos/vacation/mountains/OLDPHOTO3.JPG",
	"photos/vacation/mountains/photo4.webp",
	"photos/vacation/mountains/OLD_PHOTO5.JPG",
	"photos/vacation/mountains/Öffnen.txt",
	"projects/project1/README.md",
	"projects/project1/index.html",
	"projects/project1/styles/main.css",
	"projects/project2/CHANGELOG.txt",
	"projects/project2/assets/logo.png",
	"projects/project2/index.html",
	"projects/project3/src/main.java",
	"videos/funny_cats.mp4",
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
			"photos/vacation/mountains/photo1.jpg",
		},
		Args: []string{"-f", "photo", "-R", "-E", "^old", "-E", "webp$"},
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
			"videos/funny_cats.mp4",
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
		Name: "match files at the top level",
		Want: []string{"LICENSE.txt", "Makefile", "README.md", "main.go"},
		Args: []string{"-f", ".*"},
	},
}

func findTest(t *testing.T, cases []testutil.TestCase) {
	t.Helper()

	testDir := testutil.SetupFileSystem(t, "find", findFileSystem)

	for i := range cases {
		tc := cases[i]

		t.Run(tc.Name, func(t *testing.T) {
			testutil.UpdateBaseDir(tc.Want, testDir)

			config := testutil.GetConfig(t, tc.Args, testDir)

			changes, err := find.Find(config)
			if err != nil {
				t.Fatal(err)
			}

			testutil.CompareSourcePath(t, tc.Want, changes)
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
