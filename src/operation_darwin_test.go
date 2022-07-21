//go:build darwin
// +build darwin

package f2

import "testing"

func TestDarwin(t *testing.T) {
	cases := retrieveTestCases(t, "darwin.json")
	runTestCases(t, cases)
}
