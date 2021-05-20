// +build !windows

package f2

func isHidden(filename, baseDir string) (bool, error) {
	return filename[0] == dotCharacter, nil
}
