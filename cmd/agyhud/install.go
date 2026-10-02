package cmd

import (
	"fmt"

	"github.com/d-chevez/agyhud/internal/installer"
	"github.com/spf13/cobra"
)

var (
	installClassic bool
	customBinPath  string
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Configure agyhud as the active statusLine hook in Antigravity CLI",
	Long: `Discovers ~/.gemini/antigravity-cli/settings.json, creates a safety backup,
and registers 'agyhud render' as the statusLine command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configuredCmd, err := installer.Install(customBinPath, installClassic)
		if err != nil {
			return fmt.Errorf("installation failed: %w", err)
		}

		fmt.Println("✓ agyhud successfully integrated with Antigravity CLI!")
		fmt.Printf("  Configured command: %s\n", configuredCmd)
		fmt.Println("  Restart 'agy' or start a new prompt to see your new HUD in action.")
		return nil
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Safely remove agyhud from Antigravity CLI settings",
	Long: `Restores the original statusLine configuration in settings.json
or removes the hook if none was previously set.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := installer.Uninstall(); err != nil {
			return fmt.Errorf("uninstall failed: %w", err)
		}

		fmt.Println("✓ agyhud successfully removed from Antigravity CLI settings.")
		return nil
	},
}

func init() {
	installCmd.Flags().BoolVarP(&installClassic, "classic", "C", false, "Configure with classic ASCII/Unicode mode by default")
	installCmd.Flags().StringVar(&customBinPath, "bin", "", "Specify custom path to agyhud binary")

	RootCmd.AddCommand(installCmd)
	RootCmd.AddCommand(uninstallCmd)
}
