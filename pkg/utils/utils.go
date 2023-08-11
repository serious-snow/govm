package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func CheckSha256(fileName, sha256v string) bool {
	sha := sha256.New()
	f, err := os.Open(fileName)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = io.Copy(sha, f)
	if err != nil {
		return false
	}

	dSha256 := hex.EncodeToString(sha.Sum(nil))
	return dSha256 == sha256v
}
