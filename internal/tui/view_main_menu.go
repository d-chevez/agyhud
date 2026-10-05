package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/d-chevez/agyhud/internal/config"
)

type mainMenuItem struct {
	number      string
	title       string
	badge       string
	badgeColor  string
	description string
}

func (m *Model) renderMainMenu() string {
	// Dynamically compute badges based on current config state
	hookBadge := "[○ Inactive]"
	hookColor := m.config.Theme.Dim
	if m.hookStatus.Active && m.hookStatus.IsAgyhud {
		hookBadge = "[● Active]"
		hookColor = m.config.Theme.Success
	}

	layoutBadge := "[Dynamic Auto-Wrap]"
	if m.config.Responsive.Mode == config.LayoutModeManual {
		layoutBadge = "[Manual Rows]"
	}

	widgetsBadge := fmt.Sprintf("[%d widgets]", m.getTotalWidgetsCount())

	items := []mainMenuItem{
		{
			number:      "1",
			title:       "Terminal & Integration",
			badge:       hookBadge,
			badgeColor:  hookColor,
			description: "Manage Antigravity CLI hook, font glyph sets (Nerd / ASCII), and Git cache.",
		},
		{
			number:      "2",
			title:       "HUD Layout & Flow",
			badge:       layoutBadge,
			badgeColor:  m.config.Theme.Accent,
			description: "Configure dynamic wrapping, edge-to-edge full width, and width breakpoints.",
		},
		{
			number:      "3",
			title:       "Widgets (Lego Builder)",
			badge:       widgetsBadge,
			badgeColor:  m.config.Theme.Success,
			description: "Add, delete, reorder positions, add spacers, and customize labels.",
		},
		{
			number:      "4",
			title:       "Widget Colors",
			badge:       "[Live Palette]",
			badgeColor:  m.config.Theme.Accent,
			description: "Customize individual colors for each statusline widget with live previews.",
		},
		{
			number:      "5",
			title:       "Save Configuration",
			badge:       "[Write to Disk]",
			badgeColor:  m.config.Theme.Dim,
			description: "Persist all current settings to config.json.",
		},
		{
			number:      "6",
			title:       "Exit",
			badge:       "[Quit Studio]",
			badgeColor:  m.config.Theme.Dim,
			description: "Close agyhud studio and return to terminal.",
		},
	}

	var sections []string
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Text))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim))

	for _, item := range items {
		badgeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(item.badgeColor)).Bold(true)
		title := headerStyle.Render(item.title)
		line1 := fmt.Sprintf("%s. %-32s %s", item.number, title, badgeStyle.Render(item.badge))
		line2 := descStyle.Render("     " + item.description)
		sections = append(sections, line1+"\n"+line2)
	}

	return m.renderStructuredList(sections, m.mainCursor)
}
