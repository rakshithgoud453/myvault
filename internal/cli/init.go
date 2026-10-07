package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/crypto"
	"github.com/rakshithgoud453/myvault/internal/storage"
	"github.com/rakshithgoud453/myvault/internal/vault"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new encrypted vault and set master passphrase",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		vaultPath, err := storage.DefaultVaultPath()
		if err != nil {
			return err
		}

		// Check if vault already exists
		if _, err := storage.ReadVaultFile(vaultPath); err == nil {
			return fmt.Errorf("vault already exists at %s", vaultPath)
		}

		fmt.Println("Initializing new vault...")
		pass1, err := promptPassphrase("Set master passphrase: ")
		if err != nil {
			return err
		}
		if pass1 == "" {
			return fmt.Errorf("passphrase cannot be empty")
		}

		pass2, err := promptPassphrase("Confirm master passphrase: ")
		if err != nil {
			return err
		}

		if pass1 != pass2 {
			return fmt.Errorf("passphrases do not match")
		}

		v := vault.New()
		plaintext, err := v.Serialize()
		if err != nil {
			return err
		}

		ciphertext, err := crypto.Encrypt(plaintext, pass1)
		if err != nil {
			return err
		}

		if err := storage.WriteVaultFile(vaultPath, ciphertext); err != nil {
			return fmt.Errorf("writing vault file: %w", err)
		}

		_ = storage.WriteVaultIndex(v.BuildIndex())

		fmt.Println("Vault initialized successfully!")
		return nil
	},
}
