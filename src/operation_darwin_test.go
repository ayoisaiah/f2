//go:build darwin
// +build darwin

package f2

import "testing"

func TestDarwin(t *testing.T) {
	cases := h2(t, "darwin.json")
	h(t, cases)
}
