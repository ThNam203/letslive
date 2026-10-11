package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

type argon2Params struct {
	memoryKiB uint32
	passes    uint32
	threads   uint8
	saltLen   uint32
	keyLen    uint32
}

// OWASP's baseline for Argon2id
var currentArgon2Params = argon2Params{
	memoryKiB: 19 * 1024,
	passes:    2,
	threads:   1,
	saltLen:   16,
	keyLen:    32,
}

const (
	argon2idPrefix      = "$argon2id$"
	bcryptMaxInputBytes = 72
)

var errMalformedHash = errors.New("malformed argon2id hash")

// HashPassword returns an Argon2id hash in the standard PHC string format.
func HashPassword(password string) (string, error) {
	p := currentArgon2Params
	salt := make([]byte, p.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	return encodeArgon2id(p, salt, argon2IDKey(password, p, salt)), nil
}

// VerifyPassword reports whether password matches hash, and whether the hash
// should be replaced: bcrypt hashes from before Argon2id, and Argon2id hashes
// made with other parameters.
func VerifyPassword(hash, password string) (match bool, needsRehash bool) {
	if !strings.HasPrefix(hash, argon2idPrefix) {
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
			return false, false
		}
		// bcrypt ignores everything past byte 72, so a longer input may carry
		// extra characters the real password doesn't have; rehashing it would
		// lock those characters in and the real password would stop working
		return true, len(password) <= bcryptMaxInputBytes
	}

	p, salt, key, err := decodeArgon2id(hash)
	if err != nil {
		return false, false
	}
	if subtle.ConstantTimeCompare(key, argon2IDKey(password, p, salt)) != 1 {
		return false, false
	}
	return true, p != currentArgon2Params
}

func argon2IDKey(password string, p argon2Params, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, p.passes, p.memoryKiB, p.threads, p.keyLen)
}

func encodeArgon2id(p argon2Params, salt, key []byte) string {
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.memoryKiB, p.passes, p.threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key))
}

func decodeArgon2id(hash string) (argon2Params, []byte, []byte, error) {
	// "", "argon2id", "v=19", "m=..,t=..,p=..", salt, key
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return argon2Params{}, nil, nil, errMalformedHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return argon2Params{}, nil, nil, errMalformedHash
	}

	var p argon2Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memoryKiB, &p.passes, &p.threads); err != nil {
		return argon2Params{}, nil, nil, errMalformedHash
	}
	if p.memoryKiB == 0 || p.passes == 0 || p.threads == 0 {
		return argon2Params{}, nil, nil, errMalformedHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) == 0 {
		return argon2Params{}, nil, nil, errMalformedHash
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(key) == 0 {
		return argon2Params{}, nil, nil, errMalformedHash
	}

	p.saltLen = uint32(len(salt))
	p.keyLen = uint32(len(key))
	return p, salt, key, nil
}
