package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/d-chevez/agyhud/internal/config"
)

func (m *Model) renderTerminalTab() string {
	var sections []string

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Accent))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim))

	// Section 1: Antigravity CLI Hook
	hookBadge := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Success)).Bold(true).Render("[Active]")
	if !m.hookStatus.Active || !m.hookStatus.IsAgyhud {
		hookBadge = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warning)).Render("[Not Configured]")
	}
	secHook := fmt.Sprintf("1. %-32s %s\n     %s",
		headerStyle.Render("Antigravity Hook Integration"),
		hookBadge,
		descStyle.Render("Registers 'agyhud render' into Antigravity settings.json (Press Enter to toggle)"),
	)
	sections = append(sections, secHook)

	// Section 2: Font Glyphs
	fontBadge := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Accent)).Bold(true).Render("[Nerd Font]")
	if m.config.IconSet == config.IconSetClassic {
		fontBadge = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Text)).Bold(true).Render("[Classic (ASCII)]")
	}
	secFont := fmt.Sprintf("2. %-32s %s\n     %s",
		headerStyle.Render("Font & Glyphs Mode"),
		fontBadge,
		descStyle.Render("Switch between rich Nerd Font devicons or universal ASCII symbols for unpatched fonts"),
	)
	sections = append(sections, secFont)

	// Section 3: Git Telemetry Refresh
	gitBadge := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Text)).Bold(true).Render(fmt.Sprintf("[%d seconds]", m.config.Git.RefreshSeconds))
	secGit := fmt.Sprintf("3. %-32s %s\n     %s",
		headerStyle.Render("Git Telemetry Cache Window"),
		gitBadge,
		descStyle.Render("TTL cache interval (1s to 10s) to prevent disk overhead in large repos. Adjust with [← / →]"),
	)
	sections = append(sections, secGit)

	return m.renderStructuredList(sections, m.terminalCursor)
}
