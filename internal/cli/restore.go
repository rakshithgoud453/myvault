package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/crypto"
	"github.com/rakshithgoud453/myvault/internal/memguard"
	"github.com/rakshithgoud453/myvault/internal/storage"
	"github.com/rakshithgoud453/myvault/internal/vault"
)

func init() {
	rootCmd.AddCommand(restoreCmd)
}

var restoreCmd = &cobra.Command{
	Use:   "restore <snapshot-file>",
	Short: "Restore vault from an encrypted snapshot backup file",
	Long: `Restore your vault from an age-encrypted snapshot backup file (e.g. vault-*.bak).

Prompts for the passphrase used when the snapshot was created, verifies HMAC integrity, asks for confirmation, and overwrites ~/.password-vault/vault.age and ~/.password-vault/vault.index.

Examples:
  # Restore vault from a backup snapshot:
  myvault restore ~/Desktop/vault-20261008-013905.bak
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshotPath := args[0]

		snapshotData, err := os.ReadFile(snapshotPath)
		if err != nil {
			return fmt.Errorf("reading snapshot file: %w", err)
		}

		// Prompt for passphrase to validate snapshot
		passphrase, err := promptPassphrase("Enter passphrase to decrypt snapshot: ")
		if err != nil {
			return err
		}

		plaintext, err := crypto.Decrypt(snapshotData, passphrase)
		if err != nil {
			return fmt.Errorf("wrong passphrase or invalid snapshot file")
		}
		defer memguard.ZeroBytes(plaintext)

		// Parse snapshot vault to build index
		restoredVault, err := vault.Parse(plaintext)
		if err != nil {
			return fmt.Errorf("parsing restored vault: %w", err)
		}

		// Confirm restoration
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Restoring will overwrite current vault with %d resources from snapshot. Continue? [y/N]: ", len(restoredVault.Entries))
		confirm, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading confirmation: %w", err)
		}
		if c := strings.TrimSpace(strings.ToLower(confirm)); c != "y" && c != "yes" {
			fmt.Println("Restore cancelled.")
			return nil
		}

		vaultPath, err := storage.DefaultVaultPath()
		if err != nil {
			return err
		}

		// Write snapshot atomically (this backs up current vault first)
		if err := storage.WriteVaultFile(vaultPath, snapshotData); err != nil {
			return fmt.Errorf("writing restored vault: %w", err)
		}

		// Write restored index
		_ = storage.WriteVaultIndex(restoredVault.BuildIndex())

		fmt.Println("Vault restored successfully from snapshot!")
		return nil
	},
}
