package cmd

import (
	"fmt"
	"os"

	"github.com/d-chevez/agyhud/internal/config"
	"github.com/d-chevez/agyhud/internal/engine"
	"github.com/d-chevez/agyhud/internal/payload"
	"github.com/spf13/cobra"
)

var (
	classicFlag bool
)

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render statusline from Antigravity CLI session payload (stdin)",
	Long:  `Reads the Antigravity session JSON payload from stdin, applies configuration, and emits formatted ANSI output to stdout.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure standard terminal statusline contract: exit 0 cleanly on any failure
		p, err := payload.Parse(os.Stdin)
		if err != nil || p == nil {
			// If stdin was empty or invalid JSON, silently exit without breaking CLI UI
			return
		}

		cfg, err := config.Load(cfgPath)
		if err != nil || cfg == nil {
			cfg = config.DefaultConfig()
		}

		if classicFlag {
			cfg.IconSet = config.IconSetClassic
		}

		out := engine.Render(p, cfg)
		if out != "" {
			fmt.Print(out)
		}
	},
}

func init() {
	renderCmd.Flags().BoolVarP(&classicFlag, "classic", "C", false, "Force classic Unicode/ASCII glyphs instead of Nerd Fonts")
	RootCmd.AddCommand(renderCmd)
}
