package cli

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/storage"
)

func init() {
	rootCmd.AddCommand(completionCmd)

	// Register dynamic resource name completion on commands that take resource names
	for _, cmd := range []*cobra.Command{getCmd, copyCmd, setCmd, deleteCmd, addCmd, createCmd} {
		cmd.ValidArgsFunction = resourceNameCompletion
	}
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell autocompletion script",
	Long: `Generate shell autocompletion script for zsh, bash, fish, or powershell.

To load completions in ZSH:

1. Ensure compinit is enabled in your ~/.zshrc:
   autoload -U compinit && compinit

2. Source completions in current session:
   source <(myvault completion zsh)

3. Or add to ~/.zshrc for permanent autocompletion:
   echo 'autoload -U compinit && compinit' >> ~/.zshrc
   echo 'source <(myvault completion zsh)' >> ~/.zshrc
`,
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Args:      cobra.ExactValidArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletionV2(os.Stdout, true)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

// resourceNameCompletion dynamically autocompletes resource names using the local index.
func resourceNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	idx, err := storage.ReadVaultIndex()
	if err != nil || idx == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	toComp := strings.ToLower(toComplete)
	var matches []string

	for _, res := range idx.Resources {
		lowerName := strings.ToLower(res.Name)
		if toComp == "" || strings.HasPrefix(lowerName, toComp) || strings.Contains(lowerName, toComp) {
			matches = append(matches, res.Name)
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}
