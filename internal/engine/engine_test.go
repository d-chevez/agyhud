package engine_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/d-chevez/agyhud/internal/config"
	"github.com/d-chevez/agyhud/internal/engine"
	"github.com/d-chevez/agyhud/internal/payload"
)

func loadSamplePayload(t testing.TB) *payload.SessionPayload {
	fixturePath := filepath.Join("..", "..", ".specs", "fixtures", "payload_sample.json")
	file, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("Failed to open fixture: %v", err)
	}
	defer file.Close()

	p, err := payload.Parse(file)
	if err != nil {
		t.Fatalf("Failed to parse fixture: %v", err)
	}
	return p
}

func TestRenderOutput(t *testing.T) {
	p := loadSamplePayload(t)
	cfg := config.DefaultConfig()

	output := engine.Render(p, cfg)
	if output == "" {
		t.Fatalf("Expected non-empty render output")
	}

	lines := strings.Split(output, "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2-line HUD output for standard width, got %d lines", len(lines))
	}

	// Verify model and agent state are present
	if !strings.Contains(output, "Gemini 3.8 Flash") {
		t.Errorf("Expected output to contain 'Gemini 3.8 Flash'")
	}
	if !strings.Contains(output, "WORKING") {
		t.Errorf("Expected output to contain 'WORKING'")
	}
}

func TestRenderResponsiveCompact(t *testing.T) {
	p := loadSamplePayload(t)
	p.TerminalWidth = 70 // Force narrow width below 90

	cfg := config.DefaultConfig()
	output := engine.Render(p, cfg)

	lines := strings.Split(output, "\n")
	if len(lines) != 1 {
		t.Errorf("Expected compact 1-line HUD for narrow width (70 cols), got %d lines", len(lines))
	}
}

func BenchmarkRender(b *testing.B) {
	p := loadSamplePayload(b)
	cfg := config.DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.Render(p, cfg)
	}
}
