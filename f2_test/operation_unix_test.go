//go:build !windows
// +build !windows

package f2_test

import "testing"

// dummy function necessary for compilation in Unix.
func setHidden(path string) error {
	return nil
}

func TestUnix(t *testing.T) {
	cases := retrieveTestCases(t, "unix.json")
	runTestCases(t, cases)
}
