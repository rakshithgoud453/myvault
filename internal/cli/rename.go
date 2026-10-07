package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/crypto"
	"github.com/rakshithgoud453/myvault/internal/storage"
)

func init() {
	renameCmd.ValidArgsFunction = resourceNameCompletion
	rootCmd.AddCommand(renameCmd)
}

var renameCmd = &cobra.Command{
	Use:   "rename <old-name> <new-name>",
	Short: "Rename a secret resource",
	Long: `Rename an existing secret resource while preserving all its key-value fields and tags.

Fails if the old resource name does not exist or if the new resource name is already taken. Automatically updates the unencrypted metadata search index.

Examples:
  # Rename GITHUB_TAP to GITHUB_CI_TOKEN:
  myvault rename GITHUB_TAP GITHUB_CI_TOKEN

  # Rename MYSQL_OLD to MYSQL_PROD:
  myvault rename MYSQL_OLD MYSQL_PROD
`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldName := args[0]
		newName := args[1]

		v, passphrase, err := loadVault()
		if err != nil {
			return err
		}

		if err := v.Rename(oldName, newName); err != nil {
			return err
		}

		// Save updated vault
		plaintext, err := v.Serialize()
		if err != nil {
			return err
		}

		ciphertext, err := crypto.Encrypt(plaintext, passphrase)
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
		_ = storage.WriteVaultIndex(v.BuildIndex())

		fmt.Printf("Resource %q successfully renamed to %q.\n", strings.ToUpper(oldName), strings.ToUpper(newName))
		return nil
	},
}
