package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/pbrew/internal/agent"
	"github.com/user/pbrew/internal/config"
	"github.com/user/pbrew/internal/registry"
	"github.com/user/pbrew/internal/source"
)

var (
	listAgent string
	listScope string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available or installed AI components",
	Long: `List available components from the data source, or list installed components
for a specific agent using --agent flag.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if listAgent != "" {
			return listInstalled()
		}
		return listAvailable()
	},
}

func listAvailable() error {
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

	components, err := registry.ListComponents(src)
	if err != nil {
		return fmt.Errorf("list components: %w", err)
	}

	if len(components) == 0 {
		fmt.Println("No components found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tVERSION\tDESCRIPTION")
	for _, c := range components {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.Name, c.Type, c.Version, c.Description)
	}
	w.Flush()
	return nil
}

func listInstalled() error {
	// Validate agent
	if err := agent.Validate(listAgent); err != nil {
		return err
	}
	profile, _ := agent.Get(listAgent)

	if listScope == "" {
		listScope = "project"
	}

	// Determine project root
	var projectRoot string
	var err error
	if listScope == "project" {
		projectRoot, err = findProjectRoot()
		if err != nil {
			return err
		}
	}

	// Load records
	recordPath, err := recordFilePath(profile, listScope, projectRoot)
	if err != nil {
		return err
	}

	store, err := loadRecordStore(recordPath)
	if err != nil {
		return err
	}

	// Load source for version comparison
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var src source.Source
	if cfg.Source != "" {
		src, _ = source.OpenSource(cfg.Source)
		if src != nil {
			defer src.Close()
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tINSTALLED\tLATEST\tSTATUS")

	for _, r := range store.Records {
		if r.Agent != listAgent || r.Scope != listScope {
			continue
		}
		status := ""
		latest := "-"
		if src != nil {
			if entry, err := registry.FindComponent(src, r.Name); err == nil {
				latest = entry.Version
				if r.Version != entry.Version {
					status = "outdated"
				}
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.Name, r.Type, r.Version, latest, status)
	}
	w.Flush()
	return nil
}

func init() {
	listCmd.Flags().StringVar(&listAgent, "agent", "", "List installed components for this agent")
	listCmd.Flags().StringVar(&listScope, "scope", "", "Scope for --agent (user or project, default project)")
	rootCmd.AddCommand(listCmd)
}
