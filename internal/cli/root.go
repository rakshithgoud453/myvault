// Package cli defines the root command and all subcommands for the MyVault CLI.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags, with a default fallback.
var Version = "0.2.0-dev"

// rootCmd is the top-level command for MyVault.
var rootCmd = &cobra.Command{
	Use:   "myvault",
	Short: "MyVault — a local-first password manager",
	Long: `MyVault is a local-first password manager.

Your encrypted vault belongs to you. No cloud account required.
No third-party password storage. No vendor lock-in.

Passwords are encrypted using the age encryption format and stored
locally on your machine.`,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// versionCmd prints the current MyVault version.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the MyVault version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("myvault %s\n", Version)
	},
}

// Execute runs the root command. Called from main.
func Execute() error {
	return rootCmd.Execute()
}
