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

func TestRenderDynamicWrap(t *testing.T) {
	p := loadSamplePayload(t)
	cfg := config.DefaultConfig()
	cfg.Responsive.Mode = config.LayoutModeDynamic

	// Test wide terminal (e.g. 180 cols) -> fewer rows
	p.TerminalWidth = 180
	outputWide := engine.Render(p, cfg)
	linesWide := strings.Split(outputWide, "\n")

	// Test narrow terminal (e.g. 70 cols) -> more wrapped rows
	p.TerminalWidth = 70
	outputNarrow := engine.Render(p, cfg)
	linesNarrow := strings.Split(outputNarrow, "\n")

	if len(linesNarrow) <= len(linesWide) {
		t.Errorf("Expected narrow terminal (%d lines) to wrap into more lines than wide terminal (%d lines)", len(linesNarrow), len(linesWide))
	}
}

func TestRenderManualRows(t *testing.T) {
	p := loadSamplePayload(t)
	cfg := config.DefaultConfig()
	cfg.Responsive.Mode = config.LayoutModeManual

	output := engine.Render(p, cfg)
	lines := strings.Split(output, "\n")
	if len(lines) != 2 {
		t.Errorf("Expected exactly 2 manual rows, got %d", len(lines))
	}
}

func TestRenderRawEnclosing(t *testing.T) {
	p := loadSamplePayload(t)
	cfg := &config.Config{
		IconSet:    config.IconSetNerdFont,
		Theme:      config.AntigravityDarkTheme,
		Responsive: config.ResponsiveConfig{Mode: config.LayoutModeManual},
		Rows: [][]config.WidgetConfig{
			{
				{
					Type:      "model",
					Enabled:   true,
					RawValue:  true,
					RawPrefix: "[",
					RawSuffix: "]",
				},
			},
		},
	}

	output := engine.Render(p, cfg)
	if !strings.Contains(output, "[Gemini 3.8 Flash (High)]") {
		t.Errorf("Expected raw enclosed output with brackets '[Gemini 3.8 Flash (High)]', got: %s", output)
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
