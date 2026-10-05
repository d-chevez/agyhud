package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/d-chevez/agyhud/internal/engine"
)

func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// 1. Header & Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Accent))
	b.WriteString(titleStyle.Render(" 🚀 agyhud — Interactive Configuration & HUD Studio") + "\n\n")

	// 2. Live Preview Section
	previewContent := engine.Render(m.payload, m.config)
	modeDesc := "Dynamic Wrap"
	if m.config.Responsive.Mode == "manual_rows" {
		modeDesc = "Manual Rows"
	}
	previewBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.config.Theme.Accent)).
		Padding(0, 1).
		Render(previewContent)

	previewHeader := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim)).Render(
		fmt.Sprintf("── LIVE PREVIEW (Width: %d cols │ %s) ", m.payload.TerminalWidth, modeDesc),
	)
	b.WriteString(previewHeader + "\n")
	b.WriteString(previewBox + "\n\n")

	// 3. Navigation Tabs
	b.WriteString(m.renderTabs() + "\n\n")

	// 4. Tab Body or Active Modal
	if m.mode == editAddWidget {
		b.WriteString(m.renderAddWidgetModal())
	} else {
		switch m.activeTab {
		case tabTerminal:
			b.WriteString(m.renderTerminalTab())
		case tabHUD:
			b.WriteString(m.renderHUDTab())
		case tabWidgets:
			b.WriteString(m.renderWidgetsTab())
		case tabAppearance:
			b.WriteString(m.renderAppearanceTab())
		}
	}

	// 5. Active Text Input Prompt (if editing)
	if m.mode == editInputText {
		promptStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Warning))
		b.WriteString("\n" + promptStyle.Render(" Input: ") + m.textInput.View() + " (Enter to save, Esc to cancel)\n")
	} else if m.statusMsg != "" {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Success)).Bold(true)
		b.WriteString("\n" + statusStyle.Render(m.statusMsg) + "\n")
	} else {
		b.WriteString("\n\n")
	}

	// 6. Contextual Footer & Keybindings
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim))
	if m.mode == editAddWidget {
		b.WriteString(footerStyle.Render("[↑/↓] Select Widget │ [Enter] Add to Row │ [Esc] Cancel"))
	} else if m.mode == editInspector {
		b.WriteString(footerStyle.Render("[↑/↓] Navigate │ [Enter] Edit/Toggle │ [d] Delete Widget │ [Esc] Close Inspector"))
	} else if m.activeTab == tabWidgets {
		b.WriteString(footerStyle.Render("[↑/↓] Select │ [Enter] Open Inspector │ [Space] Toggle │ [a] Add │ [R] Add Row │ [d] Delete │ [s] Save │ [q] Exit"))
	} else {
		b.WriteString(footerStyle.Render("[Tab] Switch Tab │ [↑/↓] Navigate │ [Enter/Space] Select/Toggle │ [s] Save Config │ [q] Exit"))
	}

	return b.String()
}

func (m *Model) renderTabs() string {
	tabs := []string{
		"1. Terminal & Integration",
		"2. HUD Layout",
		"3. Widgets (Lego Builder)",
		"4. Appearance & Themes",
	}
	var rendered []string

	for i, t := range tabs {
		if tabIndex(i) == m.activeTab {
			active := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#1a1b26")).
				Background(lipgloss.Color(m.config.Theme.Accent)).
				Padding(0, 2).
				Render(t)
			rendered = append(rendered, active)
		} else {
			inactive := lipgloss.NewStyle().
				Foreground(lipgloss.Color(m.config.Theme.Dim)).
				Padding(0, 2).
				Render(t)
			rendered = append(rendered, inactive)
		}
	}

	return strings.Join(rendered, " ")
}

func (m *Model) renderStructuredList(items []string) string {
	var rendered []string
	selStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Accent)).Bold(true)
	normStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Text))

	cursorOffset := 0
	for _, item := range items {
		if strings.HasPrefix(item, "──") {
			dim := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim)).Render(item)
			rendered = append(rendered, dim)
			cursorOffset++
			continue
		}

		itemIndex := len(rendered) - cursorOffset
		activeCursor := m.cursor
		if m.mode == editInspector {
			activeCursor = m.inspectorCursor
		}

		if itemIndex == activeCursor {
			rendered = append(rendered, selStyle.Render(" ▶ "+item))
		} else {
			rendered = append(rendered, normStyle.Render("   "+item))
		}
	}

	return strings.Join(rendered, "\n")
}
