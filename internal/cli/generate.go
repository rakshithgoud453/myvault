package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/clipboard"
	"github.com/rakshithgoud453/myvault/internal/password"
)

var passwordLength int

func init() {
	generateCmd.Flags().IntVarP(&passwordLength, "length", "l", password.DefaultLength, "Password length")
	rootCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a random password",
	RunE: func(cmd *cobra.Command, args []string) error {
		pw, err := password.Generate(passwordLength)
		if err != nil {
			return fmt.Errorf("generating password: %w", err)
		}

		fmt.Println(pw)

		// Also copy to clipboard.
		if err := clipboard.CopyAndClear(pw, clipboard.DefaultClearTimeout); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not copy to clipboard: %v\n", err)
			return nil
		}
		fmt.Println("Password copied to clipboard. Clipboard will be cleared in 30 seconds.")
		return nil
	},
}
