// Package password provides cryptographically secure password generation
// using Go's crypto/rand.
//
// Replaces the v0.1 dependency on openssl rand.
package password

import (
	"crypto/rand"
	"math/big"
)

const (
	// DefaultLength is the default generated password length.
	DefaultLength = 24

	// Character sets for password generation.
	lowerChars   = "abcdefghijklmnopqrstuvwxyz"
	upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars   = "0123456789"
	specialChars = "!@#$%^&*()-_=+[]{}|;:,.<>?"

	// AllChars is the full character set used for generation.
	AllChars = lowerChars + upperChars + digitChars + specialChars
)

// Generate creates a cryptographically secure random password of the given length.
// The password is guaranteed to contain at least one character from each character set
// (lowercase, uppercase, digit, special) when length >= 4.
func Generate(length int) (string, error) {
	if length <= 0 {
		length = DefaultLength
	}

	password := make([]byte, length)

	// Guarantee at least one from each set when long enough.
	if length >= 4 {
		sets := []string{lowerChars, upperChars, digitChars, specialChars}
		for i, set := range sets {
			c, err := randomChar(set)
			if err != nil {
				return "", err
			}
			password[i] = c
		}
		// Fill the rest randomly from the full set.
		for i := 4; i < length; i++ {
			c, err := randomChar(AllChars)
			if err != nil {
				return "", err
			}
			password[i] = c
		}
	} else {
		for i := 0; i < length; i++ {
			c, err := randomChar(AllChars)
			if err != nil {
				return "", err
			}
			password[i] = c
		}
	}

	// Shuffle to avoid predictable positions for guaranteed characters.
	if err := shuffle(password); err != nil {
		return "", err
	}

	return string(password), nil
}

// randomChar returns a cryptographically random character from the given charset.
func randomChar(charset string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	if err != nil {
		return 0, err
	}
	return charset[n.Int64()], nil
}

// shuffle performs a Fisher-Yates shuffle using crypto/rand.
func shuffle(b []byte) error {
	for i := len(b) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return err
		}
		b[i], b[j.Int64()] = b[j.Int64()], b[i]
	}
	return nil
}
