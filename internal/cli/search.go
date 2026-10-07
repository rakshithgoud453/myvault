package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:     "search <query>",
	Short:   "Search across resource names, tags, and field keys/values",
	Long: `Search across resource names, tags, and field keys/values in the vault.

Returns matching resource names and highlights where the match occurred (e.g. matched on name, tag, or field key).

Examples:
  # Search for resources matching 'mysql':
  myvault search mysql
  myvault find mysql

  # Search for a domain or tag:
  myvault search visionwaves
  myvault search work
`,
	Aliases: []string{"find"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		v, _, err := loadVault()
		if err != nil {
			return err
		}

		matches := v.Search(query)
		if len(matches) == 0 {
			fmt.Printf("No resources matching %q found.\n", query)
			return nil
		}

		fmt.Printf("Found %d matching resource(s):\n\n", len(matches))
		for _, m := range matches {
			entry, canonicalName, _ := v.Get(m.ResourceName)
			tagsStr := ""
			if len(entry.Tags) > 0 {
				tagsStr = fmt.Sprintf(" (%s)", strings.Join(entry.Tags, " "))
			}

			fmt.Printf("• %s%s  [matched: %s]\n", canonicalName, tagsStr, m.MatchedOn)
		}
		return nil
	},
}
