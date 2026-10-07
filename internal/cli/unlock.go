package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/crypto"
	"github.com/rakshithgoud453/myvault/internal/memguard"
	"github.com/rakshithgoud453/myvault/internal/session"
	"github.com/rakshithgoud453/myvault/internal/storage"
)

func init() {
	rootCmd.AddCommand(unlockCmd)
}

var unlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Unlock the vault (cache passphrase for subsequent commands)",
	Long: `Unlock the vault by entering your passphrase once.

After unlocking, subsequent commands (list, get, copy, add, delete) will
not ask for the passphrase again until the session expires (15 minutes)
or you run 'myvault lock'.

Each command you run refreshes the session timer.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if already unlocked.
		if session.IsActive() {
			remaining := session.RemainingTime()
			fmt.Printf("Vault is already unlocked (%s remaining).\n", formatDuration(remaining))
			return nil
		}

		// Read the vault file to validate the passphrase.
		vaultPath, err := storage.DefaultVaultPath()
		if err != nil {
			return err
		}

		ciphertext, err := storage.ReadVaultFile(vaultPath)
		if err != nil {
			return err
		}

		// Prompt for passphrase.
		passphrase, err := promptPassphrase("Enter vault passphrase: ")
		if err != nil {
			return err
		}

		// Validate by attempting decryption.
		plaintext, err := crypto.Decrypt(ciphertext, passphrase)
		if err != nil {
			return fmt.Errorf("wrong passphrase")
		}
		// Zero decrypted bytes immediately — we only needed them to validate.
		memguard.ZeroBytes(plaintext)

		// Create session.
		if err := session.Create(passphrase, session.DefaultTimeout); err != nil {
			return fmt.Errorf("creating session: %w", err)
		}

		fmt.Println("Vault unlocked.")
		fmt.Printf("Session expires in %s. Run 'myvault lock' to lock manually.\n",
			formatDuration(session.DefaultTimeout))
		return nil
	},
}

// formatDuration formats a duration into a human-readable string.
func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	if minutes < 1 {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if minutes == 1 {
		return "1 minute"
	}
	return fmt.Sprintf("%d minutes", minutes)
}
