package f2

import (
	"testing"
)

func TestGetNewPath(t *testing.T) {
	type m map[string][]struct {
		sourcePath string
		index      int
	}

	cases := []struct {
		input  string
		output string
		m      m
	}{
		{
			input:  "an_image.png",
			output: "an_image (2).png",
			m:      nil,
		},
		{
			input:  "an_image (2).png",
			output: "an_image (3).png",
			m:      nil,
		},
		{
			input:  "an_image (4).png",
			output: "an_image (5).png",
			m:      nil,
		},
		{
			input:  "an_image (8).png",
			output: "an_image (12).png",
			m: m{
				"an_image (8).png": {
					{
						sourcePath: "img.png",
						index:      3,
					},
				},
				"an_image (9).png": {
					{
						sourcePath: "img-2.png",
						index:      5,
					},
				},
				"an_image (10).png": {
					{
						sourcePath: "img-3.png",
						index:      8,
					},
				},
				"an_image (11).png": {
					{
						sourcePath: "img-4.png",
						index:      6,
					},
				},
			},
		},
	}

	for _, v := range cases {
		ch := Change{
			Target:  v.input,
			BaseDir: ".",
		}

		out := newTarget(&ch, v.m)
		if out != v.output {
			t.Fatalf(
				"Incorrect output from getNewPath. Want: %s, got %s",
				v.output,
				out,
			)
		}
	}
}
