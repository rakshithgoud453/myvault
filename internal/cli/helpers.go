package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/rakshithgoud453/myvault/internal/crypto"
	"github.com/rakshithgoud453/myvault/internal/memguard"
	"github.com/rakshithgoud453/myvault/internal/session"
	"github.com/rakshithgoud453/myvault/internal/storage"
	"github.com/rakshithgoud453/myvault/internal/vault"
)

// promptPassphrase asks the user for the vault passphrase with hidden input.
func promptPassphrase(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	passBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr) // newline after hidden input
	if err != nil {
		return "", fmt.Errorf("reading passphrase: %w", err)
	}
	return string(passBytes), nil
}

// loadVault reads, decrypts, and parses the vault.
// If no vault file exists on disk, it prompts the user to initialize a new vault.
// Returns the parsed vault and the passphrase (so callers can re-encrypt).
func loadVault() (*vault.Vault, string, error) {
	vaultPath, err := storage.DefaultVaultPath()
	if err != nil {
		return nil, "", err
	}

	ciphertext, err := storage.ReadVaultFile(vaultPath)
	if err != nil {
		if errors.Is(err, storage.ErrVaultNotFound) {
			fmt.Println("No vault found. Setting up a new vault...")
			pass1, err := promptPassphrase("Set master passphrase: ")
			if err != nil {
				return nil, "", err
			}
			if pass1 == "" {
				return nil, "", fmt.Errorf("passphrase cannot be empty")
			}
			pass2, err := promptPassphrase("Confirm master passphrase: ")
			if err != nil {
				return nil, "", err
			}
			if pass1 != pass2 {
				return nil, "", fmt.Errorf("passphrases do not match")
			}

			v := vault.New()
			serialized, err := v.Serialize()
			if err != nil {
				return nil, "", err
			}
			enc, err := crypto.Encrypt(serialized, pass1)
			if err != nil {
				return nil, "", err
			}
			if err := storage.WriteVaultFile(vaultPath, enc); err != nil {
				return nil, "", fmt.Errorf("writing vault file: %w", err)
			}
			_ = storage.WriteVaultIndex(v.BuildIndex())
			fmt.Println("New vault initialized successfully.")
			return v, pass1, nil
		}
		return nil, "", err
	}

	// Try to use cached passphrase from an active session.
	var passphrase string
	if pass, err := session.Load(); err == nil {
		passphrase = pass
		// Refresh the session timer (sliding window).
		_ = session.Refresh(session.DefaultTimeout)
	} else {
		// No active session — prompt interactively.
		passphrase, err = promptPassphrase("Enter vault passphrase: ")
		if err != nil {
			return nil, "", err
		}
	}

	plaintext, err := crypto.Decrypt(ciphertext, passphrase)
	if err != nil {
		return nil, "", fmt.Errorf("wrong passphrase or corrupted vault")
	}
	// Zero decrypted plaintext as soon as parsing is done.
	defer memguard.ZeroBytes(plaintext)

	v, err := vault.Parse(plaintext)
	if err != nil {
		return nil, "", err
	}

	return v, passphrase, nil
}

// KVPair holds a key and value pair extracted from arguments.
type KVPair struct {
	Key   string
	Value string
}

// parseKeyValueArgs parses CLI arguments into key-value pairs.
// Seamlessly supports both 'key=value' and 'key value' (space-separated) syntaxes.
// Examples:
//
//	["username=alice", "port=3306"]
//	["username", "alice", "port", "3306"]
//	["username=alice", "port", "3306"]  (mixed)
func parseKeyValueArgs(args []string) ([]KVPair, error) {
	var pairs []KVPair
	i := 0
	for i < len(args) {
		arg := args[i]
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			if k == "" {
				return nil, fmt.Errorf("field key cannot be empty in %q", arg)
			}
			pairs = append(pairs, KVPair{Key: k, Value: v})
			i++
		} else {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for key %q (use 'key value' or 'key=value')", arg)
			}
			k := strings.TrimSpace(arg)
			v := strings.TrimSpace(args[i+1])
			if k == "" {
				return nil, fmt.Errorf("field key cannot be empty")
			}
			pairs = append(pairs, KVPair{Key: k, Value: v})
			i += 2
		}
	}
	return pairs, nil
}
