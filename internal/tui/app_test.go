package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/d-chevez/agyhud/internal/tui"
)

func TestInitialModel(t *testing.T) {
	m, err := tui.InitialModel("")
	if err != nil {
		t.Fatalf("Failed to initialize TUI model: %v", err)
	}

	view := m.View()
	if view == "" {
		t.Fatalf("Expected non-empty view output")
	}

	// Verify header title and tabs
	if !testing.Short() {
		// Basic sanity check that Lipgloss formatted the title
		if len(view) < 50 {
			t.Errorf("View output too short: %s", view)
		}
	}
}

func TestWidgetSubmenuNavigation(t *testing.T) {
	m, err := tui.InitialModel("")
	if err != nil {
		t.Fatalf("Failed to initialize TUI model: %v", err)
	}

	// 1. Navigate to Widgets screen (option 3 in Main Menu)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = newModel.(*tui.Model)

	view := m.View()
	if !strings.Contains(view, "Widgets (CRUD & Layout)") {
		t.Fatalf("Expected Widgets screen, got: %s", view)
	}

	// 2. Press 'a' to open Add Widget Categories submenu
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = newModel.(*tui.Model)

	view = m.View()
	if !strings.Contains(view, "SELECT WIDGET CATEGORY") {
		t.Fatalf("Expected Categories submenu, got: %s", view)
	}
	if !strings.Contains(view, "Git Telemetry") {
		t.Fatalf("Expected category list to include Git Telemetry")
	}

	// 3. Press '2' to open 'Git Telemetry' category
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m = newModel.(*tui.Model)

	view = m.View()
	if !strings.Contains(view, "CATEGORY: 🌿 Git Telemetry") {
		t.Fatalf("Expected Git Telemetry category screen, got: %s", view)
	}
	if !strings.Contains(view, "git_branch") {
		t.Fatalf("Expected category to contain git_branch")
	}

	// 4. Press Esc to go back to Categories submenu
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = newModel.(*tui.Model)

	view = m.View()
	if !strings.Contains(view, "SELECT WIDGET CATEGORY") {
		t.Fatalf("Expected back to Categories submenu, got: %s", view)
	}

	// 5. Select category again and press Enter on first item
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(*tui.Model)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(*tui.Model)

	// Should be back in Widgets screen with widget added
	view = m.View()
	if !strings.Contains(view, "Added") {
		t.Fatalf("Expected widget added status, got: %s", view)
	}
}
