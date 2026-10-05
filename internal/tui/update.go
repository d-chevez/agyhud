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
			if !m.popScreen() {
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil

		case "up", "k":
			m.moveCursor(-1)

		case "down", "j":
			m.moveCursor(1)

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

		case "enter":
			return m.handleEnterSelection()

		case " ":
			return m.handleSpaceToggle()

		case "a":
			if m.currentScreen() == screenWidgets {
				m.catalogCursor = 0
				m.pushScreen(screenWidgetCatalog)
				return m, nil
			}

		case "p", "+":
			if m.currentScreen() == screenWidgets {
				m.addSpacer()
				return m, nil
			}

		case "d", "delete":
			if m.currentScreen() == screenWidgets {
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

		case "R":
			if m.currentScreen() == screenWidgets {
				m.config.Rows = append(m.config.Rows, []config.WidgetConfig{})
				m.statusMsg = fmt.Sprintf("✓ Created Row %d (Press 'a' to add widgets)", len(m.config.Rows))
			}

		case "s", "ctrl+s":
			m.saveConfig()
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
	case screenWidgetCatalog:
		m.catalogCursor = clamp(m.catalogCursor+delta, 0, len(widgets.AvailableWidgetTypes)-1)
	case screenAppearance:
		widgetsCount := len(m.getWidgetsList())
		max := 2 + widgetsCount + len(m.colorFields) - 1
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

func (m *Model) handleSpaceToggle() (tea.Model, tea.Cmd) {
	if m.currentScreen() == screenWidgets {
		r, w := m.resolveWidgetIndices(m.widgetsCursor)
		if r >= 0 && w >= 0 {
			m.config.Rows[r][w].Enabled = !m.config.Rows[r][w].Enabled
		}
		return m, nil
	}
	return m.handleEnterSelection()
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

	case screenWidgetCatalog:
		selectedType := widgets.AvailableWidgetTypes[m.catalogCursor]
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
		}
		if selectedType == "custom_symbol" {
			newWidget.CustomSymbol = "•"
		} else if selectedType == "separator" {
			newWidget.Separator = "│"
		}
		m.config.Rows[targetRow] = append(m.config.Rows[targetRow], newWidget)
		m.statusMsg = fmt.Sprintf("✓ Added '%s' to Row %d", selectedType, targetRow+1)
		m.popScreen()

	case screenAppearance:
		widgetsList := m.getWidgetsList()
		widgetsCount := len(widgetsList)

		if m.appearanceCursor == 0 {
			// Select Default Antigravity Dark Theme
			m.config.ThemeMode = "default"
			m.config.Theme = config.AntigravityDarkTheme
			// Reset custom colors on widgets so clean theme palette takes over
			for r := range m.config.Rows {
				for c := range m.config.Rows[r] {
					m.config.Rows[r][c].Color = ""
				}
			}
			m.statusMsg = "✓ Antigravity Dark Theme active (reset widget colors to default)"
		} else if m.appearanceCursor == 1 {
			// Switch to Custom Theme mode
			m.config.ThemeMode = "custom"
			m.statusMsg = "✓ Custom Theme mode active"
		} else if m.appearanceCursor < 2+widgetsCount {
			// Selected a specific widget to customize its color
			widgetIdx := m.appearanceCursor - 2
			ref := widgetsList[widgetIdx]
			w := &m.config.Rows[ref.Row][ref.Col]

			m.mode = editInputText
			m.inputTargetField = "widget_color"
			m.textInput.SetValue(w.Color)
			m.textInput.Placeholder = "#7aa2f7 or leave empty for default"
			m.textInput.Focus()
			return m, textinput.Blink
		} else {
			// Selected a global palette field
			colorIdx := m.appearanceCursor - (2 + widgetsCount)
			if colorIdx >= 0 && colorIdx < len(m.colorFields) {
				m.mode = editInputText
				m.inputTargetField = "global_color"
				currentVal := m.colorFields[colorIdx].get(&m.config.Theme)
				m.textInput.SetValue(currentVal)
				m.textInput.Placeholder = "#7aa2f7 or color name"
				m.textInput.Focus()
				return m, textinput.Blink
			}
		}
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
	} else {
		m.mode = editInputText
		m.inputTargetField = "widget_label"
		m.textInput.SetValue(wCfg.Label)
		m.textInput.Placeholder = "Label prefix (leave empty for none)"
		m.textInput.Focus()
		return m, textinput.Blink
	}
}

func (m *Model) moveWidget(delta int) {
	r, w := m.resolveWidgetIndices(m.widgetsCursor)
	if r < 0 || w < 0 {
		return
	}

	if delta < 0 {
		// Move UP / LEFT
		if w > 0 {
			m.config.Rows[r][w], m.config.Rows[r][w-1] = m.config.Rows[r][w-1], m.config.Rows[r][w]
			m.widgetsCursor--
			m.statusMsg = fmt.Sprintf("✓ Moved %s up", m.config.Rows[r][w-1].Type)
		} else if r > 0 {
			item := m.config.Rows[r][w]
			m.config.Rows[r] = append(m.config.Rows[r][:w], m.config.Rows[r][w+1:]...)
			m.config.Rows[r-1] = append(m.config.Rows[r-1], item)
			m.widgetsCursor--
			m.statusMsg = fmt.Sprintf("✓ Moved %s to Row %d", item.Type, r)
		}
	} else if delta > 0 {
		// Move DOWN / RIGHT
		if w < len(m.config.Rows[r])-1 {
			m.config.Rows[r][w], m.config.Rows[r][w+1] = m.config.Rows[r][w+1], m.config.Rows[r][w]
			m.widgetsCursor++
			m.statusMsg = fmt.Sprintf("✓ Moved %s down", m.config.Rows[r][w+1].Type)
		} else if r < len(m.config.Rows)-1 {
			item := m.config.Rows[r][w]
			m.config.Rows[r] = append(m.config.Rows[r][:w], m.config.Rows[r][w+1:]...)
			m.config.Rows[r+1] = append([]config.WidgetConfig{item}, m.config.Rows[r+1]...)
			m.widgetsCursor++
			m.statusMsg = fmt.Sprintf("✓ Moved %s to Row %d", item.Type, r+2)
		}
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
				widgetIdx := m.appearanceCursor - 2
				if widgetIdx >= 0 && widgetIdx < len(widgetsList) {
					ref := widgetsList[widgetIdx]
					if val != "" && !strings.HasPrefix(val, "#") && len(val) == 6 {
						val = "#" + val
					}
					m.config.Rows[ref.Row][ref.Col].Color = val
					m.config.ThemeMode = "custom" // Deselects default theme!
					if val != "" {
						m.statusMsg = fmt.Sprintf("✓ Set custom color %s for %s (Switched to Custom Theme)", val, m.config.Rows[ref.Row][ref.Col].Type)
					} else {
						m.statusMsg = fmt.Sprintf("✓ Reset %s to theme default", m.config.Rows[ref.Row][ref.Col].Type)
					}
				}
				m.mode = editNone

			case "global_color":
				if val != "" && !strings.HasPrefix(val, "#") && len(val) == 6 {
					val = "#" + val
				}
				widgetsCount := len(m.getWidgetsList())
				idx := m.appearanceCursor - (2 + widgetsCount)
				if idx >= 0 && idx < len(m.colorFields) {
					m.colorFields[idx].set(&m.config.Theme, val)
					m.config.ThemeMode = "custom" // Deselects default theme!
					m.statusMsg = fmt.Sprintf("✓ Updated %s to %s (Switched to Custom Theme)", m.colorFields[idx].label, val)
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
