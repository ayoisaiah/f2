package variables

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"io"
	"os"

	"github.com/OneOfOne/xxhash"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
)

// Replace replaces a hash variable with the corresponding hash value.
func (hv hashVars) Replace(
	_ context.Context,
	conf *config.Config,
	change *file.Change,
) error {
	if len(hv.matches) == 0 {
		return nil
	}

	target, err := replaceFileHashVars(conf, change, hv)
	if err != nil {
		return err
	}

	change.Target = target

	return nil
}

// getHash retrieves the appropriate hash value for the specified file.
func getHash(filePath string, hashValue hashAlgorithm) (string, error) {
	openedFile, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	defer openedFile.Close()

	var newHash hash.Hash

	switch hashValue {
	case sha1Hash:
		newHash = sha1.New()
	case sha256Hash:
		newHash = sha256.New()
	case sha512Hash:
		newHash = sha512.New()
	case md5Hash:
		newHash = md5.New()
	case xxh32Hash:
		newHash = xxhash.New32()
	case xxh64Hash:
		newHash = xxhash.New64()
	default:
		return "", nil
	}

	if _, err := io.Copy(newHash, openedFile); err != nil {
		return "", err
	}

	return hex.EncodeToString(newHash.Sum(nil)), nil
}

// replaceFileHashVars replaces a hash variable with the corresponding
// hash value.
func replaceFileHashVars(
	conf *config.Config,
	change *file.Change,
	hashMatches hashVars,
) (string, error) {
	target := change.Target

	for i := range hashMatches.matches {
		current := hashMatches.matches[i]

		var (
			hashValue string
			err       error
		)

		if change.HashData != nil {
			if val, ok := change.HashData[string(current.hashFn)]; ok {
				hashValue = val
			}
		}

		if hashValue == "" {
			hashValue, err = getHash(change.SourcePath, current.hashFn)
			if err != nil {
				return "", err
			}

			if change.HashData == nil {
				change.HashData = make(map[string]string)
			}

			change.HashData[string(current.hashFn)] = hashValue
		}

		hashValue = transformString(conf, hashValue, current.transformToken)

		target = RegexReplace(current.regex, target, hashValue, 0, nil)
	}

	return target, nil
}
