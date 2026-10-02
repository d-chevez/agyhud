package payload_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/d-chevez/agyhud/internal/payload"
)

func TestParsePayload(t *testing.T) {
	fixturePath := filepath.Join("..", "..", ".specs", "fixtures", "payload_sample.json")
	file, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("Failed to open fixture: %v", err)
	}
	defer file.Close()

	p, err := payload.Parse(file)
	if err != nil {
		t.Fatalf("Failed to parse payload: %v", err)
	}

	if p.Model.DisplayName != "Gemini 3.8 Flash (High)" {
		t.Errorf("Expected model 'Gemini 3.8 Flash (High)', got '%s'", p.Model.DisplayName)
	}
	if p.AgentState != "working" {
		t.Errorf("Expected agent_state 'working', got '%s'", p.AgentState)
	}
	if p.ContextWindow.ContextWindowSize != 1048576 {
		t.Errorf("Expected context_window_size 1048576, got %d", p.ContextWindow.ContextWindowSize)
	}
	if len(p.Quota) == 0 {
		t.Errorf("Expected non-empty quota map")
	}
	if p.TerminalWidth != 188 {
		t.Errorf("Expected terminal_width 188, got %d", p.TerminalWidth)
	}
}
