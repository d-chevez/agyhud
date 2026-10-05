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
				prefix = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warning)).Bold(true).Render("[MOVING] ")
			}

			if w.Type == "row_break" {
				currentRow++
				tag := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warning)).Bold(true).Render("[Row Break / Next Row]")
				items = append(items, fmt.Sprintf("%s%-20s %-24s %s", prefix, "row_break", "", tag))
				items = append(items, fmt.Sprintf("── ROW %d ──", currentRow))
				itemIdx++
				continue
			}

			var badges []string

			if w.Type == "context_bar" {
				disp := w.ContextDisplay
				if disp == "" {
					disp = "both"
				}
				mode := w.ContextMode
				if mode == "" {
					mode = "used"
				}
				dispLabel := "BAR+%"
				if disp == "bar" {
					dispLabel = "BAR"
				} else if disp == "percentage" {
					dispLabel = "%"
				}
				modeLabel := "USED"
				if mode == "remaining" {
					modeLabel = "REMAINING"
				}
				badges = append(badges, lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Accent)).Render(fmt.Sprintf("[%s|%s]", dispLabel, modeLabel)))
			}

			if w.Merge {
				badges = append(badges, lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warning)).Render("[MERGE]"))
			}

			if w.Bold {
				badges = append(badges, lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warning)).Bold(true).Render("[BOLD]"))
			}

			if w.RawValue {
				badges = append(badges, lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Accent)).Render("[RAW]"))
			}

			if w.RawValue && w.Enclose {
				badges = append(badges, lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Accent)).Render("[ENCLOSE]"))
			}

			extraCol := fmt.Sprintf("%-24s", "")
			if w.Type == "custom_symbol" {
				extraCol = fmt.Sprintf("%-24s", fmt.Sprintf("'%s'", w.CustomSymbol))
			} else if w.Type == "separator" {
				extraCol = fmt.Sprintf("%-24s", fmt.Sprintf("'%s' (Spacer)", w.Separator))
			} else if !w.RawValue {
				if w.Label != "" {
					extraCol = fmt.Sprintf("%-24s", fmt.Sprintf("'%s'", w.Label))
				} else if def := widgets.DefaultLabel(w.Type); def != "" {
					formatted := fmt.Sprintf("%-24s", fmt.Sprintf("'%s'", def))
					extraCol = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim)).Render(formatted)
				}
			}

			typeCol := fmt.Sprintf("%-20s", w.Type)
			badgeStr := strings.Join(badges, " ")

			items = append(items, fmt.Sprintf("%s%s %s %s", prefix, typeCol, extraCol, badgeStr))
			itemIdx++
		}
	}

	return m.renderStructuredList(items, m.widgetsCursor)
}

func (m *Model) renderWidgetCategories() string {
	var items []string
	items = append(items, "── SELECT WIDGET CATEGORY (Enter to Browse, Esc to Cancel) ──")

	badgeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Accent)).Bold(true)
	for i, cat := range widgets.CatalogCategories {
		countBadge := badgeStyle.Render(fmt.Sprintf("[%d widgets]", len(cat.Widgets)))
		items = append(items, fmt.Sprintf("%d. %-32s %s", i+1, cat.Name, countBadge))
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

	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim))
	for _, item := range cat.Widgets {
		items = append(items, fmt.Sprintf("%-24s %s", item.Type, descStyle.Render(item.Description)))
	}

	return m.renderStructuredList(items, m.catalogCursor)
}
