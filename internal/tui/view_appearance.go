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

	items = append(items, headerStyle.Render("── WIDGET COLORS (Press Enter to set Hex color, leave empty for Auto) ──"))

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

		// Resolve active effective color
		effectiveColor := w.Color
		if effectiveColor == "" {
			effectiveColor = m.config.Theme.Accent
			if w.Type == "separator" || w.Type == "agent_state" {
				effectiveColor = m.config.Theme.Dim
			}
		}

		// 1. Text of widget styled directly with its color (Realtime visual color)
		colorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(effectiveColor)).Bold(true)
		typeText := colorStyle.Render(fmt.Sprintf("%-16s", w.Type))

		// 2. Readable live text preview styled directly in color
		livePreview := ""
		if renderer, ok := widgets.Registry[w.Type]; ok {
			livePreview = strings.TrimSpace(renderer.Render(ctx, w))
		}
		if livePreview == "" {
			livePreview = colorStyle.Render("(active)")
		}

		// 3. Hex specification (Only shown here)
		hexStr := ""
		if w.Color != "" {
			swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(w.Color)).Render("■■■")
			hexStr = fmt.Sprintf("[%s %s]", w.Color, swatch)
		} else {
			swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(effectiveColor)).Render("■■■")
			hexStr = fmt.Sprintf("[Auto %s %s]", effectiveColor, swatch)
		}

		rowPrefix := dimStyle.Render(fmt.Sprintf("R%d", ref.Row+1))
		previewBox := fmt.Sprintf("│ %-28s", livePreview)

		line := fmt.Sprintf("%s %s %s %s", rowPrefix, typeText, previewBox, hexStr)
		items = append(items, line)
	}

	return m.renderStructuredList(items, m.appearanceCursor)
}
