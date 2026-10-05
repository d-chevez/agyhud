package tui

import (
	"fmt"
	"strings"

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

	widgetsBadge := fmt.Sprintf("[%d of %d active]", m.getActiveWidgetsCount(), m.getTotalWidgetsCount())

	items := []mainMenuItem{
		{
			number:      "1",
			title:       "🔌 Terminal & Integration",
			badge:       hookBadge,
			badgeColor:  hookColor,
			description: "Manage Antigravity CLI hook, font glyph sets (Nerd / ASCII), and Git cache.",
		},
		{
			number:      "2",
			title:       "📐 HUD Layout & Flow",
			badge:       layoutBadge,
			badgeColor:  m.config.Theme.Accent,
			description: "Configure dynamic wrapping, edge-to-edge full width, and width breakpoints.",
		},
		{
			number:      "3",
			title:       "🧩 Widgets (Lego Builder)",
			badge:       widgetsBadge,
			badgeColor:  m.config.Theme.Success,
			description: "Arrange statusline widgets, customize labels, colors, and merge behavior.",
		},
		{
			number:      "4",
			title:       "🎨 Appearance & Themes",
			badge: func() string {
				if m.config.ThemeMode == "custom" {
					return "[Custom Theme]"
				}
				return "[Antigravity Dark]"
			}(),
			badgeColor:  m.config.Theme.Accent,
			description: "Manage Antigravity Dark theme and customize individual widget colors.",
		},
		{
			number:      "5",
			title:       "💾 Save Configuration",
			badge:       "[Write to Disk]",
			badgeColor:  m.config.Theme.Dim,
			description: "Persist all current settings to config.json.",
		},
		{
			number:      "6",
			title:       "🚪 Exit",
			badge:       "[Quit Studio]",
			badgeColor:  m.config.Theme.Dim,
			description: "Close agyhud studio and return to terminal.",
		},
	}

	var rendered []string
	selTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Accent))
	normTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Text))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim))

	for i, item := range items {
		badgeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(item.badgeColor))
		if i == m.mainCursor {
			header := selTitleStyle.Render("▶ " + item.number + ". " + item.title)
			line1 := fmt.Sprintf("%-50s %s", header, badgeStyle.Render(item.badgeColor, item.badge))
			line2 := descStyle.Render("     " + item.description)
			rendered = append(rendered, line1+"\n"+line2)
		} else {
			header := normTitleStyle.Render("  " + item.number + ". " + item.title)
			line1 := fmt.Sprintf("%-50s %s", header, badgeStyle.Render(item.badgeColor, item.badge))
			line2 := descStyle.Render("     " + item.description)
			rendered = append(rendered, line1+"\n"+line2)
		}
	}

	return strings.Join(rendered, "\n\n")
}
