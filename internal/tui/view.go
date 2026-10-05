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
		fmt.Sprintf("── LIVE PREVIEW (Terminal: %d cols │ Flow: %s) ", m.payload.TerminalWidth, modeDesc),
	)
	b.WriteString(previewHeader + "\n")
	b.WriteString(previewBox + "\n\n")

	// 3. Breadcrumb Path
	b.WriteString(m.renderBreadcrumbs() + "\n\n")

	// 4. Screen Body
	switch m.currentScreen() {
	case screenMainMenu:
		b.WriteString(m.renderMainMenu())
	case screenTerminal:
		b.WriteString(m.renderTerminalTab())
	case screenHUD:
		b.WriteString(m.renderHUDTab())
	case screenWidgets:
		b.WriteString(m.renderWidgetsTab())
	case screenWidgetCatalog:
		b.WriteString(m.renderAddWidgetModal())
	case screenAppearance:
		b.WriteString(m.renderAppearanceTab())
	}

	// 5. Active Text Input Prompt (if editing inline)
	if m.mode == editInputText {
		promptStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Warning))
		b.WriteString("\n\n" + promptStyle.Render(" ✍ Input: ") + m.textInput.View() + " (Enter to save, Esc to cancel)")
	} else if m.statusMsg != "" {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Success)).Bold(true)
		b.WriteString("\n\n" + statusStyle.Render(" "+m.statusMsg))
	} else {
		b.WriteString("\n")
	}

	// 6. Contextual Footer & Keybindings
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim))
	b.WriteString("\n" + footerStyle.Render(m.renderFooterKeybindings()))

	return b.String()
}

func (m *Model) renderBreadcrumbs() string {
	sep := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim)).Render(" › ")
	home := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Accent)).Render("agyhud")

	crumbs := []string{home}
	for i, s := range m.screenStack {
		if i == 0 {
			if len(m.screenStack) == 1 {
				crumbs = append(crumbs, lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Text)).Render("Main Menu"))
			}
			continue
		}
		switch s {
		case screenTerminal:
			crumbs = append(crumbs, "🔌 Terminal & Integration")
		case screenHUD:
			crumbs = append(crumbs, "📐 HUD Layout & Flow")
		case screenWidgets:
			crumbs = append(crumbs, "🧩 Widgets (CRUD & Layout)")
		case screenWidgetCatalog:
			crumbs = append(crumbs, "➕ Add Widget Catalog")
		case screenAppearance:
			crumbs = append(crumbs, "🎨 Appearance & Themes")
		}
	}

	for i := 1; i < len(crumbs); i++ {
		if i == len(crumbs)-1 {
			crumbs[i] = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Text)).Render(crumbs[i])
		} else {
			crumbs[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim)).Render(crumbs[i])
		}
	}

	return "📍 " + strings.Join(crumbs, sep)
}

func (m *Model) renderFooterKeybindings() string {
	if m.mode == editInputText {
		return "[Enter] Confirm Input │ [Esc] Cancel Input"
	}

	switch m.currentScreen() {
	case screenMainMenu:
		return "[↑/↓] Navigate │ [1-6/Enter] Open Category │ [s] Save Config │ [q/Esc] Quit"
	case screenTerminal, screenHUD:
		return "[↑/↓] Navigate │ [Enter/Space] Toggle │ [←/→] Adjust │ [Esc] Back to Menu │ [s] Save Config"
	case screenWidgets:
		return "[↑/↓] Navigate │ [Space] Toggle │ [K/J] Move Up/Down │ [a] Add │ [p] Add Spacer │ [d] Delete │ [m] Merge │ [e] Edit Label │ [Esc] Back"
	case screenWidgetCatalog:
		return "[↑/↓] Select Widget │ [Enter] Add to Row │ [Esc] Cancel"
	case screenAppearance:
		return "[↑/↓] Navigate │ [Enter] Select Theme / Set Widget Color │ [Esc] Back to Menu │ [s] Save Config"
	default:
		return "[↑/↓] Navigate │ [Enter] Select │ [Esc] Back │ [q] Quit"
	}
}

func (m *Model) renderStructuredList(items []string, activeCursor int) string {
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
		if itemIndex == activeCursor {
			rendered = append(rendered, selStyle.Render(" ▶ "+item))
		} else {
			rendered = append(rendered, normStyle.Render("   "+item))
		}
	}

	return strings.Join(rendered, "\n")
}
