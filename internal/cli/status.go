package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/session"
	"github.com/rakshithgoud453/myvault/internal/storage"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show vault status (locked/unlocked, vault path)",
	RunE: func(cmd *cobra.Command, args []string) error {
		vaultPath, err := storage.DefaultVaultPath()
		if err != nil {
			return err
		}

		fmt.Printf("Vault:   %s\n", vaultPath)

		if session.IsActive() {
			remaining := session.RemainingTime()
			fmt.Printf("Status:  🔓 Unlocked (%s remaining)\n", formatDuration(remaining))
		} else {
			fmt.Println("Status:  🔒 Locked")
		}

		return nil
	},
}
