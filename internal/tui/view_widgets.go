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

	currentRow := 1
	items = append(items, fmt.Sprintf("── ROW %d ──", currentRow))

	itemIdx := 0
	for _, row := range m.config.Rows {
		for _, w := range row {
			prefix := ""
			if m.isReordering && itemIdx == m.widgetsCursor {
				prefix = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warning)).Bold(true).Render("↕ [MOVING] ")
			}

			if w.Type == "row_break" {
				currentRow++
				tag := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warning)).Bold(true).Render("[Row Break / Next Row ↓]")
				items = append(items, fmt.Sprintf("%s↵ %-18s %s", prefix, "row_break", tag))
				items = append(items, fmt.Sprintf("── ROW %d ──", currentRow))
				itemIdx++
				continue
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

			items = append(items, fmt.Sprintf("%s%-18s%s%s%s", prefix, w.Type, extra, mergeTag, rawTag))
			itemIdx++
		}
	}

	return m.renderStructuredList(items, m.widgetsCursor)
}

func (m *Model) renderAddWidgetModal() string {
	var items []string
	items = append(items, "── SELECT WIDGET FROM CATALOG (Enter to Add, Esc to Cancel) ──")

	flatIdx := 0
	for _, cat := range widgets.CatalogCategories {
		items = append(items, fmt.Sprintf("── %s ──", cat.Name))
		for _, item := range cat.Widgets {
			sel := "   "
			if flatIdx == m.catalogCursor {
				sel = " ▶ "
			}
			nameStyled := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Text)).Render(fmt.Sprintf("%-22s", item.Type))
			descStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim)).Render(item.Description)
			items = append(items, fmt.Sprintf("%s%s %s", sel, nameStyled, descStyled))
			flatIdx++
		}
	}
	return strings.Join(items, "\n")
}
