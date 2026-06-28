package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/user/pbrew/internal/logger"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:          "pbrew",
	Short:        "pbrew — AI component manager for agents",
	Long:         "pbrew manages AI components (rules, skills, hooks) across different AI agents like Claude Code and CodeBuddy Code.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logger.Init(verbose)
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
		if verbose {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "output verbose logs")
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
