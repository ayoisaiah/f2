package report_test

import (
	"bytes"
	"testing"

	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/status"
	"github.com/ayoisaiah/f2/internal/testutil"
	"github.com/ayoisaiah/f2/report"
)

func TestReport(t *testing.T) {
	testCases := []testutil.TestCase{
		{
			Name: "report unchanged file names",
			Changes: []*file.Change{
				{
					RelSourcePath: "macos_update_notes_2023.txt",
					RelTargetPath: "macos_update_notes_2023.txt",
					Status:        status.Unchanged,
				},
				{
					RelSourcePath: "macos_user_guide_macos_sierra.pdf",
					RelTargetPath: "macos_user_guide_macos_sierra.pdf",
					Status:        status.Unchanged,
				},
			},
		},
	}

	reportTest(t, testCases)
}

func reportTest(t *testing.T, cases []testutil.TestCase) {
	t.Helper()

	for i := range cases {
		tc := cases[i]

		for i := range tc.Changes {
			tc.Changes[i].Position = i
		}

		t.Run(tc.Name, func(t *testing.T) {
			if tc.SetupFunc != nil {
				t.Cleanup(tc.SetupFunc(t))
			}

			var buf bytes.Buffer
			report.Stdout = &buf
			report.NonInteractive(tc.Changes)

			testutil.CompareGoldenFile(t, &tc, buf.Bytes())
		})
	}
}
