// Package session manages the unlock/lock lifecycle for the vault.
//
// Security model:
//   - The vault passphrase is stored in the OS-native secure credential manager
//     (macOS Keychain, Linux Secret Service / DBus, Windows Credential Manager)
//     via github.com/zalando/go-keyring.
//   - The session file (~/.password-vault/.session) stores only the expiry
//     timestamp — no secrets are written to disk in plaintext.
//   - The session auto-expires after DefaultTimeout of inactivity.
//   - Explicit lock removes the keyring item and deletes the session file.
package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/zalando/go-keyring"
)

const (
	// SessionFile stores only session metadata (expiry) — no secrets.
	SessionFile = ".session"

	// DefaultTimeout is how long a session stays valid after last use.
	DefaultTimeout = 1 * time.Minute

	// FilePermissions for the session file (rw-------).
	FilePermissions = 0600

	// keychainService is the service name registered with the OS keyring.
	keychainService = "myvault"

	// keychainAccount is the account identifier for the session passphrase.
	keychainAccount = "session-passphrase"
)

// sessionMeta holds only timing information — no secrets.
type sessionMeta struct {
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// sessionPath returns the full path to the session metadata file.
func sessionPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home directory: %w", err)
	}
	return filepath.Join(home, ".password-vault", SessionFile), nil
}

// Create stores a new session: passphrase → OS Keyring, expiry → file.
func Create(passphrase string, timeout time.Duration) error {
	// Store passphrase in OS Keyring.
	if err := keyring.Set(keychainService, keychainAccount, passphrase); err != nil {
		return fmt.Errorf("storing passphrase in keyring: %w", err)
	}

	// Write only the expiry metadata to disk.
	path, err := sessionPath()
	if err != nil {
		return err
	}

	now := time.Now()
	meta := sessionMeta{
		CreatedAt: now,
		ExpiresAt: now.Add(timeout),
	}

	data, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("encoding session metadata: %w", err)
	}

	if err := os.WriteFile(path, data, FilePermissions); err != nil {
		return fmt.Errorf("writing session file: %w", err)
	}

	return nil
}

// Load returns the cached passphrase if a valid session exists.
func Load() (passphrase string, err error) {
	path, errP := sessionPath()
	if errP != nil {
		return "", errP
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("vault is locked — run 'myvault unlock' first")
		}
		return "", fmt.Errorf("reading session file: %w", err)
	}

	var meta sessionMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		_ = Delete()
		return "", fmt.Errorf("vault is locked — run 'myvault unlock' first")
	}

	// Check expiry.
	if time.Now().After(meta.ExpiresAt) {
		_ = Delete()
		return "", fmt.Errorf("session expired — run 'myvault unlock' again")
	}

	// Retrieve passphrase from OS Keyring.
	pw, err := keyring.Get(keychainService, keychainAccount)
	if err != nil {
		_ = Delete()
		return "", fmt.Errorf("vault is locked — run 'myvault unlock' first")
	}

	return pw, nil
}

// IsActive returns true if a valid (non-expired) session exists.
func IsActive() bool {
	_, err := Load()
	return err == nil
}

// Refresh extends the session expiry from now.
func Refresh(timeout time.Duration) error {
	path, err := sessionPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading session file: %w", err)
	}

	var meta sessionMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("decoding session: %w", err)
	}

	meta.ExpiresAt = time.Now().Add(timeout)

	updated, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("encoding session metadata: %w", err)
	}

	return os.WriteFile(path, updated, FilePermissions)
}

// Delete removes the session: OS Keyring item + session file.
func Delete() error {
	_ = keyring.Delete(keychainService, keychainAccount)

	path, err := sessionPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("deleting session file: %w", err)
	}

	return nil
}

// RemainingTime returns how long until the session expires.
func RemainingTime() time.Duration {
	path, err := sessionPath()
	if err != nil {
		return 0
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	var meta sessionMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return 0
	}

	remaining := time.Until(meta.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}
