package engine

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/d-chevez/agyhud/internal/config"
	"github.com/d-chevez/agyhud/internal/git"
	"github.com/d-chevez/agyhud/internal/payload"
	"github.com/d-chevez/agyhud/internal/widgets"
)

// Render orchestrates the evaluation of configuration, payload, and widgets into a formatted ANSI HUD.
func Render(p *payload.SessionPayload, cfg *config.Config) string {
	if p == nil || cfg == nil {
		return ""
	}

	// 1. Gather Git Telemetry
	targetDir := p.Workspace.CurrentDir
	if targetDir == "" {
		targetDir = p.CWD
	}
	gitInfo := git.GetInfo(targetDir, cfg.Git.RefreshSeconds, cfg.Git.TimeoutMs)

	ctx := widgets.Context{
		Payload: p,
		Config:  cfg,
		GitInfo: gitInfo,
	}

	// 2. Responsive handling: collapse to compact single line if terminal is narrow
	rowsToRender := cfg.Rows
	if cfg.Responsive.Enabled && p.TerminalWidth > 0 && p.TerminalWidth < cfg.Responsive.BreakpointWidth {
		rowsToRender = buildCompactRow(cfg)
	}

	// 3. Render rows
	var renderedRows []string
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Theme.Dim))

	for _, row := range rowsToRender {
		var rowParts []string
		for _, wCfg := range row {
			if !wCfg.Enabled {
				continue
			}

			w, exists := widgets.Registry[wCfg.Type]
			if !exists {
				continue
			}

			rendered := w.Render(ctx, wCfg)
			if strings.TrimSpace(rendered) == "" {
				continue
			}

			// Apply padding
			pad := strings.Repeat(" ", wCfg.Padding)
			content := pad + rendered + pad

			rowParts = append(rowParts, content)
		}

		if len(rowParts) > 0 {
			// Join widgets with styled separator
			joined := strings.Join(rowParts, sepStyle.Render("│"))
			renderedRows = append(renderedRows, joined)
		}
	}

	return strings.Join(renderedRows, "\n")
}

// buildCompactRow creates an adaptive single-line layout when terminal width is constrained.
func buildCompactRow(cfg *config.Config) [][]config.WidgetConfig {
	return [][]config.WidgetConfig{
		{
			{Type: "agent_state", Enabled: true, Padding: 0, Separator: "│"},
			{Type: "model", Enabled: true, Padding: 1, Separator: "│"},
			{Type: "context_bar", Enabled: true, Padding: 1, Separator: "│"},
			{Type: "git", Enabled: true, Padding: 1, Separator: ""},
		},
	}
}
