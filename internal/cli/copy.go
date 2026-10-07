package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/clipboard"
)

func init() {
	rootCmd.AddCommand(copyCmd)
}

var copyCmd = &cobra.Command{
	Use:     "copy <resource> [field]",
	Short:   "Copy a field (or default password/secret) to clipboard with 30s auto-clear",
	Aliases: []string{"cp"},
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, _, err := loadVault()
		if err != nil {
			return err
		}

		resourceName := args[0]
		var valToCopy string
		var fieldName string
		var canonicalResource string

		if len(args) == 2 {
			// Copy specified field
			requestedField := args[1]
			field, canonRes, canonField, err := v.GetField(resourceName, requestedField)
			if err != nil {
				return err
			}
			valToCopy = field.Value
			fieldName = canonField
			canonicalResource = canonRes
		} else {
			// Copy default field (password, secret, token, or first concealed key)
			entry, canonRes, err := v.Get(resourceName)
			if err != nil {
				return err
			}
			field, defFieldName, found := entry.DefaultField()
			if !found {
				return fmt.Errorf("resource %q has no fields to copy", canonRes)
			}
			valToCopy = field.Value
			fieldName = defFieldName
			canonicalResource = canonRes
		}

		if err := clipboard.CopyAndClear(valToCopy, clipboard.DefaultClearTimeout); err != nil {
			return err
		}

		fmt.Printf("Field '%s' for %s copied to clipboard.\n", fieldName, canonicalResource)
		fmt.Println("Clipboard will be cleared in 30 seconds.")
		return nil
	},
}
