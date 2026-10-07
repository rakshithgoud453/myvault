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

To load completions:

ZSH:
  # To load completions in your current shell session:
  source <(myvault completion zsh)

  # To load completions permanently:
  myvault completion zsh > "${fpath[1]}/_myvault"

BASH:
  # Linux:
  myvault completion bash > /etc/bash_completion.d/myvault

  # macOS (with bash-completion):
  myvault completion bash > $(brew --prefix)/etc/bash_completion.d/myvault

FISH:
  myvault completion fish > ~/.config/fish/completions/myvault.fish
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
		if strings.HasPrefix(strings.ToLower(res.Name), toComp) {
			matches = append(matches, res.Name)
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}
