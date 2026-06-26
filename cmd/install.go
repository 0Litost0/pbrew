package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/pbrew/internal/config"
	"github.com/user/pbrew/internal/installer"
	"github.com/user/pbrew/internal/source"
)

var (
	installAgent string
	installScope string
	installForce bool
)

var installCmd = &cobra.Command{
	Use:   "install <name...>",
	Short: "Install AI components to a target agent",
	Long: `Install one or more AI components from the configured data source to a target agent.

The --agent flag is required. Use --scope to choose between user-level and project-level
installation (default: project). Use --force to skip confirmation prompts.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if installAgent == "" {
			return fmt.Errorf("--agent is required")
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if cfg.Source == "" {
			return fmt.Errorf("data source not configured. Run: pbrew config set source <url-or-path>")
		}

		src, err := source.OpenSource(cfg.Source)
		if err != nil {
			return fmt.Errorf("open source: %w", err)
		}
		defer src.Close()

		if installScope == "" {
			installScope = "project"
		}

		anyFailed := false
		for _, name := range args {
			if err := installer.Install(installer.InstallParams{
				ComponentName: name,
				Agent:         installAgent,
				Scope:         installScope,
				Source:        src,
				Force:         installForce,
			}); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error installing %s: %v\n", name, err)
				anyFailed = true
			} else {
				fmt.Printf("installed %s\n", name)
			}
		}

		if anyFailed {
			return fmt.Errorf("some components failed to install")
		}
		return nil
	},
}

func init() {
	installCmd.Flags().StringVar(&installAgent, "agent", "", "Target agent (required, e.g. claude, codebuddy)")
	installCmd.Flags().StringVar(&installScope, "scope", "project", "Installation scope: user or project")
	installCmd.Flags().BoolVarP(&installForce, "force", "y", false, "Skip confirmation prompts")
	rootCmd.AddCommand(installCmd)
}
