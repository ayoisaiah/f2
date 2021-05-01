// +build !windows

package f2

func isHidden(filename, baseDir string) (bool, error) {
	if filename[0] == dotCharacter {
		return true, nil
	}

	return false, nil
}
