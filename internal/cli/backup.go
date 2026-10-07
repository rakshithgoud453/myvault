package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/storage"
)

func init() {
	rootCmd.AddCommand(backupCmd)
}

var backupCmd = &cobra.Command{
	Use:   "backup [target-dir]",
	Short: "Create a timestamped encrypted snapshot backup of the vault",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vaultPath, err := storage.DefaultVaultPath()
		if err != nil {
			return err
		}

		data, err := storage.ReadVaultFile(vaultPath)
		if err != nil {
			return err
		}

		targetDir := filepath.Join(filepath.Dir(vaultPath), storage.BackupDir)
		if len(args) == 1 {
			targetDir = args[0]
		}

		if err := os.MkdirAll(targetDir, storage.DirPermissions); err != nil {
			return fmt.Errorf("creating backup target directory: %w", err)
		}

		filename := fmt.Sprintf("vault-%s.bak", time.Now().Format("20060102-150405"))
		backupPath := filepath.Join(targetDir, filename)

		if err := os.WriteFile(backupPath, data, storage.FilePermissions); err != nil {
			return fmt.Errorf("writing backup snapshot: %w", err)
		}

		fmt.Printf("Encrypted vault snapshot created: %s\n", backupPath)
		return nil
	},
}
