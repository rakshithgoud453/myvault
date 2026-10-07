package cli

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/crypto"
	"github.com/rakshithgoud453/myvault/internal/storage"
	"github.com/rakshithgoud453/myvault/internal/vault"
)

var importFormat string

func init() {
	importCmd.Flags().StringVarP(&importFormat, "format", "f", "auto", "Import format: auto, json, csv, bitwarden, 1password")
	rootCmd.AddCommand(importCmd)
}

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import entries from JSON, CSV, Bitwarden, or 1Password exports",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading import file: %w", err)
		}

		v, passphrase, err := loadVault()
		if err != nil {
			return err
		}

		importedCount := 0
		text := strings.TrimSpace(string(data))

		// Auto detect JSON vs CSV if format is auto
		fmtType := strings.ToLower(importFormat)
		if fmtType == "auto" {
			if len(text) > 0 && text[0] == '{' {
				fmtType = "json"
			} else {
				fmtType = "csv"
			}
		}

		switch fmtType {
		case "json":
			importedVault, err := vault.Parse(data)
			if err != nil {
				return fmt.Errorf("parsing JSON import: %w", err)
			}
			for resName, entry := range importedVault.Entries {
				for fKey, fVal := range entry.Fields {
					v.SetField(resName, fKey, fVal.Value, fVal.Concealed)
				}
				if len(entry.Tags) > 0 {
					_ = v.SetTags(resName, entry.Tags)
				}
				importedCount++
			}

		case "csv", "bitwarden", "1password":
			r := csv.NewReader(strings.NewReader(text))
			header, err := r.Read()
			if err != nil {
				return fmt.Errorf("reading CSV header: %w", err)
			}

			// Map header indices
			colMap := make(map[string]int)
			for idx, col := range header {
				colMap[strings.ToLower(strings.TrimSpace(col))] = idx
			}

			for {
				record, err := r.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					continue
				}

				// Check if native myvault CSV format (resource, field, value, concealed, tags)
				if _, ok := colMap["resource"]; ok {
					resName := getCol(record, colMap, "resource")
					fKey := getCol(record, colMap, "field")
					fVal := getCol(record, colMap, "value")
					concealedStr := getCol(record, colMap, "concealed")
					tagsStr := getCol(record, colMap, "tags")

					if resName != "" && fKey != "" {
						concealed := concealedStr == "true" || vault.IsConcealedKey(fKey)
						v.SetField(resName, fKey, fVal, concealed)
						if tagsStr != "" {
							tags := strings.Fields(tagsStr)
							_ = v.SetTags(resName, tags)
						}
						importedCount++
					}
					continue
				}

				// Bitwarden / 1Password format (name/title, username, password, login_uri/url, notes)
				nameIdx := findAnyCol(colMap, "name", "title")
				userIdx := findAnyCol(colMap, "username", "login_username")
				passIdx := findAnyCol(colMap, "password", "login_password")
				urlIdx := findAnyCol(colMap, "login_uri", "url")
				notesIdx := findAnyCol(colMap, "notes")

				if nameIdx != -1 {
					resName := record[nameIdx]
					if resName == "" {
						continue
					}
					if userIdx != -1 && record[userIdx] != "" {
						v.SetField(resName, "username", record[userIdx], false)
					}
					if passIdx != -1 && record[passIdx] != "" {
						v.SetField(resName, "password", record[passIdx], true)
					}
					if urlIdx != -1 && record[urlIdx] != "" {
						v.SetField(resName, "url", record[urlIdx], false)
					}
					if notesIdx != -1 && record[notesIdx] != "" {
						v.SetField(resName, "notes", record[notesIdx], false)
					}
					importedCount++
				}
			}
		default:
			return fmt.Errorf("unsupported import format %q", importFormat)
		}

		if importedCount == 0 {
			fmt.Println("No entries imported.")
			return nil
		}

		// Save vault
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

		fmt.Printf("Successfully imported %d secret resource(s) into vault!\n", importedCount)
		return nil
	},
}

func getCol(record []string, colMap map[string]int, key string) string {
	if idx, ok := colMap[key]; ok && idx < len(record) {
		return strings.TrimSpace(record[idx])
	}
	return ""
}

func findAnyCol(colMap map[string]int, keys ...string) int {
	for _, k := range keys {
		if idx, ok := colMap[k]; ok {
			return idx
		}
	}
	return -1
}
