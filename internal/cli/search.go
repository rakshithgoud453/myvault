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
