package engine

import (
	"strings"

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

	// 3. Render rows with atomic merging logic
	var renderedRows []string

	for _, row := range rowsToRender {
		var rowBuffer strings.Builder
		needSpace := false

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

			if needSpace {
				rowBuffer.WriteString(" ")
			}

			rowBuffer.WriteString(rendered)

			// Merge control: if Merge is true, the next widget attaches without space
			needSpace = !wCfg.Merge
		}

		res := rowBuffer.String()
		if strings.TrimSpace(res) != "" {
			renderedRows = append(renderedRows, " "+res+" ")
		}
	}

	return strings.Join(renderedRows, "\n")
}

// buildCompactRow creates an adaptive single-line layout when terminal width is constrained.
func buildCompactRow(cfg *config.Config) [][]config.WidgetConfig {
	return [][]config.WidgetConfig{
		{
			{Type: "agent_state", Enabled: true, Merge: false},
			{Type: "separator", Separator: "│", Enabled: true, Merge: false},
			{Type: "model", RawValue: true, Enabled: true, Merge: false},
			{Type: "separator", Separator: "│", Enabled: true, Merge: false},
			{Type: "context_percentage", RawValue: true, Enabled: true, Merge: false},
			{Type: "separator", Separator: "│", Enabled: true, Merge: false},
			{Type: "git_branch", RawValue: true, Enabled: true, Merge: true},
			{Type: "git_status", Enabled: true, Merge: false},
		},
	}
}
