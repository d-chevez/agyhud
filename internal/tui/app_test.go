package tui_test

import (
	"testing"

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
