package cmd

import (
	"os"

	"github.com/d-chevez/agyhud/internal/tui"
	"github.com/spf13/cobra"
)

var (
	cfgPath string
	version = "0.1.0-alpha"
)

// RootCmd is the base command for agyhud.
var RootCmd = &cobra.Command{
	Use:   "agyhud",
	Short: "Modern HUD and statusline engine for Google Antigravity CLI",
	Long: `agyhud is a high-performance HUD and statusline for Google Antigravity CLI (agy).
When executed interactively without arguments, it launches the interactive configuration TUI.
When called with the 'render' subcommand, it consumes session JSON on stdin and renders ANSI output.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run(cfgPath)
	},
}

// Execute runs the root command.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "", "Path to custom configuration file")
	RootCmd.Version = version
}
