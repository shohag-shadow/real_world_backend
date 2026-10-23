package services

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	memory      = 64 * 1024
	iterations  = 3
	parallelism = 4
	saltLength  = 16
	keyLength   = 32
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)
	base64salt := base64.RawStdEncoding.EncodeToString(salt)
	base64hash := base64.RawStdEncoding.EncodeToString(hash)
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d$t=%d$p=%d$%s$%s", argon2.Version, memory, iterations, parallelism, base64salt, base64hash)
	return encodedHash, nil
}

func VerifyPassword(password string, hash string) bool {
	parts := strings.Split(hash, "$")
	saltdecoded, err := base64.RawStdEncoding.DecodeString(parts[6])
	if err != nil {
		return false
	}
	hashdecoded, hasherr := base64.RawStdEncoding.DecodeString(parts[7])
	if hasherr != nil {
		return false
	}
	newHash := argon2.IDKey([]byte(password), saltdecoded, iterations, memory, parallelism, keyLength)
	return subtle.ConstantTimeCompare(newHash, hashdecoded) == 1
}
