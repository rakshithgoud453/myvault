package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/session"
)

func init() {
	rootCmd.AddCommand(lockCmd)
}

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Lock the vault (clear cached passphrase)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !session.IsActive() {
			fmt.Println("Vault is already locked.")
			return nil
		}

		if err := session.Delete(); err != nil {
			return fmt.Errorf("locking vault: %w", err)
		}

		fmt.Println("Vault locked. Session cleared.")
		return nil
	},
}
