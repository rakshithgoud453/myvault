package cli

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(execCmd)
}

var execCmd = &cobra.Command{
	Use:   "exec <resource...> -- <command> [args...]",
	Short: "Execute a command with vault secrets injected into environment variables",
	Long: `Fetch secret fields for one or more resources and inject them as environment variables into a child command process.

Example:
  myvault exec MYSQL -- npm start
  myvault exec JUMP_SERVER AWS -- bash -c 'echo $HOSTNAME'
`,
	DisableFlagParsing: false,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find '--' separator in args if passed
		dashIdx := -1
		for i, arg := range args {
			if arg == "--" {
				dashIdx = i
				break
			}
		}

		var resourceNames []string
		var targetCmd string
		var targetArgs []string

		if dashIdx != -1 {
			resourceNames = args[:dashIdx]
			if dashIdx+1 < len(args) {
				targetCmd = args[dashIdx+1]
				targetArgs = args[dashIdx+2:]
			}
		} else {
			if len(args) < 2 {
				return fmt.Errorf("usage: myvault exec <resource...> -- <command> [args...]")
			}
			// Fallback: first N-1 args are resources, last arg is command
			resourceNames = args[:len(args)-1]
			targetCmd = args[len(args)-1]
		}

		if len(resourceNames) == 0 {
			return fmt.Errorf("at least one resource name must be specified before '--'")
		}
		if targetCmd == "" {
			return fmt.Errorf("no command specified to execute")
		}

		v, _, err := loadVault()
		if err != nil {
			return err
		}

		// Prepare environment variables array starting with current OS environment
		envMap := make(map[string]string)
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}

		// Inject fields from each resource
		injectedCount := 0
		for _, resName := range resourceNames {
			entry, canonName, err := v.Get(resName)
			if err != nil {
				return fmt.Errorf("resource %q not found: %w", resName, err)
			}

			// Clean prefix for environment variable naming if multiple resources are specified
			prefix := strings.ToUpper(strings.ReplaceAll(canonName, "-", "_"))

			for fk, fval := range entry.Fields {
				upperKey := strings.ToUpper(strings.ReplaceAll(fk, "-", "_"))

				// Direct variable name (e.g. USERNAME, PASSWORD)
				envMap[upperKey] = fval.Value

				// Prefixed variable name (e.g. MYSQL_USERNAME, MYSQL_PASSWORD)
				prefixedKey := fmt.Sprintf("%s_%s", prefix, upperKey)
				envMap[prefixedKey] = fval.Value

				injectedCount++
			}
		}

		// Build final environment slice
		finalEnv := make([]string, 0, len(envMap))
		for k, val := range envMap {
			finalEnv = append(finalEnv, fmt.Sprintf("%s=%s", k, val))
		}

		// Prepare child command
		c := exec.Command(targetCmd, targetArgs...)
		c.Env = finalEnv
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		// Signal forwarding
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		go func() {
			for sig := range sigChan {
				if c.Process != nil {
					_ = c.Process.Signal(sig)
				}
			}
		}()

		if err := c.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			return fmt.Errorf("executing command %q: %w", targetCmd, err)
		}

		return nil
	},
}
