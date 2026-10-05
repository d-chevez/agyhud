package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/d-chevez/agyhud/internal/config"
	"github.com/d-chevez/agyhud/internal/installer"
	"github.com/d-chevez/agyhud/internal/payload"
)

type tabIndex int

const (
	tabTerminal tabIndex = iota
	tabHUD
	tabWidgets
	tabAppearance
)

type editMode int

const (
	editNone editMode = iota
	editInputText   // Generic text input for hex color, label, symbol
	editAddWidget   // Catalog picker modal
	editInspector   // Dedicated widget inspector modal
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

	// Modal & Inspector State
	mode             editMode
	inputTargetField string // what field is being edited: "global_color", "widget_label", "widget_color", "widget_symbol"
	textInput        textinput.Model
	colorFields      []colorField
	catalogIndex     int
	inspectorCursor  int
}

// InitialModel prepares the TUI model.
func InitialModel(cfgPath string) (*Model, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	hStatus, _ := installer.GetStatus()

	ti := textinput.New()
	ti.CharLimit = 32

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
		activeTab:   tabTerminal,
		cursor:      0,
		hookStatus:  hStatus,
		mode:        editNone,
		textInput:   ti,
		colorFields: fields,
	}, nil
}

func (m *Model) resolveWidgetIndices(cursor int) (int, int) {
	idx := 0
	for r, row := range m.config.Rows {
		for w := range row {
			if idx == cursor {
				return r, w
			}
			idx++
		}
	}
	return -1, -1
}

func (m *Model) getTotalWidgetsCount() int {
	count := 0
	for _, row := range m.config.Rows {
		count += len(row)
	}
	return count
}
