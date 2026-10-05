package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/d-chevez/agyhud/internal/config"
	"github.com/d-chevez/agyhud/internal/installer"
	"github.com/d-chevez/agyhud/internal/payload"
)

type screenState int

const (
	screenMainMenu screenState = iota
	screenTerminal
	screenHUD
	screenWidgets
	screenWidgetCatalog
	screenAppearance
)

type editMode int

const (
	editNone editMode = iota
	editInputText // Text input prompt active
)

type widgetRef struct {
	Row int
	Col int
}

// Model is the main Bubbletea state model for the agyhud configuration TUI.
type Model struct {
	config      *config.Config
	configPath  string
	payload     *payload.SessionPayload
	screenStack []screenState
	statusMsg   string
	width       int
	height      int
	hookStatus  installer.HookStatus
	quitting    bool

	// Dedicated cursors per screen level to preserve user position
	mainCursor       int
	terminalCursor   int
	hudCursor        int
	widgetsCursor    int
	catalogCursor    int
	appearanceCursor int

	// Reorder / Move state in Widgets screen
	isReordering bool

	// Modal / Inline Text Input State
	mode             editMode
	inputTargetField string // "widget_color", "widget_label", "widget_symbol"
	textInput        textinput.Model
}

// InitialModel prepares the TUI model starting at the root Main Menu.
func InitialModel(cfgPath string) (*Model, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	hStatus, _ := installer.GetStatus()

	ti := textinput.New()
	ti.CharLimit = 32

	return &Model{
		config:      cfg,
		configPath:  cfgPath,
		payload:     GetSamplePayload(),
		screenStack: []screenState{screenMainMenu},
		hookStatus:  hStatus,
		mode:        editNone,
		textInput:   ti,
	}, nil
}

func (m *Model) currentScreen() screenState {
	if len(m.screenStack) == 0 {
		return screenMainMenu
	}
	return m.screenStack[len(m.screenStack)-1]
}

func (m *Model) pushScreen(s screenState) {
	m.screenStack = append(m.screenStack, s)
	m.statusMsg = ""
}

func (m *Model) popScreen() bool {
	if len(m.screenStack) <= 1 {
		return false
	}
	m.screenStack = m.screenStack[:len(m.screenStack)-1]
	m.statusMsg = ""
	return true
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

func (m *Model) getWidgetsList() []widgetRef {
	var refs []widgetRef
	for r, row := range m.config.Rows {
		for c, w := range row {
			if w.Type != "row_break" {
				refs = append(refs, widgetRef{Row: r, Col: c})
			}
		}
	}
	return refs
}

func (m *Model) getTotalWidgetsCount() int {
	count := 0
	for _, row := range m.config.Rows {
		count += len(row)
	}
	return count
}

func (m *Model) getActiveWidgetsCount() int {
	count := 0
	for _, row := range m.config.Rows {
		for _, w := range row {
			if w.Enabled {
				count++
			}
		}
	}
	return count
}
