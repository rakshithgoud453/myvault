package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

var showUnmasked bool

func init() {
	getCmd.Flags().BoolVarP(&showUnmasked, "show", "s", false, "Show concealed/password fields in plaintext")
	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get <resource> [field]",
	Short: "Get all fields or a specific field for a secret resource",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, _, err := loadVault()
		if err != nil {
			return err
		}

		resourceName := args[0]

		// Option A: Single field request (myvault get mysql host / myvault get mysql password)
		if len(args) == 2 {
			fieldName := args[1]
			field, _, _, err := v.GetField(resourceName, fieldName)
			if err != nil {
				return err
			}
			fmt.Println(field.Value)
			return nil
		}

		// Option B: Full resource output (myvault get mysql)
		entry, canonicalName, err := v.Get(resourceName)
		if err != nil {
			return err
		}

		fmt.Printf("[%s]\n", canonicalName)

		if len(entry.Fields) == 0 {
			fmt.Println("  (no fields)")
			return nil
		}

		// Sort field keys for predictable display
		fieldKeys := make([]string, 0, len(entry.Fields))
		maxLen := 0
		for fKey := range entry.Fields {
			fieldKeys = append(fieldKeys, fKey)
			if len(fKey) > maxLen {
				maxLen = len(fKey)
			}
		}
		sort.Strings(fieldKeys)

		for _, fKey := range fieldKeys {
			field := entry.Fields[fKey]
			val := field.Value
			if field.Concealed && !showUnmasked {
				val = "[REDACTED]"
			}
			fmt.Printf("%-*s : %s\n", maxLen, fKey, val)
		}

		return nil
	},
}
