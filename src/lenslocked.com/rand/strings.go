package rand

import (
	"crypto/rand"
	"encoding/base64"
)

// A byte has 256 possible values so a one-byte remember token could
// only have 256 unique tokens and 2 bytes would be 65,536 possible
// combinations (256 ^2 bytes). 32 bytes has 256 ^32 combos or 1e77
// possible combos.
const RememberTokenBytes = 32

func RememberToken() (string, error) {
	return String(RememberTokenBytes)
}

// String generates a random string using the specified number of bytes
// that is base64 encoded. URL encoding is just a run of the mill encoding
// package to use with base64
func String(nBytes int) (string, error) {
	b, err := Bytes(nBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Bytes will help us generate n random bytes, or will
// return an error if there was one. This uses the
// crypto/rand package so it is safe to use with things
// like remember tokens.
func Bytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}
