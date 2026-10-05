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
	if m.mode == editInputText || m.mode == editAddWidget {
		return m.updateModalInput(msg)
	}

	if m.mode == editInspector {
		return m.updateInspectorInput(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.payload.TerminalWidth = msg.Width

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "tab":
			m.activeTab = (m.activeTab + 1) % 4
			m.cursor = 0
			m.statusMsg = ""

		case "shift+tab":
			if m.activeTab == 0 {
				m.activeTab = tabAppearance
			} else {
				m.activeTab--
			}
			m.cursor = 0
			m.statusMsg = ""

		case "1":
			m.activeTab = tabTerminal
			m.cursor = 0
		case "2":
			m.activeTab = tabHUD
			m.cursor = 0
		case "3":
			m.activeTab = tabWidgets
			m.cursor = 0
		case "4":
			m.activeTab = tabAppearance
			m.cursor = 0

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			maxCursor := m.getMaxCursorForTab()
			if m.cursor < maxCursor {
				m.cursor++
			}

		case "left", "h":
			m.handleHorizontalAdjust(-1)

		case "right", "l":
			m.handleHorizontalAdjust(1)

		case "enter":
			return m.handleEnterSelection()

		case " ":
			return m.handleSpaceToggle()

		case "a":
			if m.activeTab == tabWidgets {
				m.mode = editAddWidget
				m.catalogIndex = 0
				return m, nil
			}

		case "d", "delete":
			if m.activeTab == tabWidgets {
				m.deleteCurrentWidget()
			}

		case "R":
			if m.activeTab == tabWidgets {
				m.config.Rows = append(m.config.Rows, []config.WidgetConfig{})
				m.statusMsg = fmt.Sprintf("✓ Created Row %d (Press 'a' to add widgets to it)", len(m.config.Rows))
			}

		case "s", "ctrl+s":
			if err := config.Save(m.configPath, m.config); err != nil {
				m.statusMsg = fmt.Sprintf("Error saving config: %v", err)
			} else {
				m.statusMsg = "✓ Configuration saved successfully!"
			}
		}
	}

	return m, nil
}

func (m *Model) getMaxCursorForTab() int {
	switch m.activeTab {
	case tabTerminal:
		return 2 // 3 options (0..2)
	case tabHUD:
		return 2 // 3 options (0..2)
	case tabWidgets:
		count := m.getTotalWidgetsCount()
		if count > 0 {
			return count - 1
		}
		return 0
	case tabAppearance:
		return len(themePresets) + len(m.colorFields) - 1
	}
	return 0
}

func (m *Model) handleHorizontalAdjust(delta int) {
	switch m.activeTab {
	case tabTerminal:
		if m.cursor == 2 { // Git cache interval
			newVal := m.config.Git.RefreshSeconds + delta
			if newVal >= 1 && newVal <= 10 {
				m.config.Git.RefreshSeconds = newVal
			}
		}
	case tabHUD:
		if m.cursor == 2 { // Breakpoint width
			newBP := m.config.Responsive.BreakpointWidth + (delta * 5)
			if newBP >= 40 && newBP <= 200 {
				m.config.Responsive.BreakpointWidth = newBP
			}
		}
	}
}

func (m *Model) handleSpaceToggle() (tea.Model, tea.Cmd) {
	if m.activeTab == tabWidgets {
		r, w := m.resolveWidgetIndices(m.cursor)
		if r >= 0 && w >= 0 {
			m.config.Rows[r][w].Enabled = !m.config.Rows[r][w].Enabled
		}
	} else {
		return m.handleEnterSelection()
	}
	return m, nil
}

func (m *Model) handleEnterSelection() (tea.Model, tea.Cmd) {
	switch m.activeTab {
	case tabTerminal:
		switch m.cursor {
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

	case tabHUD:
		switch m.cursor {
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

	case tabWidgets:
		// Open the Widget Inspector modal for the selected widget!
		r, w := m.resolveWidgetIndices(m.cursor)
		if r >= 0 && w >= 0 {
			m.mode = editInspector
			m.inspectorCursor = 0
		}

	case tabAppearance:
		if m.cursor < len(themePresets) {
			m.config.Theme = themePresets[m.cursor].theme
			m.statusMsg = fmt.Sprintf("✓ Applied preset theme: %s", themePresets[m.cursor].name)
		} else {
			colorIdx := m.cursor - len(themePresets)
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

func (m *Model) updateInspectorInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			m.mode = editNone
			return m, nil
		case "up", "k":
			if m.inspectorCursor > 0 {
				m.inspectorCursor--
			}
		case "down", "j":
			maxInspect := 4
			r, w := m.resolveWidgetIndices(m.cursor)
			if r >= 0 && w >= 0 && (m.config.Rows[r][w].Type == "custom_symbol" || m.config.Rows[r][w].Type == "separator") {
				maxInspect = 5
			}
			if m.inspectorCursor < maxInspect {
				m.inspectorCursor++
			}
		case "d", "delete":
			m.deleteCurrentWidget()
			m.mode = editNone
			return m, nil
		case "enter", " ":
			r, w := m.resolveWidgetIndices(m.cursor)
			if r < 0 || w < 0 {
				m.mode = editNone
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
		}
	}
	return m, nil
}

func (m *Model) updateModalInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			if m.mode == editInputText && (m.inputTargetField == "widget_label" || m.inputTargetField == "widget_color" || m.inputTargetField == "widget_symbol") {
				m.mode = editInspector
			} else {
				m.mode = editNone
			}
			m.statusMsg = "Cancelled input."
			return m, nil

		case "up", "k":
			if m.mode == editAddWidget {
				if m.catalogIndex > 0 {
					m.catalogIndex--
				}
				return m, nil
			}

		case "down", "j":
			if m.mode == editAddWidget {
				if m.catalogIndex < len(widgets.AvailableWidgetTypes)-1 {
					m.catalogIndex++
				}
				return m, nil
			}

		case "enter":
			val := strings.TrimSpace(m.textInput.Value())

			switch m.mode {
			case editInputText:
				switch m.inputTargetField {
				case "global_color":
					if val != "" && !strings.HasPrefix(val, "#") && len(val) == 6 {
						val = "#" + val
					}
					idx := m.cursor - len(themePresets)
					if idx >= 0 && idx < len(m.colorFields) {
						m.colorFields[idx].set(&m.config.Theme, val)
						m.statusMsg = fmt.Sprintf("✓ Updated %s color to %s", m.colorFields[idx].label, val)
					}
					m.mode = editNone

				case "widget_label":
					r, w := m.resolveWidgetIndices(m.cursor)
					if r >= 0 && w >= 0 {
						m.config.Rows[r][w].Label = val
						m.statusMsg = fmt.Sprintf("✓ Updated label for %s to '%s'", m.config.Rows[r][w].Type, val)
					}
					m.mode = editInspector

				case "widget_color":
					r, w := m.resolveWidgetIndices(m.cursor)
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
					m.mode = editInspector

				case "widget_symbol":
					r, w := m.resolveWidgetIndices(m.cursor)
					if r >= 0 && w >= 0 {
						if m.config.Rows[r][w].Type == "custom_symbol" {
							m.config.Rows[r][w].CustomSymbol = val
						} else {
							m.config.Rows[r][w].Separator = val
						}
						m.statusMsg = fmt.Sprintf("✓ Updated symbol to '%s'", val)
					}
					m.mode = editInspector
				}
				return m, nil

			case editAddWidget:
				selectedType := widgets.AvailableWidgetTypes[m.catalogIndex]
				targetRow := 0
				if len(m.config.Rows) > 1 && m.cursor >= len(m.config.Rows[0]) {
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
				m.mode = editNone
				return m, nil
			}
		}
	}

	if m.mode == editInputText {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *Model) deleteCurrentWidget() {
	r, w := m.resolveWidgetIndices(m.cursor)
	if r >= 0 && w >= 0 {
		deletedType := m.config.Rows[r][w].Type
		m.config.Rows[r] = append(m.config.Rows[r][:w], m.config.Rows[r][w+1:]...)
		if m.cursor > 0 {
			m.cursor--
		}
		m.statusMsg = fmt.Sprintf("✓ Removed %s from Row %d", deletedType, r+1)
	}
}
