//go:build !windows
// +build !windows

package find

// checkIfHidden checks if a file is hidden on Unix operating systems
// the nil error is returned to match the signature of the Windows
// version of the function.
func checkIfHidden(filename, _ string) (bool, error) {
	return filename[0] == dotCharacter, nil
}
