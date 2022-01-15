//go:build !windows
// +build !windows

package f2

import (
	"testing"
)

func TestAutoDir(t *testing.T) {
	testDir := setupFileSystem(t)
	cases := []testCase{
		{
			name: "Auto create necessary dir1 and dir2 directories",
			want: []Change{
				{
					Source:  "abc.pdf",
					BaseDir: testDir,
					Target:  "dir1/dir2/abc.pdf",
				},
				{
					Source:  "abc.epub",
					BaseDir: testDir,
					Target:  "dir1/dir2/abc.epub",
				},
			},
			args: "-f (abc) -r dir1/dir2/$1 -x " + testDir,
		},
	}

	runFindReplaceHelper(t, cases)
}

// func TestShortHelp(t *testing.T) {
// 	help := shortHelp(newApp())
//
// 	g := goldie.New(t, goldie.WithFixtureDir(fixtures))
// 	g.Assert(t, "help", []byte(help))
// }
