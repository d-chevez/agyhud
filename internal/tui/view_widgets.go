package tui

import (
	"fmt"

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

			boldTag := ""
			if w.Bold {
				boldTag = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warning)).Bold(true).Render(" [BOLD]")
			}

			rawTag := ""
			if w.RawValue {
				rawTag = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Accent)).Render(" [RAW]")
			}

			encloseTag := ""
			if w.RawValue && w.Enclose {
				encloseTag = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Accent)).Render(" [ENCLOSE]")
			}

			extra := ""
			if w.Type == "custom_symbol" {
				extra = fmt.Sprintf(" '%s'", w.CustomSymbol)
			} else if w.Type == "separator" {
				extra = fmt.Sprintf(" '%s' (Spacer)", w.Separator)
			} else if !w.RawValue {
				if w.Label != "" {
					extra = fmt.Sprintf(" '%s'", w.Label)
				} else if def := widgets.DefaultLabel(w.Type); def != "" {
					extra = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim)).Render(fmt.Sprintf(" '%s'", def))
				}
			}

			items = append(items, fmt.Sprintf("%s%-18s%s%s%s%s%s", prefix, w.Type, extra, mergeTag, boldTag, rawTag, encloseTag))
			itemIdx++
		}
	}

	return m.renderStructuredList(items, m.widgetsCursor)
}

func (m *Model) renderWidgetCategories() string {
	var items []string
	items = append(items, "── SELECT WIDGET CATEGORY (Enter to Browse, Esc to Cancel) ──")

	for _, cat := range widgets.CatalogCategories {
		items = append(items, fmt.Sprintf("%-32s (%d widgets)", cat.Name, len(cat.Widgets)))
	}

	return m.renderStructuredList(items, m.catalogCategoryCursor)
}

func (m *Model) renderWidgetCatalog() string {
	if m.catalogCategoryCursor < 0 || m.catalogCategoryCursor >= len(widgets.CatalogCategories) {
		m.catalogCategoryCursor = 0
	}
	cat := widgets.CatalogCategories[m.catalogCategoryCursor]

	var items []string
	items = append(items, fmt.Sprintf("── CATEGORY: %s (Enter to Add, Esc to Back) ──", cat.Name))

	for _, item := range cat.Widgets {
		items = append(items, fmt.Sprintf("%-22s %s", item.Type, item.Description))
	}

	return m.renderStructuredList(items, m.catalogCursor)
}
