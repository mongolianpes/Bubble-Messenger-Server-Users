package crypto

import (
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"io"
)

const (
	saltSize = 32
	iter     = 3
)

var mainSaltForPassword = []byte("?da;lmjgagma&qgmaf")
var mainSaltForUserID = []byte("g*amgo#23p65!s")

func HashString(passwordStr string, isPassword bool) (string, error) {
	password := []byte(passwordStr)

	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}
	h := sha512.New()
	h.Write(salt)
	if isPassword {
		h.Write(mainSaltForPassword)
	} else {
		h.Write(mainSaltForUserID)
	}
	h.Write(password)
	sum := h.Sum(nil)

	for i := 1; i < iter; i++ {
		h2 := sha512.New()
		h2.Write(sum)
		sum = h2.Sum(nil)
	}

	result := hex.EncodeToString(sum) + hex.EncodeToString(salt)

	return result, nil
}

func verifyHash(candidateStr, hashAndSaltHex string) bool {
	salt, err := hex.DecodeString(hashAndSaltHex[128:])
	if err != nil {
		return false
	}
	expectedHash, err := hex.DecodeString(hashAndSaltHex[:128])
	if err != nil {
		return false
	}
	candidate := []byte(candidateStr)

	h := sha512.New()
	h.Write(salt)

	h.Write(mainSaltForPassword)
	h.Write(candidate)
	sum := h.Sum(nil)

	for i := 1; i < iter; i++ {
		h2 := sha512.New()
		h2.Write(sum)
		sum = h2.Sum(nil)
	}

	if len(sum) != len(expectedHash) {
		return false
	}
	return subtle.ConstantTimeCompare(sum, expectedHash) == 1
}

func VerifyPassword(currentPassword, inputPassword string) bool {
	if !verifyHash(inputPassword, currentPassword) {
		return false
	}

	return true
}
