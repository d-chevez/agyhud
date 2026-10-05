package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/d-chevez/agyhud/internal/config"
	"github.com/d-chevez/agyhud/internal/installer"
	"github.com/d-chevez/agyhud/internal/widgets"
)

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.mode == editInputText {
		return m.updateModalInput(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.payload.TerminalWidth = msg.Width

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "q":
			if m.currentScreen() == screenMainMenu {
				m.quitting = true
				return m, tea.Quit
			}
			m.popScreen()
			return m, nil

		case "esc":
			if m.currentScreen() == screenWidgets && m.isReordering {
				m.isReordering = false
				m.statusMsg = "Exited reorder mode."
				return m, nil
			}
			if !m.popScreen() {
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil

		case "up", "k":
			if m.currentScreen() == screenWidgets && m.isReordering {
				m.moveWidget(-1)
			} else {
				m.moveCursor(-1)
			}

		case "down", "j":
			if m.currentScreen() == screenWidgets && m.isReordering {
				m.moveWidget(1)
			} else {
				m.moveCursor(1)
			}

		case "shift+up", "K":
			if m.currentScreen() == screenWidgets {
				m.moveWidget(-1)
			}

		case "shift+down", "J":
			if m.currentScreen() == screenWidgets {
				m.moveWidget(1)
			}

		case "left", "h":
			m.handleHorizontalAdjust(-1)

		case "right", "l":
			m.handleHorizontalAdjust(1)

		case "1", "2", "3", "4", "5", "6":
			if m.currentScreen() == screenMainMenu {
				num := int(msg.String()[0] - '1')
				m.mainCursor = num
				return m.handleEnterSelection()
			}
			if m.currentScreen() == screenWidgetCategories {
				num := int(msg.String()[0] - '1')
				if num >= 0 && num < len(widgets.CatalogCategories) {
					m.catalogCategoryCursor = num
					m.catalogCursor = 0
					m.pushScreen(screenWidgetCatalog)
					return m, nil
				}
			}

		case "enter":
			if m.currentScreen() == screenWidgets && m.isReordering {
				m.isReordering = false
				m.statusMsg = "✓ Widget position placed!"
				return m, nil
			}
			return m.handleEnterSelection()

		case " ":
			if m.currentScreen() == screenWidgets {
				m.isReordering = !m.isReordering
				if m.isReordering {
					r, w := m.resolveWidgetIndices(m.widgetsCursor)
					name := "widget"
					if r >= 0 && w >= 0 {
						name = m.config.Rows[r][w].Type
					}
					m.statusMsg = fmt.Sprintf("↕ Moving '%s' (use ↑/↓ to move, Space/Enter to place)", name)
				} else {
					m.statusMsg = "✓ Widget position placed!"
				}
				return m, nil
			}
			return m.handleEnterSelection()

		case "a":
			if m.currentScreen() == screenWidgets {
				m.isReordering = false
				m.catalogCategoryCursor = 0
				m.catalogCursor = 0
				m.pushScreen(screenWidgetCategories)
				return m, nil
			}

		case "p", "+":
			if m.currentScreen() == screenWidgets {
				m.addSpacer()
				return m, nil
			}

		case "d", "delete":
			if m.currentScreen() == screenWidgets {
				m.isReordering = false
				m.deleteCurrentWidget()
			}

		case "m":
			if m.currentScreen() == screenWidgets {
				r, w := m.resolveWidgetIndices(m.widgetsCursor)
				if r >= 0 && w >= 0 {
					m.config.Rows[r][w].Merge = !m.config.Rows[r][w].Merge
					status := "OFF"
					if m.config.Rows[r][w].Merge {
						status = "ON (Merged)"
					}
					m.statusMsg = fmt.Sprintf("✓ Merge %s for %s", status, m.config.Rows[r][w].Type)
				}
			}

		case "b":
			if m.currentScreen() == screenWidgets {
				r, w := m.resolveWidgetIndices(m.widgetsCursor)
				if r >= 0 && w >= 0 {
					m.config.Rows[r][w].Bold = !m.config.Rows[r][w].Bold
					status := "OFF"
					if m.config.Rows[r][w].Bold {
						status = "ON (Bold)"
					}
					m.statusMsg = fmt.Sprintf("✓ Bold mode %s for '%s'", status, m.config.Rows[r][w].Type)
				}
				return m, nil
			}

		case "c":
			if m.currentScreen() == screenWidgets {
				r, w := m.resolveWidgetIndices(m.widgetsCursor)
				if r >= 0 && w >= 0 {
					wCfg := &m.config.Rows[r][w]
					wCfg.RawValue = true
					m.mode = editInputText
					m.inputTargetField = "widget_raw_prefix"
					m.textInput.Placeholder = "Enclose (e.g. [], (), <>, [ ], or empty to clear)"
					current := ""
					if wCfg.RawPrefix != "" || wCfg.RawSuffix != "" {
						current = wCfg.RawPrefix + wCfg.RawSuffix
					}
					m.textInput.SetValue(current)
					m.textInput.Focus()
					return m, textinput.Blink
				}
			}

		case "r":
			if m.currentScreen() == screenWidgets {
				r, w := m.resolveWidgetIndices(m.widgetsCursor)
				if r >= 0 && w >= 0 {
					m.config.Rows[r][w].RawValue = !m.config.Rows[r][w].RawValue
					status := "OFF"
					if m.config.Rows[r][w].RawValue {
						status = "ON (Raw)"
					}
					m.statusMsg = fmt.Sprintf("✓ Raw Value %s for %s", status, m.config.Rows[r][w].Type)
				}
			}

		case "e":
			if m.currentScreen() == screenWidgets {
				return m.startWidgetEditing()
			}

		}
	}

	return m, nil
}

func (m *Model) moveCursor(delta int) {
	switch m.currentScreen() {
	case screenMainMenu:
		m.mainCursor = clamp(m.mainCursor+delta, 0, 5)
	case screenTerminal:
		m.terminalCursor = clamp(m.terminalCursor+delta, 0, 2)
	case screenHUD:
		m.hudCursor = clamp(m.hudCursor+delta, 0, 2)
	case screenWidgets:
		count := m.getTotalWidgetsCount()
		max := 0
		if count > 0 {
			max = count - 1
		}
		m.widgetsCursor = clamp(m.widgetsCursor+delta, 0, max)
	case screenWidgetCategories:
		m.catalogCategoryCursor = clamp(m.catalogCategoryCursor+delta, 0, len(widgets.CatalogCategories)-1)
	case screenWidgetCatalog:
		if m.catalogCategoryCursor < 0 || m.catalogCategoryCursor >= len(widgets.CatalogCategories) {
			m.catalogCategoryCursor = 0
		}
		cat := widgets.CatalogCategories[m.catalogCategoryCursor]
		max := 0
		if len(cat.Widgets) > 0 {
			max = len(cat.Widgets) - 1
		}
		m.catalogCursor = clamp(m.catalogCursor+delta, 0, max)
	case screenAppearance:
		widgetsCount := len(m.getWidgetsList())
		max := 0
		if widgetsCount > 0 {
			max = widgetsCount - 1
		}
		m.appearanceCursor = clamp(m.appearanceCursor+delta, 0, max)
	}
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func (m *Model) handleHorizontalAdjust(delta int) {
	switch m.currentScreen() {
	case screenTerminal:
		if m.terminalCursor == 2 {
			newVal := m.config.Git.RefreshSeconds + delta
			if newVal >= 1 && newVal <= 10 {
				m.config.Git.RefreshSeconds = newVal
			}
		}
	case screenHUD:
		if m.hudCursor == 2 {
			newBP := m.config.Responsive.BreakpointWidth + (delta * 5)
			if newBP >= 40 && newBP <= 200 {
				m.config.Responsive.BreakpointWidth = newBP
			}
		}
	}
}



func (m *Model) handleEnterSelection() (tea.Model, tea.Cmd) {
	switch m.currentScreen() {
	case screenMainMenu:
		switch m.mainCursor {
		case 0:
			m.pushScreen(screenTerminal)
		case 1:
			m.pushScreen(screenHUD)
		case 2:
			m.pushScreen(screenWidgets)
		case 3:
			m.pushScreen(screenAppearance)
		case 4:
			m.saveConfig()
		case 5:
			m.quitting = true
			return m, tea.Quit
		}

	case screenTerminal:
		switch m.terminalCursor {
		case 0:
			if m.hookStatus.Active && m.hookStatus.IsAgyhud {
				_ = installer.Uninstall()
				m.statusMsg = "✓ agyhud uninstalled from Antigravity settings."
			} else {
				_, _ = installer.Install("", m.config.IconSet == config.IconSetClassic)
				m.statusMsg = "✓ agyhud integrated with Antigravity settings!"
			}
			m.hookStatus, _ = installer.GetStatus()
		case 1:
			if m.config.IconSet == config.IconSetNerdFont {
				m.config.IconSet = config.IconSetClassic
			} else {
				m.config.IconSet = config.IconSetNerdFont
			}
		case 2:
			if m.config.Git.RefreshSeconds < 10 {
				m.config.Git.RefreshSeconds++
			} else {
				m.config.Git.RefreshSeconds = 1
			}
		}

	case screenHUD:
		switch m.hudCursor {
		case 0:
			if m.config.Responsive.Mode == config.LayoutModeDynamic {
				m.config.Responsive.Mode = config.LayoutModeManual
				m.statusMsg = "Switched to Manual Fixed Rows mode."
			} else {
				m.config.Responsive.Mode = config.LayoutModeDynamic
				m.statusMsg = "Switched to Dynamic Auto-Wrap mode."
			}
		case 1:
			m.config.Responsive.FullWidth = !m.config.Responsive.FullWidth
		case 2:
			if m.config.Responsive.BreakpointWidth < 120 {
				m.config.Responsive.BreakpointWidth += 10
			} else {
				m.config.Responsive.BreakpointWidth = 60
			}
		}

	case screenWidgets:
		return m.startWidgetEditing()

	case screenWidgetCategories:
		m.catalogCursor = 0
		m.pushScreen(screenWidgetCatalog)

	case screenWidgetCatalog:
		if m.catalogCategoryCursor < 0 || m.catalogCategoryCursor >= len(widgets.CatalogCategories) {
			m.catalogCategoryCursor = 0
		}
		cat := widgets.CatalogCategories[m.catalogCategoryCursor]
		if m.catalogCursor < 0 || m.catalogCursor >= len(cat.Widgets) {
			m.catalogCursor = 0
		}
		selectedType := cat.Widgets[m.catalogCursor].Type
		targetRow := 0
		if len(m.config.Rows) > 1 && m.widgetsCursor >= len(m.config.Rows[0]) {
			targetRow = 1
		}
		if len(m.config.Rows) == 0 {
			m.config.Rows = append(m.config.Rows, []config.WidgetConfig{})
		}
		newWidget := config.WidgetConfig{
			Type:    selectedType,
			Enabled: true,
			Merge:   false,
			Label:   "", // Placeholder is computed dynamically when Label is empty
		}
		if selectedType == "custom_symbol" {
			newWidget.CustomSymbol = "•"
		} else if selectedType == "separator" {
			newWidget.Separator = "│"
		}
		m.config.Rows[targetRow] = append(m.config.Rows[targetRow], newWidget)
		m.popScreen() // exit screenWidgetCatalog
		m.popScreen() // exit screenWidgetCategories back to screenWidgets
		m.statusMsg = fmt.Sprintf("✓ Added '%s' to Row %d", selectedType, targetRow+1)

	case screenAppearance:
		widgetsList := m.getWidgetsList()
		if len(widgetsList) == 0 {
			return m, nil
		}
		if m.appearanceCursor < 0 || m.appearanceCursor >= len(widgetsList) {
			m.appearanceCursor = 0
		}
		ref := widgetsList[m.appearanceCursor]
		w := &m.config.Rows[ref.Row][ref.Col]

		m.mode = editInputText
		m.inputTargetField = "widget_color"
		m.textInput.SetValue(w.Color)
		m.textInput.Placeholder = "#7aa2f7, red, or empty for auto"
		m.textInput.Focus()
		return m, textinput.Blink
	}

	return m, nil
}

func (m *Model) startWidgetEditing() (tea.Model, tea.Cmd) {
	r, w := m.resolveWidgetIndices(m.widgetsCursor)
	if r < 0 || w < 0 {
		return m, nil
	}
	wCfg := &m.config.Rows[r][w]

	if wCfg.Type == "custom_symbol" {
		m.mode = editInputText
		m.inputTargetField = "widget_symbol"
		m.textInput.SetValue(wCfg.CustomSymbol)
		m.textInput.Placeholder = "Enter symbol glyph"
		m.textInput.Focus()
		return m, textinput.Blink
	} else if wCfg.Type == "separator" {
		m.mode = editInputText
		m.inputTargetField = "widget_symbol"
		m.textInput.SetValue(wCfg.Separator)
		m.textInput.Placeholder = "Enter separator (e.g. │, •, |)"
		m.textInput.Focus()
		return m, textinput.Blink
	} else if wCfg.RawValue {
		m.mode = editInputText
		m.inputTargetField = "widget_raw_prefix"
		m.textInput.Placeholder = "Enclose (e.g. [], (), <>, [ ], or empty to clear)"
		current := ""
		if wCfg.RawPrefix != "" || wCfg.RawSuffix != "" {
			current = wCfg.RawPrefix + wCfg.RawSuffix
		}
		m.textInput.SetValue(current)
		m.textInput.Focus()
		return m, textinput.Blink
	} else {
		m.mode = editInputText
		m.inputTargetField = "widget_label"
		m.textInput.SetValue(wCfg.Label)
		m.textInput.Placeholder = "Label prefix (leave empty for default placeholder)"
		m.textInput.Focus()
		return m, textinput.Blink
	}
}

func (m *Model) moveWidget(delta int) {
	r, w := m.resolveWidgetIndices(m.widgetsCursor)
	if r < 0 || w < 0 {
		return
	}

	target := w + delta
	if target >= 0 && target < len(m.config.Rows[r]) {
		m.config.Rows[r][w], m.config.Rows[r][target] = m.config.Rows[r][target], m.config.Rows[r][w]
		m.widgetsCursor = target
		m.statusMsg = fmt.Sprintf("↕ Moved '%s' to position %d", m.config.Rows[r][target].Type, target+1)
		return
	}

	if delta < 0 && r > 0 && w == 0 {
		item := m.config.Rows[r][w]
		m.config.Rows[r] = append(m.config.Rows[r][:w], m.config.Rows[r][w+1:]...)
		m.config.Rows[r-1] = append(m.config.Rows[r-1], item)
		m.widgetsCursor--
		m.statusMsg = fmt.Sprintf("↕ Moved '%s' to Row %d", item.Type, r)
	} else if delta > 0 && r < len(m.config.Rows)-1 && w == len(m.config.Rows[r])-1 {
		item := m.config.Rows[r][w]
		m.config.Rows[r] = append(m.config.Rows[r][:w], m.config.Rows[r][w+1:]...)
		m.config.Rows[r+1] = append([]config.WidgetConfig{item}, m.config.Rows[r+1]...)
		m.widgetsCursor++
		m.statusMsg = fmt.Sprintf("↕ Moved '%s' to Row %d", item.Type, r+2)
	}
}

func (m *Model) addSpacer() {
	r, w := m.resolveWidgetIndices(m.widgetsCursor)
	if r < 0 {
		if len(m.config.Rows) == 0 {
			m.config.Rows = append(m.config.Rows, []config.WidgetConfig{})
		}
		r = 0
		w = 0
	}
	spacer := config.WidgetConfig{
		Type:      "separator",
		Separator: "│",
		Enabled:   true,
	}

	if len(m.config.Rows[r]) == 0 || w+1 >= len(m.config.Rows[r]) {
		m.config.Rows[r] = append(m.config.Rows[r], spacer)
	} else {
		m.config.Rows[r] = append(m.config.Rows[r][:w+1], append([]config.WidgetConfig{spacer}, m.config.Rows[r][w+1:]...)...)
	}
	m.widgetsCursor++
	m.statusMsg = "✓ Added spacer separator '│'"
}

func (m *Model) updateModalInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			m.mode = editNone
			m.statusMsg = "Cancelled input."
			return m, nil

		case "enter":
			val := strings.TrimSpace(m.textInput.Value())

			switch m.inputTargetField {
			case "widget_color":
				widgetsList := m.getWidgetsList()
				if m.appearanceCursor >= 0 && m.appearanceCursor < len(widgetsList) {
					ref := widgetsList[m.appearanceCursor]
					if val != "" && !strings.HasPrefix(val, "#") && len(val) == 6 {
						val = "#" + val
					}
					m.config.Rows[ref.Row][ref.Col].Color = val
					if val != "" {
						m.statusMsg = fmt.Sprintf("✓ Set custom color %s for %s", val, m.config.Rows[ref.Row][ref.Col].Type)
					} else {
						m.statusMsg = fmt.Sprintf("✓ Reset %s to auto/default color", m.config.Rows[ref.Row][ref.Col].Type)
					}
				}
				m.mode = editNone

			case "widget_label":
				r, w := m.resolveWidgetIndices(m.widgetsCursor)
				if r >= 0 && w >= 0 {
					m.config.Rows[r][w].Label = val
					m.statusMsg = fmt.Sprintf("✓ Updated label for %s to '%s'", m.config.Rows[r][w].Type, val)
				}
				m.mode = editNone

			case "widget_symbol":
				r, w := m.resolveWidgetIndices(m.widgetsCursor)
				if r >= 0 && w >= 0 {
					if m.config.Rows[r][w].Type == "custom_symbol" {
						m.config.Rows[r][w].CustomSymbol = val
					} else {
						m.config.Rows[r][w].Separator = val
					}
					m.statusMsg = fmt.Sprintf("✓ Updated symbol to '%s'", val)
				}
				m.mode = editNone

			case "widget_raw_prefix":
				r, w := m.resolveWidgetIndices(m.widgetsCursor)
				if r >= 0 && w >= 0 {
					wCfg := &m.config.Rows[r][w]
					if val == "" {
						wCfg.RawPrefix = ""
						wCfg.RawSuffix = ""
						m.statusMsg = fmt.Sprintf("✓ Cleared enclosing characters for %s", wCfg.Type)
						m.mode = editNone
					} else if len([]rune(val)) == 2 && !strings.Contains(val, " ") {
						runes := []rune(val)
						wCfg.RawPrefix = string(runes[0])
						wCfg.RawSuffix = string(runes[1])
						m.statusMsg = fmt.Sprintf("✓ Enclosed %s with '%s' and '%s'", wCfg.Type, wCfg.RawPrefix, wCfg.RawSuffix)
						m.mode = editNone
					} else if strings.Contains(val, " ") {
						parts := strings.Fields(val)
						if len(parts) >= 2 {
							wCfg.RawPrefix = parts[0]
							wCfg.RawSuffix = parts[1]
							m.statusMsg = fmt.Sprintf("✓ Enclosed %s with '%s' and '%s'", wCfg.Type, wCfg.RawPrefix, wCfg.RawSuffix)
							m.mode = editNone
						} else {
							wCfg.RawPrefix = parts[0]
							m.inputTargetField = "widget_raw_suffix"
							m.textInput.Placeholder = "Closing character (e.g. ], ), >)"
							m.textInput.SetValue(matchClosingChar(wCfg.RawPrefix))
							return m, nil
						}
					} else {
						wCfg.RawPrefix = val
						m.inputTargetField = "widget_raw_suffix"
						m.textInput.Placeholder = "Closing character (e.g. ], ), >)"
						m.textInput.SetValue(matchClosingChar(wCfg.RawPrefix))
						return m, nil
					}
				} else {
					m.mode = editNone
				}

			case "widget_raw_suffix":
				r, w := m.resolveWidgetIndices(m.widgetsCursor)
				if r >= 0 && w >= 0 {
					wCfg := &m.config.Rows[r][w]
					wCfg.RawSuffix = val
					m.statusMsg = fmt.Sprintf("✓ Enclosed %s with '%s' and '%s'", wCfg.Type, wCfg.RawPrefix, val)
				}
				m.mode = editNone
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *Model) deleteCurrentWidget() {
	r, w := m.resolveWidgetIndices(m.widgetsCursor)
	if r >= 0 && w >= 0 {
		deletedType := m.config.Rows[r][w].Type
		m.config.Rows[r] = append(m.config.Rows[r][:w], m.config.Rows[r][w+1:]...)
		if m.widgetsCursor > 0 {
			m.widgetsCursor--
		}
		m.statusMsg = fmt.Sprintf("✓ Removed %s from Row %d", deletedType, r+1)
	}
}

func (m *Model) saveConfig() {
	if err := config.Save(m.configPath, m.config); err != nil {
		m.statusMsg = fmt.Sprintf("Error saving config: %v", err)
	} else {
		m.statusMsg = "✓ Configuration saved successfully!"
	}
}

func matchClosingChar(open string) string {
	switch open {
	case "[":
		return "]"
	case "(":
		return ")"
	case "<":
		return ">"
	case "{":
		return "}"
	case "«":
		return "»"
	case "|":
		return "|"
	default:
		return ""
	}
}
