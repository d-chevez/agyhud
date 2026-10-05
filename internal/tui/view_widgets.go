package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/d-chevez/agyhud/internal/widgets"
)

func (m *Model) renderWidgetsTab() string {
	if m.mode == editInspector {
		return m.renderWidgetInspectorModal()
	}

	var items []string

	if m.config.Responsive.Mode == "dynamic_wrap" {
		items = append(items, "── DYNAMIC FLOW (Widgets wrap onto new lines automatically based on terminal width) ──")
	}

	for r, row := range m.config.Rows {
		items = append(items, fmt.Sprintf("── ROW %d (%d widgets) ──", r+1, len(row)))
		for _, w := range row {
			status := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim)).Render("[ ]")
			if w.Enabled {
				status = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Success)).Bold(true).Render("[x]")
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
				extra = fmt.Sprintf(" '%s'", w.Separator)
			} else if w.Label != "" {
				extra = fmt.Sprintf(" '%s'", w.Label)
			}

			colorTag := ""
			if w.Color != "" {
				swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(w.Color)).Render("■■")
				colorTag = fmt.Sprintf(" [%s %s]", w.Color, swatch)
			}

			items = append(items, fmt.Sprintf("%s %-16s%s%s%s%s", status, w.Type, extra, mergeTag, rawTag, colorTag))
		}
	}

	return m.renderStructuredList(items)
}

func (m *Model) renderWidgetInspectorModal() string {
	r, w := m.resolveWidgetIndices(m.cursor)
	if r < 0 || w < 0 {
		return "Widget not found."
	}
	wCfg := m.config.Rows[r][w]

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Accent))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim))
	optStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Text))

	var lines []string
	lines = append(lines, titleStyle.Render(fmt.Sprintf("── WIDGET INSPECTOR: %s (Row %d) ──", strings.ToUpper(wCfg.Type), r+1)))

	// Option 0: State
	enStr := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Success)).Render("ENABLED")
	if !wCfg.Enabled {
		enStr = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim)).Render("DISABLED")
	}
	lines = append(lines, fmt.Sprintf("1. Status:       [%s] (Enter/Space to toggle)", enStr))

	// Option 1: Label
	labelVal := wCfg.Label
	if labelVal == "" {
		labelVal = "(None)"
	}
	lines = append(lines, fmt.Sprintf("2. Label Prefix: [%s] (Enter to edit)", optStyle.Render(labelVal)))

	// Option 2: RawValue
	rawStr := "OFF (Display with label)"
	if wCfg.RawValue {
		rawStr = "ON (Bare value only)"
	}
	lines = append(lines, fmt.Sprintf("3. Raw Value:    [%s] (Enter to toggle)", optStyle.Render(rawStr)))

	// Option 3: Merge
	mergeStr := "OFF (Normal trailing space)"
	if wCfg.Merge {
		mergeStr = "ON (Fuse seamlessly with next widget)"
	}
	lines = append(lines, fmt.Sprintf("4. Merge:        [%s] (Enter to toggle)", optStyle.Render(mergeStr)))

	// Option 4: Color
	colorStr := "Theme Default"
	swatch := ""
	if wCfg.Color != "" {
		colorStr = wCfg.Color
		swatch = " " + lipgloss.NewStyle().Foreground(lipgloss.Color(wCfg.Color)).Render("■■■")
	}
	lines = append(lines, fmt.Sprintf("5. Color:        [%s%s] (Enter to edit Hex/ANSI)", optStyle.Render(colorStr), swatch))

	// Option 5: Symbol / Separator
	if wCfg.Type == "custom_symbol" {
		lines = append(lines, fmt.Sprintf("6. Symbol:       [%s] (Enter to change glyph)", optStyle.Render(wCfg.CustomSymbol)))
	} else if wCfg.Type == "separator" {
		lines = append(lines, fmt.Sprintf("6. Separator:    [%s] (Enter to change char)", optStyle.Render(wCfg.Separator)))
	}

	lines = append(lines, dimStyle.Render("── Actions: [Esc] Back to Widgets │ [d] Delete Widget ──"))

	return m.renderStructuredList(lines)
}

func (m *Model) renderAddWidgetModal() string {
	var items []string
	items = append(items, "── SELECT WIDGET TO ADD (Press Enter to Add, Esc to Cancel) ──")
	for i, t := range widgets.AvailableWidgetTypes {
		sel := "   "
		if i == m.catalogIndex {
			sel = " ▶ "
		}
		items = append(items, sel+t)
	}
	return strings.Join(items, "\n")
}
