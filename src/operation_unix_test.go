//go:build !windows
// +build !windows

package f2

import "testing"

func TestUnix(t *testing.T) {
	cases := h2(t, "unix.json")
	h(t, cases)
}
