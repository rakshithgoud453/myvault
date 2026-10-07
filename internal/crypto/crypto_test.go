package crypto_test

import (
	"testing"

	"github.com/rakshithgoud453/myvault/internal/crypto"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	passphrase := "correct-horse-battery-staple"
	plaintext := []byte("secret-vault-contents-12345")

	ciphertext, err := crypto.Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Fatalf("ciphertext is empty")
	}

	decrypted, err := crypto.Decrypt(ciphertext, passphrase)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("expected decrypted %q, got %q", string(plaintext), string(decrypted))
	}
}

func TestDecryptWrongPassphrase(t *testing.T) {
	passphrase := "correct-passphrase"
	plaintext := []byte("secret-payload")

	ciphertext, err := crypto.Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = crypto.Decrypt(ciphertext, "wrong-passphrase")
	if err == nil {
		t.Fatalf("expected decryption error with wrong passphrase, but succeeded")
	}
}

func TestDecryptCorruptedData(t *testing.T) {
	passphrase := "my-passphrase"
	corruptedCiphertext := []byte("not-an-age-encrypted-payload")

	_, err := crypto.Decrypt(corruptedCiphertext, passphrase)
	if err == nil {
		t.Fatalf("expected decryption error for corrupted data, but succeeded")
	}
}
