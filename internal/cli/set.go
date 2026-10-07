package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/crypto"
	"github.com/rakshithgoud453/myvault/internal/storage"
	"github.com/rakshithgoud453/myvault/internal/vault"
)

var forceConceal bool

func init() {
	setCmd.Flags().BoolVarP(&forceConceal, "conceal", "c", false, "Explicitly mark set fields as concealed (masked)")
	rootCmd.AddCommand(setCmd)
}

var setCmd = &cobra.Command{
	Use:   "set <resource> <key> <value> [key2 value2 | key2=value2 ...]",
	Short: "Set or update fields in a secret resource (supports both key=value and key value)",
	Long: `Set or update key-value fields inside an existing or new secret resource.

Seamlessly supports both 'key=value' and space-separated 'key value' argument syntax. Pass '-c' or '--conceal' to explicitly mark custom fields as concealed/masked.

Examples:
  # Update a single field (key=value):
  myvault set JUMP_SERVER hostname=jump.visionwaves.com

  # Update a single field (space-separated):
  myvault set JUMP_SERVER hostname jump.visionwaves.com

  # Update multiple fields at once:
  myvault set MYSQL host db.prod.internal port 3306 password newsecret123

  # Force a field to be marked as concealed (masked by default in get):
  myvault set API_KEY custom_token "secret_val" -c
`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		resourceName := args[0]

		pairs, err := parseKeyValueArgs(args[1:])
		if err != nil {
			return err
		}

		v, passphrase, err := loadVault()
		if err != nil {
			return err
		}

		fieldsModified := 0
		for _, pair := range pairs {
			concealed := forceConceal || vault.IsConcealedKey(pair.Key)
			if v.SetField(resourceName, pair.Key, pair.Value, concealed) {
				fieldsModified++
			}
		}

		_, canonRes, _ := v.Get(resourceName)

		if fieldsModified == 0 {
			fmt.Printf("Resource %q fields are unchanged (identical values).\n", canonRes)
			return nil
		}

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

		fmt.Printf("Resource %q updated (%d field(s) set).\n", canonRes, fieldsModified)
		return nil
	},
}
