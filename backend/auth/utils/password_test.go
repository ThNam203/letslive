package utils

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashAndVerifyPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"ascii", "Password123!"},
		{"72 ascii", "Aa!" + strings.Repeat("x", 69)},
		{"72 vietnamese (213 bytes)", "Ữữ!" + strings.Repeat("ữ", 69)},
		{"emoji (over 72 bytes)", "Aa!" + strings.Repeat("😀", 34) + "x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if err != nil {
				t.Fatalf("HashPassword: %v", err)
			}
			if !strings.HasPrefix(hash, "$argon2id$v=19$m=19456,t=2,p=1$") {
				t.Fatalf("unexpected hash format %q", hash)
			}
			if match, rehash := VerifyPassword(hash, tt.password); !match || rehash {
				t.Fatalf("right password: match=%v rehash=%v, want true false", match, rehash)
			}
			if match, _ := VerifyPassword(hash, tt.password+"x"); match {
				t.Fatal("a different password matched")
			}
		})
	}
}

func TestLongPasswordsAreNotCutAt72Bytes(t *testing.T) {
	prefix := strings.Repeat("ữ", 40) // 120 bytes
	hash, err := HashPassword(prefix + "A")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if match, _ := VerifyPassword(hash, prefix+"B"); match {
		t.Fatal("passwords differing after byte 72 matched")
	}
}

func TestHashesAreSalted(t *testing.T) {
	a, _ := HashPassword("Password123!")
	b, _ := HashPassword("Password123!")
	if a == b {
		t.Fatal("two hashes of the same password are identical")
	}
}

func TestBcryptHashesVerifyAndAskForRehash(t *testing.T) {
	for _, password := range []string{"Password123!", "Ữữ@ữ" + strings.Repeat("a", 10), "Aa!" + strings.Repeat("x", 69)} {
		legacy, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
		if err != nil {
			t.Fatalf("bcrypt: %v", err)
		}
		if match, rehash := VerifyPassword(string(legacy), password); !match || !rehash {
			t.Fatalf("bcrypt hash of %q: match=%v rehash=%v, want true true", password, match, rehash)
		}
		if match, _ := VerifyPassword(string(legacy), "x"+password[1:]); match {
			t.Fatalf("bcrypt hash of %q matched a different password", password)
		}
	}
}

// bcrypt matches any input whose first 72 bytes are right. Such a login keeps
// working as before, but must not be rehashed, or the extra characters would
// become part of the password and the real one would stop working.
func TestBcryptMatchPast72BytesIsNotRehashed(t *testing.T) {
	password := strings.Repeat("ữ", 24) // exactly 72 bytes
	legacy, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}

	match, rehash := VerifyPassword(string(legacy), password+"typo")
	if !match {
		t.Fatal("bcrypt behaviour changed: an input with the right first 72 bytes no longer matches")
	}
	if rehash {
		t.Fatal("an input longer than 72 bytes would be rehashed and lock the user out")
	}
}

func TestOutdatedArgon2ParamsAskForRehash(t *testing.T) {
	old := argon2Params{memoryKiB: 8 * 1024, passes: 1, threads: 1, saltLen: 16, keyLen: 32}
	salt := []byte("0123456789abcdef")
	hash := encodeArgon2id(old, salt, argon2IDKey("Password123!", old, salt))

	if match, rehash := VerifyPassword(hash, "Password123!"); !match || !rehash {
		t.Fatalf("match=%v rehash=%v, want true true", match, rehash)
	}
}

func TestMalformedHashesNeverMatch(t *testing.T) {
	valid, _ := HashPassword("Password123!")
	parts := strings.Split(valid, "$")

	hashes := map[string]string{
		"empty (google account)": "",
		"garbage":                "not a hash",
		"prefix only":            "$argon2id$",
		"wrong version":          strings.Replace(valid, "v=19", "v=16", 1),
		"zero memory":            strings.Replace(valid, "m=19456", "m=0", 1),
		"bad salt":               strings.Join([]string{parts[0], parts[1], parts[2], parts[3], "!!!", parts[5]}, "$"),
		"bad key":                strings.Join([]string{parts[0], parts[1], parts[2], parts[3], parts[4], "!!!"}, "$"),
		"missing key":            strings.Join(parts[:5], "$"),
	}
	for name, hash := range hashes {
		t.Run(name, func(t *testing.T) {
			if match, rehash := VerifyPassword(hash, "Password123!"); match || rehash {
				t.Fatalf("match=%v rehash=%v, want false false", match, rehash)
			}
		})
	}
}
