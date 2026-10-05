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

		case "d", "delete":
			if m.currentScreen() == screenWidgets {
				m.deleteCurrentWidget()
			} else if m.currentScreen() == screenWidgetInspector {
				m.deleteCurrentWidget()
				m.popScreen()
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
	case screenWidgetInspector:
		maxInspect := 4
		r, w := m.resolveWidgetIndices(m.widgetsCursor)
		if r >= 0 && w >= 0 && (m.config.Rows[r][w].Type == "custom_symbol" || m.config.Rows[r][w].Type == "separator") {
			maxInspect = 5
		}
		m.inspectorCursor = clamp(m.inspectorCursor+delta, 0, maxInspect)
	case screenWidgetCatalog:
		m.catalogCursor = clamp(m.catalogCursor+delta, 0, len(widgets.AvailableWidgetTypes)-1)
	case screenAppearance:
		max := len(themePresets) + len(m.colorFields) - 1
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
		if m.terminalCursor == 2 { // Git cache interval
			newVal := m.config.Git.RefreshSeconds + delta
			if newVal >= 1 && newVal <= 10 {
				m.config.Git.RefreshSeconds = newVal
			}
		}
	case screenHUD:
		if m.hudCursor == 2 { // Breakpoint width
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
		case 0: // Terminal & Integration
			m.pushScreen(screenTerminal)
		case 1: // HUD Layout & Flow
			m.pushScreen(screenHUD)
		case 2: // Widgets Lego Builder
			m.pushScreen(screenWidgets)
		case 3: // Appearance & Themes
			m.pushScreen(screenAppearance)
		case 4: // Save Configuration
			m.saveConfig()
		case 5: // Exit
			m.quitting = true
			return m, tea.Quit
		}

	case screenTerminal:
		switch m.terminalCursor {
		case 0: // Toggle Antigravity hook
			if m.hookStatus.Active && m.hookStatus.IsAgyhud {
				_ = installer.Uninstall()
				m.statusMsg = "✓ agyhud uninstalled from Antigravity settings."
			} else {
				_, _ = installer.Install("", m.config.IconSet == config.IconSetClassic)
				m.statusMsg = "✓ agyhud integrated with Antigravity settings!"
			}
			m.hookStatus, _ = installer.GetStatus()
		case 1: // Toggle Font Glyphs
			if m.config.IconSet == config.IconSetNerdFont {
				m.config.IconSet = config.IconSetClassic
			} else {
				m.config.IconSet = config.IconSetNerdFont
			}
		case 2: // Git cache cycle
			if m.config.Git.RefreshSeconds < 10 {
				m.config.Git.RefreshSeconds++
			} else {
				m.config.Git.RefreshSeconds = 1
			}
		}

	case screenHUD:
		switch m.hudCursor {
		case 0: // Toggle Layout Mode
			if m.config.Responsive.Mode == config.LayoutModeDynamic {
				m.config.Responsive.Mode = config.LayoutModeManual
				m.statusMsg = "Switched to Manual Fixed Rows mode."
			} else {
				m.config.Responsive.Mode = config.LayoutModeDynamic
				m.statusMsg = "Switched to Dynamic Auto-Wrap mode."
			}
		case 1: // Toggle Full-Width
			m.config.Responsive.FullWidth = !m.config.Responsive.FullWidth
		case 2: // Cycle Breakpoint
			if m.config.Responsive.BreakpointWidth < 120 {
				m.config.Responsive.BreakpointWidth += 10
			} else {
				m.config.Responsive.BreakpointWidth = 60
			}
		}

	case screenWidgets:
		r, w := m.resolveWidgetIndices(m.widgetsCursor)
		if r >= 0 && w >= 0 {
			m.inspectorCursor = 0
			m.pushScreen(screenWidgetInspector)
		}

	case screenWidgetInspector:
		r, w := m.resolveWidgetIndices(m.widgetsCursor)
		if r < 0 || w < 0 {
			m.popScreen()
			return m, nil
		}
		wCfg := &m.config.Rows[r][w]

		switch m.inspectorCursor {
		case 0: // Toggle Status
			wCfg.Enabled = !wCfg.Enabled
		case 1: // Edit Label
			m.mode = editInputText
			m.inputTargetField = "widget_label"
			m.textInput.SetValue(wCfg.Label)
			m.textInput.Placeholder = "Label prefix (leave empty for none)"
			m.textInput.Focus()
			return m, textinput.Blink
		case 2: // Toggle RawValue
			wCfg.RawValue = !wCfg.RawValue
		case 3: // Toggle Merge
			wCfg.Merge = !wCfg.Merge
		case 4: // Edit Color
			m.mode = editInputText
			m.inputTargetField = "widget_color"
			m.textInput.SetValue(wCfg.Color)
			m.textInput.Placeholder = "#7aa2f7 or leave empty for default"
			m.textInput.Focus()
			return m, textinput.Blink
		case 5: // Edit Symbol or Separator
			m.mode = editInputText
			m.inputTargetField = "widget_symbol"
			if wCfg.Type == "custom_symbol" {
				m.textInput.SetValue(wCfg.CustomSymbol)
			} else {
				m.textInput.SetValue(wCfg.Separator)
			}
			m.textInput.Focus()
			return m, textinput.Blink
		}

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
		if m.appearanceCursor < len(themePresets) {
			m.config.Theme = themePresets[m.appearanceCursor].theme
			m.statusMsg = fmt.Sprintf("✓ Applied preset theme: %s", themePresets[m.appearanceCursor].name)
		} else {
			colorIdx := m.appearanceCursor - len(themePresets)
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
			case "global_color":
				if val != "" && !strings.HasPrefix(val, "#") && len(val) == 6 {
					val = "#" + val
				}
				idx := m.appearanceCursor - len(themePresets)
				if idx >= 0 && idx < len(m.colorFields) {
					m.colorFields[idx].set(&m.config.Theme, val)
					m.statusMsg = fmt.Sprintf("✓ Updated %s color to %s", m.colorFields[idx].label, val)
				}
				m.mode = editNone

			case "widget_label":
				r, w := m.resolveWidgetIndices(m.widgetsCursor)
				if r >= 0 && w >= 0 {
					m.config.Rows[r][w].Label = val
					m.statusMsg = fmt.Sprintf("✓ Updated label for %s to '%s'", m.config.Rows[r][w].Type, val)
				}
				m.mode = editNone

			case "widget_color":
				r, w := m.resolveWidgetIndices(m.widgetsCursor)
				if r >= 0 && w >= 0 {
					if val != "" && !strings.HasPrefix(val, "#") && len(val) == 6 {
						val = "#" + val
					}
					m.config.Rows[r][w].Color = val
					if val != "" {
						m.statusMsg = fmt.Sprintf("✓ Set custom color for %s to %s", m.config.Rows[r][w].Type, val)
					} else {
						m.statusMsg = fmt.Sprintf("✓ Reset %s to theme default color", m.config.Rows[r][w].Type)
					}
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
