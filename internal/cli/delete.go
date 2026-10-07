package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/crypto"
	"github.com/rakshithgoud453/myvault/internal/storage"
)

func init() {
	rootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:     "delete <resource> [field]",
	Short:   "Delete an entire secret resource or a specific field within a resource",
	Aliases: []string{"rm"},
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		resourceName := args[0]

		v, passphrase, err := loadVault()
		if err != nil {
			return err
		}

		reader := bufio.NewReader(os.Stdin)

		if len(args) == 2 {
			// Delete single field
			fieldName := args[1]
			_, canonRes, canonField, err := v.GetField(resourceName, fieldName)
			if err != nil {
				return err
			}

			fmt.Printf("Delete field '%s' from resource %q? [y/N]: ", canonField, canonRes)
			confirm, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("reading confirmation: %w", err)
			}
			if c := strings.TrimSpace(strings.ToLower(confirm)); c != "y" && c != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}

			if err := v.DeleteField(resourceName, fieldName); err != nil {
				return err
			}

			fmt.Printf("Field '%s' removed from resource %q.\n", canonField, canonRes)
		} else {
			// Delete entire resource
			_, canonRes, err := v.Get(resourceName)
			if err != nil {
				return err
			}

			fmt.Printf("Delete resource %q and all its fields? [y/N]: ", canonRes)
			confirm, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("reading confirmation: %w", err)
			}
			if c := strings.TrimSpace(strings.ToLower(confirm)); c != "y" && c != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}

			if err := v.Delete(resourceName); err != nil {
				return err
			}

			fmt.Printf("Resource %q deleted.\n", canonRes)
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
		return nil
	},
}
