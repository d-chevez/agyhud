package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderAppearanceTab() string {
	var items []string

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Accent))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim))

	// Section 1: Active Theme Selection
	items = append(items, headerStyle.Render("── THEME MODE (Press Enter to select) ──"))
	if m.config.ThemeMode != "custom" {
		items = append(items, "● Antigravity Dark (Default - Native AGY palette)")
		items = append(items, "○ Custom Theme (Overrides active)")
	} else {
		items = append(items, "○ Antigravity Dark (Default - Native AGY palette)")
		items = append(items, "● Custom Theme (Overrides active)")
	}

	// Section 2: Active Widgets Color List
	items = append(items, headerStyle.Render("── WIDGET COLORS (Press Enter to customize color) ──"))
	widgetsList := m.getWidgetsList()
	for _, ref := range widgetsList {
		w := m.config.Rows[ref.Row][ref.Col]
		colorVal := w.Color
		swatch := ""

		nameStr := fmt.Sprintf("Row %d: %-16s", ref.Row+1, w.Type)
		if w.Label != "" {
			nameStr = fmt.Sprintf("Row %d: %-16s (%s)", ref.Row+1, w.Type, w.Label)
		} else if w.Type == "separator" {
			nameStr = fmt.Sprintf("Row %d: %-16s (%s)", ref.Row+1, w.Type, w.Separator)
		} else if w.Type == "custom_symbol" {
			nameStr = fmt.Sprintf("Row %d: %-16s (%s)", ref.Row+1, w.Type, w.CustomSymbol)
		}

		if colorVal != "" {
			swatch = " " + lipgloss.NewStyle().Foreground(lipgloss.Color(colorVal)).Render("■■■")
			items = append(items, fmt.Sprintf("%-34s [%-10s%s]", nameStr, colorVal, swatch))
		} else {
			defaultSwatch := " " + lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Accent)).Render("■■■")
			items = append(items, fmt.Sprintf("%-34s [%-10s%s]", nameStr, "Default", defaultSwatch))
		}
	}

	// Section 3: Granular Global Theme Palette
	items = append(items, headerStyle.Render("── GLOBAL THEME PALETTE (Press Enter to edit hex) ──"))
	for _, f := range m.colorFields {
		val := f.get(&m.config.Theme)
		swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(val)).Render("■■■")
		items = append(items, fmt.Sprintf("%-20s [%-10s %s]", f.label+":", val, swatch))
	}

	_ = dimStyle
	return m.renderStructuredList(items, m.appearanceCursor)
}
