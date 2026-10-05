package engine_test

import (
	"os"
	"path/filepath"
	"regexp"
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
					Type:         "model",
					Enabled:      true,
					RawValue:     true,
					Enclose:      true,
					EncloseOpen:  "[",
					EncloseClose: "]",
				},
			},
		},
	}

	output := engine.Render(p, cfg)
	if !strings.Contains(output, "[Gemini 3.8 Flash (High)]") {
		t.Errorf("Expected raw enclosed output with brackets '[Gemini 3.8 Flash (High)]', got: %s", output)
	}
}

func TestRenderSpacingAndMerge(t *testing.T) {
	p := loadSamplePayload(t)

	// 1. Without Merge: space exists between widgets, no leading row space
	cfgUnmerged := &config.Config{
		IconSet:    config.IconSetNerdFont,
		Theme:      config.AntigravityDarkTheme,
		Responsive: config.ResponsiveConfig{Mode: config.LayoutModeManual},
		Rows: [][]config.WidgetConfig{
			{
				{Type: "workspace", Enabled: true, RawValue: true, Merge: false},
				{Type: "model", Enabled: true, RawValue: true, Enclose: true, EncloseOpen: "[", EncloseClose: "]"},
			},
		},
	}
	outUnmerged := engine.Render(p, cfgUnmerged)
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	cleanUnmerged := ansiRegex.ReplaceAllString(outUnmerged, "")
	if !strings.HasPrefix(cleanUnmerged, " ") {
		t.Errorf("Row should have leading space from first widget: %q", cleanUnmerged)
	}
	if !strings.Contains(cleanUnmerged, "  [") {
		t.Errorf("Expected 2 spaces of padding between unmerged widgets: %q", cleanUnmerged)
	}
	if !strings.HasSuffix(cleanUnmerged, " ") {
		t.Errorf("Row should have trailing space from last widget: %q", cleanUnmerged)
	}

	// 2. With Merge: only the posterior padding of the merged widget is removed, leaving the anterior padding of the next widget (1 space total)
	cfgMerged := &config.Config{
		IconSet:    config.IconSetNerdFont,
		Theme:      config.AntigravityDarkTheme,
		Responsive: config.ResponsiveConfig{Mode: config.LayoutModeManual},
		Rows: [][]config.WidgetConfig{
			{
				{Type: "workspace", Enabled: true, RawValue: true, Merge: true},
				{Type: "model", Enabled: true, RawValue: true, Enclose: true, EncloseOpen: "[", EncloseClose: "]"},
			},
		},
	}
	outMerged := engine.Render(p, cfgMerged)
	cleanMerged := ansiRegex.ReplaceAllString(outMerged, "")
	if strings.Contains(cleanMerged, "  [") {
		t.Errorf("Expected merged widget to eliminate trailing space (no double space): %q", cleanMerged)
	}
	if !strings.Contains(cleanMerged, " [") {
		t.Errorf("Expected next widget to keep its leading space (single space): %q", cleanMerged)
	}
}

func TestContextBarPresentationsAndModes(t *testing.T) {
	p := loadSamplePayload(t) // usedPct is 25.0
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

	// 1. Both + Used (default)
	cfgDefault := &config.Config{
		IconSet:    config.IconSetNerdFont,
		Theme:      config.AntigravityDarkTheme,
		Responsive: config.ResponsiveConfig{Mode: config.LayoutModeManual},
		Rows: [][]config.WidgetConfig{
			{
				{Type: "context_bar", Enabled: true, RawValue: true},
			},
		},
	}
	outDefault := ansiRegex.ReplaceAllString(engine.Render(p, cfgDefault), "")
	if !strings.Contains(outDefault, "17.8%") || !strings.Contains(outDefault, "██░░░░░░░░") {
		t.Errorf("Expected both bar and 17.8%% in default mode, got: %s", outDefault)
	}

	// 2. Bar only
	cfgBarOnly := &config.Config{
		IconSet:    config.IconSetNerdFont,
		Theme:      config.AntigravityDarkTheme,
		Responsive: config.ResponsiveConfig{Mode: config.LayoutModeManual},
		Rows: [][]config.WidgetConfig{
			{
				{Type: "context_bar", Enabled: true, RawValue: true, ContextDisplay: "bar"},
			},
		},
	}
	outBarOnly := ansiRegex.ReplaceAllString(engine.Render(p, cfgBarOnly), "")
	if strings.Contains(outBarOnly, "17.8%") || !strings.Contains(outBarOnly, "██░░░░░░░░") {
		t.Errorf("Expected only bar without percentage, got: %s", outBarOnly)
	}

	// 3. Percentage only + Remaining mode
	cfgPctRemaining := &config.Config{
		IconSet:    config.IconSetNerdFont,
		Theme:      config.AntigravityDarkTheme,
		Responsive: config.ResponsiveConfig{Mode: config.LayoutModeManual},
		Rows: [][]config.WidgetConfig{
			{
				{Type: "context_bar", Enabled: true, RawValue: true, ContextDisplay: "percentage", ContextMode: "remaining"},
			},
		},
	}
	outPctRemaining := ansiRegex.ReplaceAllString(engine.Render(p, cfgPctRemaining), "")
	if strings.Contains(outPctRemaining, "░") || !strings.Contains(outPctRemaining, "82.2%") {
		t.Errorf("Expected 82.2%% remaining without bar, got: %s", outPctRemaining)
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
