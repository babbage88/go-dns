package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var (
	// clean-zones flags
	file    string
	timeout time.Duration
	workers int
	dryRun  bool
)

var rootCmd = &cobra.Command{
	Use:   "dnsctl",
	Short: "DNS zone management tool",
}

var cleanZonesCmd = &cobra.Command{
	Use:   "clean-zones",
	Short: "Clean and validate DNS zones",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCleanZones(file, timeout, workers, dryRun)
	},
}

var completionCmd = &cobra.Command{
	Use:    "completion",
	Short:  "Generate shell completion script",
	Hidden: true,
}

var bashCompletionCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generate bash completion script",
	RunE: func(cmd *cobra.Command, args []string) error {
		return rootCmd.GenBashCompletion(cmd.OutOrStdout())
	},
}

var zshCompletionCmd = &cobra.Command{
	Use:   "zsh",
	Short: "Generate zsh completion script",
	RunE: func(cmd *cobra.Command, args []string) error {
		return rootCmd.GenZshCompletion(cmd.OutOrStdout())
	},
}

func init() {
	cleanZonesCmd.Flags().StringVar(&file, "file", "", "YAML file to process (required)")
	cleanZonesCmd.Flags().DurationVar(&timeout, "timeout", 2*time.Second, "Ping timeout")
	cleanZonesCmd.Flags().IntVar(&workers, "workers", 8, "Number of parallel ping workers")
	cleanZonesCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Do not modify output")
	cleanZonesCmd.MarkFlagRequired("file")

	completionCmd.AddCommand(bashCompletionCmd, zshCompletionCmd)
	rootCmd.AddCommand(cleanZonesCmd, completionCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
