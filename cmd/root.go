package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "pbrew",
	Short: "pbrew — AI component manager for agents",
	Long:  "pbrew manages AI components (rules, skills, hooks) across different AI agents like Claude Code and CodeBuddy Code.",
	SilenceUsage:  true,
	SilenceErrors: true,
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

// Verbose returns whether verbose mode is enabled.
func Verbose() bool {
	return verbose
}
