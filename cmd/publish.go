package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/pbrew/internal/config"
	"github.com/user/pbrew/internal/publisher"
	"github.com/user/pbrew/internal/source"
)

var (
	publishName    string
	publishDir     string
	publishVersion string
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish an AI component to the data source",
	Long: `Publish a local AI component directory to the configured data source.

The component files are converted to the repository default format (Claude Code),
a .pbrew.json metadata file is generated, and the _index.json is updated.

If --version is not specified, the version is auto-incremented:
  - New component: 0.1.0
  - Existing component: patch +1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if publishName == "" {
			return fmt.Errorf("--name is required")
		}
		if publishDir == "" {
			return fmt.Errorf("--dir is required")
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

		result, err := publisher.Publish(publisher.PublishParams{
			Name:    publishName,
			Dir:     publishDir,
			Version: publishVersion,
			Source:  src,
		})
		if err != nil {
			return err
		}

		fmt.Printf("published %s %s\n", result.Name, result.Version)
		sourceRoot := src.Root()
		fmt.Printf("\nTo commit and push:\n")
		fmt.Printf("  cd %s && git add -A && git commit -m 'publish %s@%s' && git push\n", sourceRoot, result.Name, result.Version)

		return nil
	},
}

func init() {
	publishCmd.Flags().StringVar(&publishName, "name", "", "Component name (required)")
	publishCmd.Flags().StringVar(&publishDir, "dir", "", "Component source directory (required)")
	publishCmd.Flags().StringVar(&publishVersion, "version", "", "Explicit version (default: auto-increment)")
	rootCmd.AddCommand(publishCmd)
}
