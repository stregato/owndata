package core

import (
	"crypto/rand"
)

func GenerateRandomBytes(length int) []byte {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if IsErr(err, "cannot generate random bytes: %v", err) {
		panic(err)
	}
	return randomBytes
}
