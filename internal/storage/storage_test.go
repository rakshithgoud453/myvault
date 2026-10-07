package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rakshithgoud453/myvault/internal/storage"
)

func TestWriteAndReadVaultFile(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.age")

	testData := []byte("encrypted-payload-data")

	// Write vault file
	if err := storage.WriteVaultFile(vaultPath, testData); err != nil {
		t.Fatalf("WriteVaultFile failed: %v", err)
	}

	// Verify permissions
	info, err := os.Stat(vaultPath)
	if err != nil {
		t.Fatalf("os.Stat failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != storage.FilePermissions {
		t.Errorf("expected permissions %o, got %o", storage.FilePermissions, perm)
	}

	// Read vault file back
	readData, err := storage.ReadVaultFile(vaultPath)
	if err != nil {
		t.Fatalf("ReadVaultFile failed: %v", err)
	}

	if string(readData) != string(testData) {
		t.Errorf("expected content %q, got %q", string(testData), string(readData))
	}
}

func TestAtomicWriteBackup(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.age")

	originalData := []byte("v1-original-data")
	updatedData := []byte("v2-updated-data")

	// Write original
	if err := storage.WriteVaultFile(vaultPath, originalData); err != nil {
		t.Fatalf("first WriteVaultFile failed: %v", err)
	}

	// Overwrite with new data (should trigger backup)
	if err := storage.WriteVaultFile(vaultPath, updatedData); err != nil {
		t.Fatalf("second WriteVaultFile failed: %v", err)
	}

	// Check backup file exists
	backupPath := filepath.Join(tmpDir, storage.BackupDir, storage.VaultFile)
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("reading backup file failed: %v", err)
	}

	if string(backupData) != string(originalData) {
		t.Errorf("backup content expected %q, got %q", string(originalData), string(backupData))
	}

	// Check current vault file has updated data
	currentData, err := storage.ReadVaultFile(vaultPath)
	if err != nil {
		t.Fatalf("reading updated vault file failed: %v", err)
	}
	if string(currentData) != string(updatedData) {
		t.Errorf("current vault content expected %q, got %q", string(updatedData), string(currentData))
	}
}
