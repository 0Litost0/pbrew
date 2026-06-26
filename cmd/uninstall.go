package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/pbrew/internal/installer"
)

var (
	uninstallAgent string
	uninstallScope string
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <name...>",
	Short: "Uninstall AI components from a target agent",
	Long: `Uninstall one or more AI components from a target agent.

Both --agent and --scope are required to avoid accidental deletion.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if uninstallAgent == "" {
			return fmt.Errorf("--agent is required")
		}
		if uninstallScope == "" {
			return fmt.Errorf("--scope is required for uninstall (to avoid accidental deletion)")
		}

		anyFailed := false
		for _, name := range args {
			if err := installer.Uninstall(installer.UninstallParams{
				ComponentName: name,
				Agent:         uninstallAgent,
				Scope:         uninstallScope,
			}); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error uninstalling %s: %v\n", name, err)
				anyFailed = true
			} else {
				fmt.Printf("uninstalled %s\n", name)
			}
		}

		if anyFailed {
			return fmt.Errorf("some components failed to uninstall")
		}
		return nil
	},
}

func init() {
	uninstallCmd.Flags().StringVar(&uninstallAgent, "agent", "", "Target agent (required)")
	uninstallCmd.Flags().StringVar(&uninstallScope, "scope", "", "Scope: user or project (required)")
	rootCmd.AddCommand(uninstallCmd)
}
