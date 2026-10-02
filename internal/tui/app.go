package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/d-chevez/agyhud/internal/config"
	"github.com/d-chevez/agyhud/internal/engine"
	"github.com/d-chevez/agyhud/internal/installer"
	"github.com/d-chevez/agyhud/internal/payload"
)

type tabIndex int

const (
	tabGeneral tabIndex = iota
	tabWidgets
	tabTheme
)

type colorField struct {
	label string
	get   func(cfg *config.ThemeConfig) string
	set   func(cfg *config.ThemeConfig, val string)
}

// Model is the main Bubbletea state model for the agyhud configuration TUI.
type Model struct {
	config       *config.Config
	configPath   string
	payload      *payload.SessionPayload
	activeTab    tabIndex
	cursor       int
	statusMsg    string
	width        int
	height       int
	hookStatus   installer.HookStatus
	quitting     bool
	editingColor bool
	textInput    textinput.Model
	colorFields  []colorField
}

// InitialModel prepares the TUI model.
func InitialModel(cfgPath string) (*Model, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	hStatus, _ := installer.GetStatus()

	ti := textinput.New()
	ti.CharLimit = 15

	fields := []colorField{
		{"Accent", func(c *config.ThemeConfig) string { return c.Accent }, func(c *config.ThemeConfig, v string) { c.Accent = v }},
		{"Dim", func(c *config.ThemeConfig) string { return c.Dim }, func(c *config.ThemeConfig, v string) { c.Dim = v }},
		{"Text", func(c *config.ThemeConfig) string { return c.Text }, func(c *config.ThemeConfig, v string) { c.Text = v }},
		{"Success", func(c *config.ThemeConfig) string { return c.Success }, func(c *config.ThemeConfig, v string) { c.Success = v }},
		{"Warning", func(c *config.ThemeConfig) string { return c.Warning }, func(c *config.ThemeConfig, v string) { c.Warning = v }},
		{"Danger", func(c *config.ThemeConfig) string { return c.Danger }, func(c *config.ThemeConfig, v string) { c.Danger = v }},
		{"BarFilled", func(c *config.ThemeConfig) string { return c.BarFilled }, func(c *config.ThemeConfig, v string) { c.BarFilled = v }},
		{"BarEmpty", func(c *config.ThemeConfig) string { return c.BarEmpty }, func(c *config.ThemeConfig, v string) { c.BarEmpty = v }},
	}

	return &Model{
		config:      cfg,
		configPath:  cfgPath,
		payload:     GetSamplePayload(),
		activeTab:   tabGeneral,
		cursor:      0,
		hookStatus:  hStatus,
		textInput:   ti,
		colorFields: fields,
	}, nil
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.editingColor {
		return m.updateColorInput(msg)
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
			m.activeTab = (m.activeTab + 1) % 3
			m.cursor = 0
			m.statusMsg = ""

		case "shift+tab":
			if m.activeTab == 0 {
				m.activeTab = tabTheme
			} else {
				m.activeTab--
			}
			m.cursor = 0
			m.statusMsg = ""

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

		case "s", "ctrl+s":
			if err := config.Save(m.configPath, m.config); err != nil {
				m.statusMsg = fmt.Sprintf("Error saving config: %v", err)
			} else {
				m.statusMsg = "✓ Configuration saved successfully!"
			}

		case "enter", " ":
			return m.handleSelection()

		case "+", "=":
			if m.activeTab == tabGeneral && m.cursor == 2 {
				if m.config.Git.RefreshSeconds < 10 {
					m.config.Git.RefreshSeconds++
				}
			}

		case "-", "_":
			if m.activeTab == tabGeneral && m.cursor == 2 {
				if m.config.Git.RefreshSeconds > 1 {
					m.config.Git.RefreshSeconds--
				}
			}
		}
	}

	return m, nil
}

func (m *Model) updateColorInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			m.editingColor = false
			m.statusMsg = "Cancelled color edit."
			return m, nil
		case "enter":
			val := strings.TrimSpace(m.textInput.Value())
			if val != "" {
				if !strings.HasPrefix(val, "#") && len(val) == 6 {
					val = "#" + val
				}
				idx := m.cursor - 5 // Presets offset
				if idx >= 0 && idx < len(m.colorFields) {
					m.colorFields[idx].set(&m.config.Theme, val)
					m.statusMsg = fmt.Sprintf("✓ Updated %s to %s", m.colorFields[idx].label, val)
				}
			}
			m.editingColor = false
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *Model) handleHorizontalAdjust(delta int) {
	if m.activeTab == tabWidgets {
		idx := 0
		for r := range m.config.Rows {
			for w := range m.config.Rows[r] {
				if idx == m.cursor {
					newPad := m.config.Rows[r][w].Padding + delta
					if newPad >= 0 && newPad <= 5 {
						m.config.Rows[r][w].Padding = newPad
					}
					return
				}
				idx++
			}
		}
	}
}

func (m *Model) getMaxCursorForTab() int {
	switch m.activeTab {
	case tabGeneral:
		return 3
	case tabWidgets:
		count := 0
		for _, row := range m.config.Rows {
			count += len(row)
		}
		if count > 0 {
			return count - 1
		}
		return 0
	case tabTheme:
		return 5 + len(m.colorFields) - 1 // 5 presets + 8 color slots
	}
	return 0
}

func (m *Model) handleSelection() (tea.Model, tea.Cmd) {
	switch m.activeTab {
	case tabGeneral:
		switch m.cursor {
		case 0:
			if m.config.IconSet == config.IconSetNerdFont {
				m.config.IconSet = config.IconSetClassic
			} else {
				m.config.IconSet = config.IconSetNerdFont
			}
		case 1:
			m.config.Responsive.Enabled = !m.config.Responsive.Enabled
		case 2:
			if m.config.Git.RefreshSeconds < 10 {
				m.config.Git.RefreshSeconds++
			} else {
				m.config.Git.RefreshSeconds = 1
			}
		case 3:
			if m.hookStatus.Active && m.hookStatus.IsAgyhud {
				_ = installer.Uninstall()
				m.statusMsg = "✓ agyhud uninstalled from Antigravity settings."
			} else {
				_, _ = installer.Install("", m.config.IconSet == config.IconSetClassic)
				m.statusMsg = "✓ agyhud integrated with Antigravity settings!"
			}
			m.hookStatus, _ = installer.GetStatus()
		}

	case tabWidgets:
		idx := 0
		for r := range m.config.Rows {
			for w := range m.config.Rows[r] {
				if idx == m.cursor {
					m.config.Rows[r][w].Enabled = !m.config.Rows[r][w].Enabled
					return m, nil
				}
				idx++
			}
		}

	case tabTheme:
		presets := []struct {
			name  string
			theme config.ThemeConfig
		}{
			{"Tokyo Night", config.ThemeConfig{Accent: "#7aa2f7", Dim: "#565f89", Text: "#c0caf5", Success: "#9ece6a", Warning: "#e0af68", Danger: "#f7768e", BarFilled: "#7aa2f7", BarEmpty: "#24283b"}},
			{"Catppuccin Mocha", config.ThemeConfig{Accent: "#89b4fa", Dim: "#6c7086", Text: "#cdd6f4", Success: "#a6e3a1", Warning: "#f9e2af", Danger: "#f38ba8", BarFilled: "#89b4fa", BarEmpty: "#313244"}},
			{"Gruvbox", config.ThemeConfig{Accent: "#83a598", Dim: "#928374", Text: "#ebdbb2", Success: "#b8bb26", Warning: "#fabd2f", Danger: "#fb4934", BarFilled: "#83a598", BarEmpty: "#3c3836"}},
			{"Nord", config.ThemeConfig{Accent: "#88c0d0", Dim: "#4c566a", Text: "#eceff4", Success: "#a3be8c", Warning: "#ebcb8b", Danger: "#bf616a", BarFilled: "#88c0d0", BarEmpty: "#3b4252"}},
			{"Cyberpunk", config.ThemeConfig{Accent: "#00ffff", Dim: "#711c91", Text: "#f8f8f2", Success: "#05ffa1", Warning: "#ffe600", Danger: "#ff0055", BarFilled: "#00ffff", BarEmpty: "#130924"}},
		}

		if m.cursor < len(presets) {
			m.config.Theme = presets[m.cursor].theme
			m.statusMsg = fmt.Sprintf("Applied theme: %s", presets[m.cursor].name)
		} else {
			// Edit individual color
			colorIdx := m.cursor - len(presets)
			if colorIdx >= 0 && colorIdx < len(m.colorFields) {
				m.editingColor = true
				currentVal := m.colorFields[colorIdx].get(&m.config.Theme)
				m.textInput.SetValue(currentVal)
				m.textInput.Focus()
				return m, textinput.Blink
			}
		}
	}
	return m, nil
}

func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Accent))
	b.WriteString(titleStyle.Render(" agyhud — Interactive Configuration & HUD Studio") + "\n\n")

	// 1. Live Preview Section
	previewContent := engine.Render(m.payload, m.config)
	previewBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.config.Theme.Accent)).
		Padding(0, 1).
		Render(previewContent)

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim)).Render("── LIVE PREVIEW ") + "\n")
	b.WriteString(previewBox + "\n\n")

	// 2. Navigation Tabs
	b.WriteString(m.renderTabs() + "\n\n")

	// 3. Tab Body
	switch m.activeTab {
	case tabGeneral:
		b.WriteString(m.renderGeneralTab())
	case tabWidgets:
		b.WriteString(m.renderWidgetsTab())
	case tabTheme:
		b.WriteString(m.renderThemeTab())
	}

	// 4. Color Editor Modal/Prompt if active
	if m.editingColor {
		colorIdx := m.cursor - 5
		fieldName := m.colorFields[colorIdx].label
		promptStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.config.Theme.Warning))
		b.WriteString("\n" + promptStyle.Render(fmt.Sprintf(" Edit Hex Color for [%s]: ", fieldName)) + m.textInput.View() + " (Enter to save, Esc to cancel)\n")
	} else if m.statusMsg != "" {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Success)).Bold(true)
		b.WriteString("\n" + statusStyle.Render(m.statusMsg) + "\n")
	} else {
		b.WriteString("\n\n")
	}

	// 5. Footer & Keybindings
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Dim))
	if m.activeTab == tabWidgets {
		b.WriteString(footerStyle.Render("[Tab] Switch Tab │ [↑/↓] Navigate │ [Space] Toggle │ [←/→] Padding │ [s] Save │ [q] Exit"))
	} else if m.activeTab == tabTheme {
		b.WriteString(footerStyle.Render("[Tab] Switch Tab │ [↑/↓] Navigate │ [Enter] Load/Edit Color │ [s] Save │ [q] Exit"))
	} else {
		b.WriteString(footerStyle.Render("[Tab] Switch Tab │ [↑/↓] Navigate │ [Space/Enter] Toggle/Apply │ [s] Save │ [q] Exit"))
	}

	return b.String()
}

func (m *Model) renderTabs() string {
	tabs := []string{"General & Setup", "Rows & Widgets", "Themes & Colors"}
	var rendered []string

	for i, t := range tabs {
		if tabIndex(i) == m.activeTab {
			active := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(m.config.Theme.Accent)).
				Background(lipgloss.Color("#24283b")).
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

func (m *Model) renderGeneralTab() string {
	var items []string

	iconSetStr := "Nerd Font (Modern devicons)"
	if m.config.IconSet == config.IconSetClassic {
		iconSetStr = "Classic (Universal ASCII/Unicode)"
	}
	items = append(items, fmt.Sprintf("Font Glyphs:         [%s]", iconSetStr))

	respStr := "Disabled"
	if m.config.Responsive.Enabled {
		respStr = fmt.Sprintf("Enabled (Breakpoint: < %d cols)", m.config.Responsive.BreakpointWidth)
	}
	items = append(items, fmt.Sprintf("Adaptive Responsive: [%s]", respStr))
	items = append(items, fmt.Sprintf("Git Cache Window:    [%d seconds] (+/- to adjust)", m.config.Git.RefreshSeconds))

	hookStr := "Not Configured (Press Enter to Activate)"
	if m.hookStatus.Active && m.hookStatus.IsAgyhud {
		hookStr = "Active in Antigravity CLI (Press Enter to Disable)"
	}
	items = append(items, fmt.Sprintf("Antigravity Hook:    [%s]", hookStr))

	return m.renderList(items)
}

func (m *Model) renderWidgetsTab() string {
	var items []string
	for r, row := range m.config.Rows {
		for _, w := range row {
			status := "[ ]"
			if w.Enabled {
				status = "[x]"
			}
			items = append(items, fmt.Sprintf("Row %d │ %s %-12s (Padding: %d │ ←/→ to adjust)", r+1, status, w.Type, w.Padding))
		}
	}
	return m.renderList(items)
}

func (m *Model) renderThemeTab() string {
	var items []string
	presets := []string{
		"Tokyo Night (Default Modern)",
		"Catppuccin Mocha (Pastel Clean)",
		"Gruvbox (Warm Vintage)",
		"Nord (Arctic Minimal)",
		"Cyberpunk (Neon Glow)",
	}

	items = append(items, "── PRESET PALETTES (Press Enter to Load) ──")
	for _, p := range presets {
		items = append(items, fmt.Sprintf("Preset: %s", p))
	}

	items = append(items, "── GRANULAR COLOR CUSTOMIZATION (Press Enter to Edit) ──")
	for _, f := range m.colorFields {
		val := f.get(&m.config.Theme)
		swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(val)).Render("■■")
		items = append(items, fmt.Sprintf("Color %-10s %-9s %s", f.label+":", val, swatch))
	}

	return m.renderList(items)
}

func (m *Model) renderList(items []string) string {
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
		if itemIndex == m.cursor {
			rendered = append(rendered, selStyle.Render(" ▶ "+item))
		} else {
			rendered = append(rendered, normStyle.Render("   "+item))
		}
	}

	return strings.Join(rendered, "\n")
}

// Run launches the interactive configuration TUI.
func Run(cfgPath string) error {
	m, err := InitialModel(cfgPath)
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
