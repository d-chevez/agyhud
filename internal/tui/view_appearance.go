package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderAppearanceTab() string {
	var items []string

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim))

	items = append(items, "── WIDGET COLORS (Press Enter to set Hex color, leave empty for Auto) ──")

	widgetsList := m.getWidgetsList()
	if len(widgetsList) == 0 {
		items = append(items, dimStyle.Render("No widgets configured. Add widgets in the Widgets tab first."))
		return m.renderStructuredList(items, 0)
	}

	for _, ref := range widgetsList {
		w := m.config.Rows[ref.Row][ref.Col]

		// Resolve active effective color
		effectiveColor := w.Color
		if effectiveColor == "" {
			effectiveColor = m.config.Theme.Accent
			if w.Type == "separator" || w.Type == "agent_state" {
				effectiveColor = m.config.Theme.Dim
			}
		}

		// 1. Widget title/type styled directly with its selected color
		displayType := w.Type
		if w.Type == "separator" && w.Separator != "" {
			displayType = fmt.Sprintf("separator '%s'", w.Separator)
		} else if w.Type == "custom_symbol" && w.CustomSymbol != "" {
			displayType = fmt.Sprintf("custom_symbol '%s'", w.CustomSymbol)
		} else if w.Label != "" {
			displayType = fmt.Sprintf("%s (%s)", w.Type, w.Label)
		}

		colorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(effectiveColor)).Bold(true)
		typeText := colorStyle.Render(fmt.Sprintf("%-28s", displayType))

		// 2. Hex code + visual color swatch
		hexStr := ""
		if w.Color != "" {
			swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(w.Color)).Render("■■■")
			hexStr = fmt.Sprintf("[%s %s]", w.Color, swatch)
		} else {
			swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(effectiveColor)).Render("■■■")
			hexStr = fmt.Sprintf("[Auto %s %s]", effectiveColor, swatch)
		}

		rowPrefix := fmt.Sprintf("Row %d:", ref.Row+1)
		rowCol := dimStyle.Render(fmt.Sprintf("%-8s", rowPrefix))
		line := fmt.Sprintf("%s %s %s", rowCol, typeText, hexStr)
		items = append(items, line)
	}

	return m.renderStructuredList(items, m.appearanceCursor)
}
