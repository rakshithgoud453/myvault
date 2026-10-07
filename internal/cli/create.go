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

var createGenerateFlag bool
var createTags []string

func init() {
	createCmd.Flags().BoolVar(&createGenerateFlag, "generate", false, "Generate a random password for a secret key")
	createCmd.Flags().StringSliceVarP(&createTags, "tag", "t", nil, "Tags to attach to the resource (e.g. --tag work --tag db)")
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create <resource> [key=value | key value ...]",
	Short: "Create a new secret resource (fails if resource already exists)",
	Long: `Create a new secret resource in the vault with flexible key-value fields and tags.

Supports both 'key=value' and space-separated 'key value' argument syntax, as well as an interactive prompt loop if no fields are supplied.

Examples:
  # Create using key=value syntax:
  myvault create MYSQL host=localhost port=3306 username=root password=secret

  # Create using space-separated syntax:
  myvault create JUMP_SERVER hostname jump.company.com username admin

  # Create with tags:
  myvault create GITHUB username=octocat password=secret --tag work --tag dev

  # Interactive creation prompt:
  myvault create AWS_PROD

  # Generate a random 32-character password automatically:
  myvault create MY_DATABASE --generate
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		v, passphrase, err := loadVault()
		if err != nil {
			return err
		}

		if _, canonName, errRes := v.Get(name); errRes == nil {
			return fmt.Errorf("resource %q already exists. Use 'myvault set %s ...' to update fields", canonName, strings.ToLower(canonName))
		}

		fieldsModified := 0
		reader := bufio.NewReader(os.Stdin)

		if len(args) > 1 {
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
		} else if createGenerateFlag {
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
			fmt.Printf("Creating secret resource %q (press Enter on empty Key to finish):\n\n", strings.ToUpper(name))

			for {
				fmt.Print("Key: ")
				k, err := reader.ReadString('\n')
				if err != nil {
					break
				}
				k = strings.TrimSpace(k)
				if k == "" {
					break
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

		if len(createTags) > 0 {
			_ = v.SetTags(name, createTags)
		}

		if fieldsModified == 0 && len(createTags) == 0 {
			fmt.Println("No fields created.")
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
		fmt.Printf("Resource %q created (%d field(s) set).\n", canonName, fieldsModified)
		return nil
	},
}
