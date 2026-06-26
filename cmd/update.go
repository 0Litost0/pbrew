package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/pbrew/internal/config"
	"github.com/user/pbrew/internal/installer"
	"github.com/user/pbrew/internal/source"
)

var (
	updateAgent string
	updateScope string
	updateAll   bool
)

var updateCmd = &cobra.Command{
	Use:   "update [name...]",
	Short: "Update installed AI components",
	Long: `Update installed AI components to the latest version from the data source.

Use --all to update all installed components for the given agent and scope.
Otherwise, specify one or more component names.`,
	Args: cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		if updateAgent == "" {
			return fmt.Errorf("--agent is required")
		}
		if !updateAll && len(args) == 0 {
			return fmt.Errorf("specify component names or use --all")
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

		if updateScope == "" {
			updateScope = "project"
		}

		return installer.Update(installer.UpdateParams{
			ComponentNames: args,
			Agent:          updateAgent,
			Scope:          updateScope,
			All:            updateAll,
			Source:         src,
		})
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateAgent, "agent", "", "Target agent (required)")
	updateCmd.Flags().StringVar(&updateScope, "scope", "project", "Scope: user or project")
	updateCmd.Flags().BoolVar(&updateAll, "all", false, "Update all installed components")
	rootCmd.AddCommand(updateCmd)
}
