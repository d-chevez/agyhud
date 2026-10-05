package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/d-chevez/agyhud/internal/widgets"
)

func (m *Model) renderWidgetsTab() string {
	var items []string

	if m.config.Responsive.Mode == "dynamic_wrap" {
		items = append(items, "── DYNAMIC FLOW (Widgets wrap onto new lines automatically based on terminal width) ──")
	}

	itemIdx := 0
	for r, row := range m.config.Rows {
		items = append(items, fmt.Sprintf("── ROW %d (%d widgets) ──", r+1, len(row)))
		for _, w := range row {
			prefix := ""
			if m.isReordering && itemIdx == m.widgetsCursor {
				prefix = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warning)).Bold(true).Render("↕ [MOVING] ")
			}

			mergeTag := ""
			if w.Merge {
				mergeTag = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warning)).Render(" [MERGE]")
			}

			rawTag := ""
			if w.RawValue {
				rawTag = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Accent)).Render(" [RAW]")
			}

			extra := ""
			if w.Type == "custom_symbol" {
				extra = fmt.Sprintf(" '%s'", w.CustomSymbol)
			} else if w.Type == "separator" {
				extra = fmt.Sprintf(" '%s' (Spacer)", w.Separator)
			} else if w.Label != "" {
				extra = fmt.Sprintf(" '%s'", w.Label)
			}

			items = append(items, fmt.Sprintf("%s%-16s%s%s%s", prefix, w.Type, extra, mergeTag, rawTag))
			itemIdx++
		}
	}

	return m.renderStructuredList(items, m.widgetsCursor)
}

func (m *Model) renderAddWidgetModal() string {
	var items []string
	items = append(items, "── SELECT WIDGET FROM CATALOG (Enter to Add, Esc to Cancel) ──")
	for i, t := range widgets.AvailableWidgetTypes {
		sel := "   "
		if i == m.catalogCursor {
			sel = " ▶ "
		}
		items = append(items, sel+t)
	}
	return strings.Join(items, "\n")
}
