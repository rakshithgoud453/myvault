package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/rakshithgoud453/myvault/internal/crypto"
	"github.com/rakshithgoud453/myvault/internal/vault"
)

func TestEncryptedExportAndImport(t *testing.T) {
	// Build a sample vault
	v := vault.New()
	v.SetField("GITHUB", "username", "octocat", false)
	v.SetField("GITHUB", "password", "supersecret123", true)
	v.SetTags("GITHUB", []string{"work", "git"})

	plaintext, err := v.Serialize()
	if err != nil {
		t.Fatalf("v.Serialize failed: %v", err)
	}

	passphrase := "export-test-pass"

	// Test Age Encryption
	encryptedData, err := crypto.Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatalf("crypto.Encrypt failed: %v", err)
	}

	if !bytes.HasPrefix(encryptedData, []byte("age-encryption.org")) {
		t.Errorf("expected age-encryption.org header prefix")
	}

	// Test Age Decryption
	decryptedData, err := crypto.Decrypt(encryptedData, passphrase)
	if err != nil {
		t.Fatalf("crypto.Decrypt failed: %v", err)
	}

	importedVault, err := vault.Parse(decryptedData)
	if err != nil {
		t.Fatalf("vault.Parse failed: %v", err)
	}

	entry, canonName, err := importedVault.Get("GITHUB")
	if err != nil {
		t.Fatalf("importedVault.Get failed: %v", err)
	}

	if canonName != "GITHUB" {
		t.Errorf("expected GITHUB, got %s", canonName)
	}

	if entry.Fields["username"].Value != "octocat" {
		t.Errorf("expected octocat, got %s", entry.Fields["username"].Value)
	}

	if entry.Fields["password"].Value != "supersecret123" {
		t.Errorf("expected supersecret123, got %s", entry.Fields["password"].Value)
	}
}

func TestFileExportImportCycle(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "myvault-cli-test-*")
	if err != nil {
		t.Fatalf("os.MkdirTemp failed: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	exportPath := filepath.Join(tmpDir, "vault_export.enc")

	v := vault.New()
	v.SetField("AWS", "access_key", "AKIAIOSFODNN7EXAMPLE", false)
	v.SetField("AWS", "secret_key", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", true)

	plaintext, err := v.Serialize()
	if err != nil {
		t.Fatalf("v.Serialize failed: %v", err)
	}

	passphrase := "my-secret-passphrase"
	ciphertext, err := crypto.Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatalf("crypto.Encrypt failed: %v", err)
	}

	if err := os.WriteFile(exportPath, ciphertext, 0600); err != nil {
		t.Fatalf("writing export file failed: %v", err)
	}

	// Verify reading and decrypting back
	fileData, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("reading export file failed: %v", err)
	}

	decrypted, err := crypto.Decrypt(fileData, passphrase)
	if err != nil {
		t.Fatalf("crypto.Decrypt failed: %v", err)
	}

	parsed, err := vault.Parse(decrypted)
	if err != nil {
		t.Fatalf("vault.Parse failed: %v", err)
	}

	if len(parsed.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(parsed.Entries))
	}
}
