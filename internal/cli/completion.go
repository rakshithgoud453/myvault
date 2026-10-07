package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rakshithgoud453/myvault/internal/storage"
)

var completionInstallFlag bool

func init() {
	completionCmd.Flags().BoolVarP(&completionInstallFlag, "install", "i", false, "Automatically append autocompletion setup to your shell configuration file")
	rootCmd.AddCommand(completionCmd)

	// Register dynamic resource name completion on commands that take resource names
	for _, cmd := range []*cobra.Command{getCmd, copyCmd, setCmd, deleteCmd, addCmd, createCmd, execCmd} {
		cmd.ValidArgsFunction = resourceNameCompletion
	}
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate or install shell autocompletion scripts",
	Long: `Generate or automatically install shell autocompletion for zsh, bash, fish, or powershell.

Example:
  myvault completion zsh
  myvault completion --install
`,
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Args:      cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shellType := ""
		if len(args) == 1 {
			shellType = args[0]
		}

		if completionInstallFlag {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			if shellType == "" {
				userShell := os.Getenv("SHELL")
				if strings.Contains(userShell, "zsh") {
					shellType = "zsh"
				} else if strings.Contains(userShell, "bash") {
					shellType = "bash"
				} else if strings.Contains(userShell, "fish") {
					shellType = "fish"
				} else {
					shellType = "zsh" // default on macOS
				}
			}

			switch shellType {
			case "zsh":
				zshrcPath := filepath.Join(home, ".zshrc")
				content, _ := os.ReadFile(zshrcPath)
				strContent := string(content)

				var linesToAdd []string
				if !strings.Contains(strContent, "compinit") {
					linesToAdd = append(linesToAdd, "autoload -U compinit && compinit")
				}
				if !strings.Contains(strContent, "myvault completion zsh") {
					linesToAdd = append(linesToAdd, "source <(myvault completion zsh)")
				}

				if len(linesToAdd) > 0 {
					f, err := os.OpenFile(zshrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						return fmt.Errorf("updating ~/.zshrc: %w", err)
					}
					defer f.Close()

					_, _ = f.WriteString("\n# myvault shell completion\n" + strings.Join(linesToAdd, "\n") + "\n")
					fmt.Printf("Autocompletion installed into %s!\n", zshrcPath)
				} else {
					fmt.Printf("Autocompletion already configured in %s.\n", zshrcPath)
				}
				fmt.Println("Run 'source ~/.zshrc' to activate in your current terminal session.")

			case "bash":
				bashrcPath := filepath.Join(home, ".bashrc")
				if _, err := os.Stat(bashrcPath); os.IsNotExist(err) {
					bashrcPath = filepath.Join(home, ".bash_profile")
				}
				content, _ := os.ReadFile(bashrcPath)
				if !strings.Contains(string(content), "myvault completion bash") {
					f, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						return fmt.Errorf("updating %s: %w", bashrcPath, err)
					}
					defer f.Close()
					_, _ = f.WriteString("\n# myvault shell completion\nsource <(myvault completion bash)\n")
					fmt.Printf("Autocompletion installed into %s!\n", bashrcPath)
				} else {
					fmt.Printf("Autocompletion already configured in %s.\n", bashrcPath)
				}
				fmt.Printf("Run 'source %s' to activate in your current terminal session.\n", bashrcPath)

			default:
				return fmt.Errorf("automatic --install for %q is not supported. Output script with 'myvault completion %s' and source manually", shellType, shellType)
			}

			return nil
		}

		if shellType == "" {
			return fmt.Errorf("specify a shell type (zsh, bash, fish, powershell) or run 'myvault completion --install'")
		}

		switch shellType {
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
