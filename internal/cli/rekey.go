package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/crypto"
	"github.com/rakshithgoud453/myvault/internal/session"
	"github.com/rakshithgoud453/myvault/internal/storage"
)

func init() {
	rootCmd.AddCommand(rekeyCmd)
}

var rekeyCmd = &cobra.Command{
	Use:   "rekey",
	Short: "Change the master vault passphrase",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Load vault (verifies current passphrase or session)
		v, _, err := loadVault()
		if err != nil {
			return err
		}

		// 2. Prompt for new passphrase
		newPass, err := promptPassphrase("Enter new vault passphrase: ")
		if err != nil {
			return err
		}
		if newPass == "" {
			return fmt.Errorf("passphrase cannot be empty")
		}

		confirmPass, err := promptPassphrase("Confirm new vault passphrase: ")
		if err != nil {
			return err
		}

		if newPass != confirmPass {
			return fmt.Errorf("passphrases do not match — cancellation")
		}

		// 3. Serialize and re-encrypt vault with new passphrase
		plaintext, err := v.Serialize()
		if err != nil {
			return err
		}

		ciphertext, err := crypto.Encrypt(plaintext, newPass)
		if err != nil {
			return err
		}

		vaultPath, err := storage.DefaultVaultPath()
		if err != nil {
			return err
		}

		if err := storage.WriteVaultFile(vaultPath, ciphertext); err != nil {
			return err
		}

		// 4. Update session keyring if active, otherwise ensure session is cleared
		if session.IsActive() {
			if err := session.Create(newPass, session.DefaultTimeout); err != nil {
				fmt.Printf("Warning: failed to update session keyring: %v\n", err)
			}
		} else {
			_ = session.Delete()
		}

		fmt.Println("Vault passphrase updated successfully!")
		return nil
	},
}
