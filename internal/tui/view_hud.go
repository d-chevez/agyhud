package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/d-chevez/agyhud/internal/config"
)

func (m *Model) renderHUDTab() string {
	var sections []string

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Accent))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim))

	// Section 1: Layout Flow Mode
	flowBadge := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Success)).Bold(true).Render("[Dynamic Auto-Wrap]")
	if m.config.Responsive.Mode == config.LayoutModeManual {
		flowBadge = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warning)).Bold(true).Render("[Manual Fixed Rows]")
	}
	secFlow := fmt.Sprintf("1. %s %s\n   %s",
		headerStyle.Render("Layout Flow Behavior:"),
		flowBadge,
		descStyle.Render("Dynamic Auto-Wrap automatically flows widgets into new rows if terminal is narrow. Manual uses strict user rows."),
	)
	sections = append(sections, secFlow)

	// Section 2: Full-Width Display
	fullBadge := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim)).Render("[OFF]")
	if m.config.Responsive.FullWidth {
		fullBadge = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Success)).Bold(true).Render("[ON (Span Width)]")
	}
	secFull := fmt.Sprintf("2. %s %s\n   %s",
		headerStyle.Render("Full-Width Expansion:"),
		fullBadge,
		descStyle.Render("Expands statusline whitespace across the entire terminal width for a complete edge-to-edge HUD."),
	)
	sections = append(sections, secFull)

	// Section 3: Breakpoint Width
	bpBadge := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Text)).Bold(true).Render(fmt.Sprintf("[%d cols]", m.config.Responsive.BreakpointWidth))
	secBP := fmt.Sprintf("3. %s %s\n   %s",
		headerStyle.Render("Responsive Breakpoint:"),
		bpBadge,
		descStyle.Render("Threshold width below which compact mode or row wrapping occurs. Adjust with [← / →]"),
	)
	sections = append(sections, secBP)

	return m.renderStructuredList(sections, m.hudCursor)
}
