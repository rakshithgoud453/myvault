package cli

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/storage"
)

var exportFormat string

func init() {
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "json", "Export format: json or csv")
	rootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export [output-file]",
	Short: "Export vault entries to JSON or CSV format",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, _, err := loadVault()
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

		// Write to stdout or file
		if len(args) == 0 {
			fmt.Print(string(exportData))
			return nil
		}

		outPath := args[0]
		if err := os.WriteFile(outPath, exportData, storage.FilePermissions); err != nil {
			return fmt.Errorf("writing export file: %w", err)
		}

		fmt.Printf("Vault exported successfully to %s (format: %s).\n", outPath, exportFormat)
		return nil
	},
}
