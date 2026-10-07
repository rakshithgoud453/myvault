// Package crypto wraps the filippo.io/age library to provide
// passphrase-based encryption and decryption of the vault file.
//
// MyVault does NOT implement custom cryptography.
// All encryption/decryption is delegated to the age library.
package crypto

import (
	"bytes"
	"fmt"
	"io"

	"filippo.io/age"
)

// Decrypt decrypts age-encrypted ciphertext using the given passphrase.
// Returns the decrypted plaintext bytes.
func Decrypt(ciphertext []byte, passphrase string) ([]byte, error) {
	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return nil, fmt.Errorf("creating scrypt identity: %w", err)
	}

	reader, err := age.Decrypt(bytes.NewReader(ciphertext), identity)
	if err != nil {
		return nil, fmt.Errorf("decrypting vault: %w", err)
	}

	plaintext, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("reading decrypted data: %w", err)
	}

	return plaintext, nil
}

// Encrypt encrypts plaintext using age passphrase-based encryption.
// Returns the encrypted ciphertext bytes.
func Encrypt(plaintext []byte, passphrase string) ([]byte, error) {
	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return nil, fmt.Errorf("creating scrypt recipient: %w", err)
	}

	var buf bytes.Buffer
	writer, err := age.Encrypt(&buf, recipient)
	if err != nil {
		return nil, fmt.Errorf("initializing encryption: %w", err)
	}

	if _, err := writer.Write(plaintext); err != nil {
		return nil, fmt.Errorf("writing encrypted data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("finalizing encryption: %w", err)
	}

	return buf.Bytes(), nil
}
