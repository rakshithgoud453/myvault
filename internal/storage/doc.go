// Package storage handles vault file I/O, paths, permissions,
// unencrypted resource index management, and backup management.
//
// Default vault location: ~/.password-vault/vault.age
// Default index location: ~/.password-vault/vault.index
// Permissions: directory 700, files 600
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// VaultDir is the directory name under the user's home directory.
	VaultDir = ".password-vault"

	// VaultFile is the encrypted vault filename.
	VaultFile = "vault.age"

	// IndexFile is the unencrypted resource names index filename.
	IndexFile = "vault.index"

	// BackupDir is the backup subdirectory name.
	BackupDir = "backup"

	// DirPermissions is the permission mode for the vault directory (rwx------).
	DirPermissions = 0700

	// FilePermissions is the permission mode for vault files (rw-------).
	FilePermissions = 0600
)

// ResourceIndexEntry stores non-sensitive resource metadata for fast listing/search.
type ResourceIndexEntry struct {
	Name string   `json:"name"`
	Tags []string `json:"tags,omitempty"`
}

// VaultIndex holds the list of resource names and tags.
type VaultIndex struct {
	Resources []ResourceIndexEntry `json:"resources"`
}

// DefaultVaultPath returns the full path to the vault file.
// e.g., /Users/<user>/.password-vault/vault.age
func DefaultVaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home directory: %w", err)
	}
	return filepath.Join(home, VaultDir, VaultFile), nil
}

// DefaultIndexPath returns the full path to the unencrypted index file.
func DefaultIndexPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home directory: %w", err)
	}
	return filepath.Join(home, VaultDir, IndexFile), nil
}

var ErrVaultNotFound = fmt.Errorf("vault file does not exist")

// ReadVaultFile reads the encrypted vault file from disk.
func ReadVaultFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrVaultNotFound
		}
		return nil, fmt.Errorf("reading vault file: %w", err)
	}
	return data, nil
}

// ReadVaultIndex reads the unencrypted resource names index from disk.
func ReadVaultIndex() (*VaultIndex, error) {
	indexPath, err := DefaultIndexPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	var index VaultIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}

	return &index, nil
}

// WriteVaultIndex writes the unencrypted resource names index to disk.
func WriteVaultIndex(index *VaultIndex) error {
	indexPath, err := DefaultIndexPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(indexPath)
	if err := os.MkdirAll(dir, DirPermissions); err != nil {
		return err
	}

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(indexPath, append(data, '\n'), FilePermissions)
}

// WriteVaultFile writes encrypted data to the vault file atomically.
//
// Safety sequence:
//  1. Back up the existing vault to backup/vault.age
//  2. Write data to a temp file in the same directory
//  3. Rename temp → vault.age (atomic on the same filesystem)
//
// A crash at any point leaves the original vault intact.
func WriteVaultFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, DirPermissions); err != nil {
		return fmt.Errorf("creating vault directory: %w", err)
	}

	// Step 1: Back up the existing vault if it exists.
	if _, err := os.Stat(path); err == nil {
		if err := backupVault(path); err != nil {
			return fmt.Errorf("backing up vault before write: %w", err)
		}
	}

	// Step 2: Write to a temp file in the same directory (same filesystem = atomic rename).
	tmp, err := os.CreateTemp(dir, ".vault-tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp vault file: %w", err)
	}
	tmpPath := tmp.Name()

	// Ensure temp file is cleaned up on any error.
	defer func() {
		if _, err := os.Stat(tmpPath); err == nil {
			os.Remove(tmpPath)
		}
	}()

	if err := tmp.Chmod(FilePermissions); err != nil {
		tmp.Close()
		return fmt.Errorf("setting temp file permissions: %w", err)
	}

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temp vault file: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("syncing temp vault file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp vault file: %w", err)
	}

	// Step 3: Atomic rename — this either succeeds or fails completely.
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replacing vault file: %w", err)
	}

	return nil
}

// backupVault copies the current vault.age to backup/vault.age.
func backupVault(vaultPath string) error {
	backupDir := filepath.Join(filepath.Dir(vaultPath), BackupDir)
	if err := os.MkdirAll(backupDir, DirPermissions); err != nil {
		return fmt.Errorf("creating backup directory: %w", err)
	}

	data, err := os.ReadFile(vaultPath)
	if err != nil {
		return fmt.Errorf("reading vault for backup: %w", err)
	}

	backupPath := filepath.Join(backupDir, VaultFile)
	if err := os.WriteFile(backupPath, data, FilePermissions); err != nil {
		return fmt.Errorf("writing backup: %w", err)
	}

	return nil
}
