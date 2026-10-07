package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/clipboard"
	"github.com/rakshithgoud453/myvault/internal/crypto"
	"github.com/rakshithgoud453/myvault/internal/password"
	"github.com/rakshithgoud453/myvault/internal/storage"
	"github.com/rakshithgoud453/myvault/internal/vault"
)

var generateFlag bool
var addTags []string

func init() {
	addCmd.Flags().BoolVar(&generateFlag, "generate", false, "Generate a random password for a secret key")
	addCmd.Flags().StringSliceVarP(&addTags, "tag", "t", nil, "Tags to attach to the resource (e.g. --tag work --tag db)")
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add <resource> [key=value | key value ...]",
	Short: "Add a new secret resource or add/update fields (supports both key=value and key value)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		v, passphrase, err := loadVault()
		if err != nil {
			return err
		}

		_, _, errRes := v.Get(name)
		isNewResource := (errRes != nil)

		fieldsModified := 0
		reader := bufio.NewReader(os.Stdin)

		if len(args) > 1 {
			// Parse CLI arguments using flexible parser (supports both key=val and key val)
			pairs, err := parseKeyValueArgs(args[1:])
			if err != nil {
				return err
			}

			for _, pair := range pairs {
				concealed := vault.IsConcealedKey(pair.Key)
				if v.SetField(name, pair.Key, pair.Value, concealed) {
					fieldsModified++
				}
			}
		} else if generateFlag {
			// --generate flag passed
			fmt.Print("Key name [password]: ")
			keyName, _ := reader.ReadString('\n')
			keyName = strings.TrimSpace(keyName)
			if keyName == "" {
				keyName = "password"
			}

			pw, err := password.Generate(password.DefaultLength)
			if err != nil {
				return fmt.Errorf("generating password: %w", err)
			}
			if v.SetField(name, keyName, pw, true) {
				fieldsModified++
			}
			fmt.Printf("Generated password for field '%s'.\n", keyName)

			if err := clipboard.CopyAndClear(pw, clipboard.DefaultClearTimeout); err == nil {
				fmt.Println("Generated password copied to clipboard.")
				fmt.Println("Clipboard will be cleared in 30 seconds.")
			}
		} else {
			// Interactive key-value entry loop
			_, canonName, errRes := v.Get(name)
			displayName := name
			if errRes == nil {
				displayName = canonName
			}

			fmt.Printf("Adding secret fields to %q (press Enter on empty Key to finish):\n\n", strings.ToUpper(displayName))

			for {
				fmt.Print("Key: ")
				k, err := reader.ReadString('\n')
				if err != nil {
					break
				}
				k = strings.TrimSpace(k)
				if k == "" {
					break // Finish loop on empty key
				}

				concealed := vault.IsConcealedKey(k)
				var val string

				if concealed {
					val, err = promptPassphrase(fmt.Sprintf("Value for '%s' (hidden): ", k))
					if err != nil {
						return err
					}
				} else {
					fmt.Printf("Value for '%s': ", k)
					val, err = reader.ReadString('\n')
					if err != nil {
						return err
					}
					val = strings.TrimSpace(val)
				}

				if v.SetField(name, k, val, concealed) {
					fieldsModified++
				}
				fmt.Println()
			}
		}

		tagsChanged := false
		if len(addTags) > 0 {
			_ = v.SetTags(name, addTags)
			tagsChanged = true
		}

		if fieldsModified == 0 && !tagsChanged {
			_, canonName, _ := v.Get(name)
			fmt.Printf("No changes for resource %q (values are identical).\n", canonName)
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

		_, canonName, _ := v.Get(name)
		if isNewResource {
			fmt.Printf("Resource %q created (%d field(s) set).\n", canonName, fieldsModified)
		} else {
			fmt.Printf("Resource %q updated (%d field(s) set).\n", canonName, fieldsModified)
		}
		return nil
	},
}
