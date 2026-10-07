package cli

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/crypto"
	"github.com/rakshithgoud453/myvault/internal/storage"
)

var (
	exportFormat   string
	exportEncrypt  bool
	exportCustomPw bool
)

func init() {
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "json", "Export format: json or csv")
	exportCmd.Flags().BoolVarP(&exportEncrypt, "encrypt", "e", false, "Encrypt the exported output file using age")
	exportCmd.Flags().BoolVarP(&exportCustomPw, "passphrase", "p", false, "Prompt for a custom export passphrase (used with --encrypt)")
	rootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export [output-file]",
	Short: "Export vault entries to JSON or CSV format (unencrypted or age-encrypted)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, vaultPassphrase, err := loadVault()
		if err != nil {
			return err
		}

		var exportData []byte

		switch strings.ToLower(exportFormat) {
		case "json":
			exportData, err = v.Serialize()
			if err != nil {
				return err
			}
		case "csv":
			var buf bytes.Buffer
			writer := csv.NewWriter(&buf)

			// Header: resource, field, value, concealed, tags
			_ = writer.Write([]string{"resource", "field", "value", "concealed", "tags"})

			resources := v.List()
			for _, resName := range resources {
				entry, canonRes, _ := v.Get(resName)
				tagsStr := strings.Join(entry.Tags, " ")

				fieldKeys := make([]string, 0, len(entry.Fields))
				for fk := range entry.Fields {
					fieldKeys = append(fieldKeys, fk)
				}
				sort.Strings(fieldKeys)

				for _, fk := range fieldKeys {
					fVal := entry.Fields[fk]
					concealedStr := "false"
					if fVal.Concealed {
						concealedStr = "true"
					}
					_ = writer.Write([]string{canonRes, fk, fVal.Value, concealedStr, tagsStr})
				}
			}
			writer.Flush()
			exportData = buf.Bytes()
		default:
			return fmt.Errorf("unsupported export format %q (use 'json' or 'csv')", exportFormat)
		}

		if exportEncrypt {
			encPass := vaultPassphrase
			if exportCustomPw {
				p1, err := promptPassphrase("Enter custom export passphrase: ")
				if err != nil {
					return err
				}
				if p1 == "" {
					return fmt.Errorf("export passphrase cannot be empty")
				}
				p2, err := promptPassphrase("Confirm custom export passphrase: ")
				if err != nil {
					return err
				}
				if p1 != p2 {
					return fmt.Errorf("export passphrases do not match")
				}
				encPass = p1
			}

			ciphertext, err := crypto.Encrypt(exportData, encPass)
			if err != nil {
				return fmt.Errorf("encrypting export data: %w", err)
			}
			exportData = ciphertext
		}

		// Write to stdout or file
		if len(args) == 0 {
			fmt.Print(string(exportData))
			return nil
		}

		outPath := args[0]
		if err := os.WriteFile(outPath, exportData, storage.FilePermissions); err != nil {
			return fmt.Errorf("writing export file: %w", err)
		}

		encLabel := ""
		if exportEncrypt {
			encLabel = " (encrypted)"
		}
		fmt.Printf("Vault exported successfully%s to %s (format: %s).\n", encLabel, outPath, exportFormat)
		return nil
	},
}
