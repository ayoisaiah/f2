package find

import (
	"path/filepath"
	"testing"
)

func TestIsMaxDepth(t *testing.T) {
	cases := []struct {
		Name        string
		RootPath    string
		CurrentPath string
		MaxDepth    int
		Expected    bool
	}{
		{
			Name:        "current path is on same level as root path",
			RootPath:    "/testdata/images",
			CurrentPath: "/testdata/images/bike.jpg",
			MaxDepth:    -1,
			Expected:    false,
		},
		{
			Name:        "current path is 1 level below root path",
			RootPath:    "/testdata/images",
			CurrentPath: "/testdata/images/jpegs/bike.jpg",
			MaxDepth:    -1,
			Expected:    true,
		},
		{
			Name:        "infinite recursion means no max depth",
			RootPath:    "/testdata/images",
			CurrentPath: "/testdata/images/jpegs/bike.jpg",
			MaxDepth:    0,
			Expected:    false,
		},
		{
			Name:        "max depth value exceeded by 1",
			RootPath:    "/testdata/images",
			CurrentPath: "/testdata/images/jpegs/unsplash/download/bike.jpg",
			MaxDepth:    2,
			Expected:    true,
		},
		{
			Name:        "max depth value is equal to 3",
			RootPath:    "/testdata/images",
			CurrentPath: "/testdata/images/jpegs/unsplash/download/bike.jpg",
			MaxDepth:    3,
			Expected:    false,
		},
	}

	for i := range cases {
		tc := cases[i]

		t.Run(tc.Name, func(t *testing.T) {
			// Ensure os-specifc separators are used
			rootPath, currentPath := filepath.Join(
				tc.RootPath,
			), filepath.Join(
				tc.CurrentPath,
			)

			got := isMaxDepth(rootPath, currentPath, tc.MaxDepth)

			if got != tc.Expected {
				t.Fatalf(
					"expected max depth to be: %t, but got: %t",
					tc.Expected,
					got,
				)
			}
		})
	}
}
