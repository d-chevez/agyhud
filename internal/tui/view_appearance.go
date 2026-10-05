package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/d-chevez/agyhud/internal/widgets"
)

func (m *Model) renderAppearanceTab() string {
	var items []string

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Accent))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim))
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Text))

	items = append(items, headerStyle.Render("── CONFIGURED WIDGET COLORS (Press Enter to edit hex/color) ──"))

	widgetsList := m.getWidgetsList()
	if len(widgetsList) == 0 {
		items = append(items, dimStyle.Render("No widgets configured. Add widgets in the Widgets tab first."))
		return m.renderStructuredList(items, 0)
	}

	ctx := widgets.Context{
		Payload: m.payload,
		Config:  m.config,
	}

	for _, ref := range widgetsList {
		w := m.config.Rows[ref.Row][ref.Col]

		// 1. Live legible text output from the widget engine
		livePreview := ""
		if renderer, ok := widgets.Registry[w.Type]; ok {
			livePreview = strings.TrimSpace(renderer.Render(ctx, w))
		}
		if livePreview == "" {
			livePreview = "(empty)"
		}

		// 2. Color badge with visual swatch
		colorBadge := ""
		if w.Color != "" {
			swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(w.Color)).Render("■■■")
			colorBadge = fmt.Sprintf("[%s %s]", w.Color, swatch)
		} else {
			defaultColor := m.config.Theme.Accent
			if w.Type == "separator" || w.Type == "agent_state" {
				defaultColor = m.config.Theme.Dim
			}
			swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(defaultColor)).Render("■■■")
			colorBadge = fmt.Sprintf("[Auto %s]", swatch)
		}

		// 3. Structured line display
		rowPrefix := dimStyle.Render(fmt.Sprintf("R%d", ref.Row+1))
		typeText := labelStyle.Render(fmt.Sprintf("%-16s", w.Type))
		previewBox := fmt.Sprintf("│ Preview: %-26s", livePreview)

		line := fmt.Sprintf("%s %s %s %s", rowPrefix, typeText, previewBox, colorBadge)
		items = append(items, line)
	}

	return m.renderStructuredList(items, m.appearanceCursor)
}
