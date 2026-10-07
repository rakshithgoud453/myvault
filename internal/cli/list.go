package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/policy"
	"github.com/rakshithgoud453/myvault/internal/storage"
)

var listTagFilter string

func init() {
	listCmd.Flags().StringVarP(&listTagFilter, "tag", "t", "", "Filter entries by tag (e.g. --tag work)")
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all secret resources in the vault (optionally filtered by tag)",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check policy requirement for list operation (policy.OpList does NOT require passphrase)
		if !policy.RequiresPassphrase(policy.OpList) {
			idx, err := storage.ReadVaultIndex()
			if err == nil && idx != nil {
				displayIndex(idx, listTagFilter)
				return nil
			}
		}

		// Fallback for legacy vaults without index: prompt passphrase once to build index
		v, _, err := loadVault()
		if err != nil {
			return err
		}

		// Save index so future list operations don't require passphrase
		_ = storage.WriteVaultIndex(v.BuildIndex())

		var names []string
		if listTagFilter != "" {
			names = v.ListByTag(listTagFilter)
		} else {
			names = v.List()
		}

		if len(names) == 0 {
			if listTagFilter != "" {
				fmt.Printf("No resources found matching tag %q.\n", listTagFilter)
			} else {
				fmt.Println("Vault is empty.")
			}
			return nil
		}

		for _, name := range names {
			entry, _, _ := v.Get(name)
			if len(entry.Tags) > 0 {
				fmt.Printf("%s  (%s)\n", name, strings.Join(entry.Tags, " "))
			} else {
				fmt.Println(name)
			}
		}
		return nil
	},
}

func displayIndex(idx *storage.VaultIndex, tagFilter string) {
	if len(idx.Resources) == 0 {
		fmt.Println("Vault is empty.")
		return
	}

	targetTag := strings.ToLower(strings.TrimPrefix(tagFilter, "#"))
	var displayed []storage.ResourceIndexEntry

	for _, res := range idx.Resources {
		if tagFilter != "" {
			match := false
			for _, t := range res.Tags {
				if strings.ToLower(strings.TrimPrefix(t, "#")) == targetTag {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		displayed = append(displayed, res)
	}

	if len(displayed) == 0 {
		if tagFilter != "" {
			fmt.Printf("No resources found matching tag %q.\n", tagFilter)
		} else {
			fmt.Println("Vault is empty.")
		}
		return
	}

	sort.Slice(displayed, func(i, j int) bool {
		return displayed[i].Name < displayed[j].Name
	})

	for _, res := range displayed {
		if len(res.Tags) > 0 {
			fmt.Printf("%s  (%s)\n", res.Name, strings.Join(res.Tags, " "))
		} else {
			fmt.Println(res.Name)
		}
	}
}
