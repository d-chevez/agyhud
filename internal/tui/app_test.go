package tui_test

import (
	"path/filepath"
	"regexp"
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
	if !strings.Contains(view, "CATEGORY: Git Telemetry") {
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

func TestWidgetBoldAndRawEnclosing(t *testing.T) {
	tempDir := t.TempDir()
	cfgFile := filepath.Join(tempDir, "config.json")

	m, err := tui.InitialModel(cfgFile)
	if err != nil {
		t.Fatalf("Failed to initialize TUI model: %v", err)
	}

	// Navigate to Widgets screen (option 3)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = newModel.(*tui.Model)

	// Move cursor to workspace widget (index 1)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(*tui.Model)

	// 1. Verify placeholder label is visible in management view when !RawValue and no custom label
	view := m.View()
	if !strings.Contains(view, "Workspace:") {
		t.Fatalf("Expected placeholder label 'Workspace:' to appear when !RawValue, got: %s", view)
	}

	// 2. Enclose option must NOT appear in footer when !RawValue
	view = m.View()
	if strings.Contains(view, "[c] Enclose") {
		t.Fatalf("Enclose option must NOT appear in footer when RAW is not active: %s", view)
	}
	if strings.Contains(view, "[e] Edit") {
		t.Fatalf("Edit option [e] must NOT appear in footer: %s", view)
	}

	// 3. Toggle Bold with 'b'
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	m = newModel.(*tui.Model)
	view = m.View()
	if !strings.Contains(view, "[BOLD]") {
		t.Fatalf("Expected [BOLD] badge after pressing 'b', got: %s", view)
	}

	// 4. Toggle Raw with 'r'
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(*tui.Model)
	view = m.View()
	if !strings.Contains(view, "[RAW]") {
		t.Fatalf("Expected [RAW] badge after pressing 'r', got: %s", view)
	}
	// Enclose option MUST now appear in footer since RAW is active
	if !strings.Contains(view, "[c] Enclose") {
		t.Fatalf("Expected [c] Enclose in footer when RAW is active, got: %s", view)
	}
	// Placeholder must NOT appear when RAW
	lines := strings.Split(view, "\n")
	for _, l := range lines {
		if strings.Contains(l, "workspace") && strings.Contains(l, "[RAW]") {
			if strings.Contains(l, "Workspace:") {
				t.Fatalf("Placeholder label must NOT appear when RAW is active: %s", l)
			}
		}
	}

	// 5. Now that RAW is active, set enclosing characters using 'c'
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = newModel.(*tui.Model)

	// Enter "[]"
	for _, ch := range "[]" {
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = newModel.(*tui.Model)
	}
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(*tui.Model)

	view = m.View()
	if !strings.Contains(view, "[RAW]") {
		t.Fatalf("Expected [RAW] badge in widgets list, got: %s", view)
	}
	if !strings.Contains(view, "[ENCLOSE]") {
		t.Fatalf("Expected separate [ENCLOSE] badge in widgets list, got: %s", view)
	}

	// 6. Pressing Enter must edit label
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(*tui.Model)
	view = m.View()
	cleanView := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`).ReplaceAllString(view, "")
	if !strings.Contains(cleanView, "Label prefix") {
		t.Fatalf("Expected Enter to open label editor with placeholder 'Label prefix', got: %s", cleanView)
	}
}

func TestContextBarTUIToggles(t *testing.T) {
	tempDir := t.TempDir()
	cfgFile := filepath.Join(tempDir, "config.json")

	m, err := tui.InitialModel(cfgFile)
	if err != nil {
		t.Fatalf("Failed to initialize TUI model: %v", err)
	}

	// Navigate to Widgets screen (option 3)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = newModel.(*tui.Model)

	// Move cursor to context_bar
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	for i := 0; i < 30; i++ {
		clean := ansiRegex.ReplaceAllString(m.View(), "")
		if strings.Contains(clean, "▶ context_bar") || strings.Contains(clean, "▶  context_bar") {
			break
		}
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = newModel.(*tui.Model)
	}

	view := m.View()
	if !strings.Contains(view, "[BAR+%|USED]") {
		t.Fatalf("Expected default badge [BAR+%%|USED], got: %s", view)
	}

	// Press 'p' to switch to Bar only
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = newModel.(*tui.Model)
	view = m.View()
	if !strings.Contains(view, "[BAR|USED]") {
		t.Fatalf("Expected badge [BAR|USED] after pressing 'p', got: %s", view)
	}

	// Press 'o' to switch to Remaining mode
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = newModel.(*tui.Model)
	view = m.View()
	if !strings.Contains(view, "[BAR|REMAINING]") {
		t.Fatalf("Expected badge [BAR|REMAINING] after pressing 'o', got: %s", view)
	}
}
