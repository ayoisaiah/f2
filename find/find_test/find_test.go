package find_test

import (
	"testing"

	"github.com/ayoisaiah/f2/find"
	"github.com/ayoisaiah/f2/internal/testutil"
)

var findFileSystem = []string{
	"backup/archive.zip",
	"backup/important_data/file1.txt",
	"backup/important_data/file2.txt",
	"backup/documents/old_resume.docx",
	"backup/documents/old_cover_letter.docx",
	"backup/photos/family/old_photo1.jpg",
	"backup/photos/family/old_photo2.jpg",
	"backup/photos/vacation/mountains/old_photo1.jpg",
	"backup/photos/vacation/mountains/old_photo2.jpg",
	"documents/cover_letter.docx",
	"documents/references.txt",
	"documents/resume.docx",
	"projects/project1/index.html",
	"projects/project1/scripts/app.js",
	"projects/project1/styles/main.css",
	"projects/project2/index.html",
	"projects/project2/assets/logo.png",
	"projects/project3/src/main.java",
	"photos/family/photo1.jpg",
	"photos/family/photo2.jpg",
	"photos/family/photo3.jpg",
	"photos/vacation/beach.jpg",
	"photos/vacation/mountains/photo1.jpg",
	"photos/vacation/mountains/photo2.webp",
	"videos/funny_cats.mp4",
	"videos/tutorials/javascript.mp4",
	"videos/tutorials/golang.mp4",
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
			"photos/family/photo1.jpg",
			"photos/family/photo2.jpg",
			"photos/family/photo3.jpg",
		},
		Args: []string{"-f", "photo", "-R", "-m", "2"},
	},

	{
		Name: "match recursively but exclude certain patterns",
		Want: []string{
			"photos/family/photo1.jpg",
			"photos/family/photo2.jpg",
			"photos/family/photo3.jpg",
			"photos/vacation/mountains/photo1.jpg",
		},
		Args: []string{"-f", "photo", "-R", "-E", "^old", "-E", "webp$"},
	},

	{
		Name: "match only directories",
		Want: []string{"backup/photos", "photos"},
		Args: []string{"-f", "photo", "-R", "-D"},
	},
}

// TestFind tests how different flags affect how files are matched including
// the following:
// exclude, hidden, include-dir, only-dir, ignore-case, ignore-ext, max-depth,
// recursive, string-mode.
func TestFind(t *testing.T) {
	testDir := testutil.SetupFileSystem(t, "find", findFileSystem)

	for _, tc := range testCases {
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
