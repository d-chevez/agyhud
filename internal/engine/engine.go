package engine

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/d-chevez/agyhud/internal/config"
	"github.com/d-chevez/agyhud/internal/git"
	"github.com/d-chevez/agyhud/internal/payload"
	"github.com/d-chevez/agyhud/internal/widgets"
	"github.com/muesli/termenv"
)

// Render orchestrates the evaluation of configuration, payload, and widgets into a formatted ANSI HUD.
func Render(p *payload.SessionPayload, cfg *config.Config) string {
	if p == nil || cfg == nil {
		return ""
	}

	lipgloss.SetColorProfile(termenv.TrueColor)

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

	termWidth := p.TerminalWidth
	if termWidth <= 0 {
		termWidth = 120
	}

	var renderedRows []string

	// 2. Dynamic Wrap Mode: widgets flow and wrap automatically when exceeding terminal width
	if cfg.Responsive.Mode == config.LayoutModeDynamic {
		renderedRows = renderDynamicWrap(ctx, cfg, termWidth)
	} else {
		// Manual Rows Mode: renders strict user-defined rows
		renderedRows = renderManualRows(ctx, cfg)
	}

	// 3. Full-width adjustment if requested
	if cfg.Responsive.FullWidth {
		for i, r := range renderedRows {
			curWidth := lipgloss.Width(r)
			if curWidth < termWidth {
				renderedRows[i] = r + strings.Repeat(" ", termWidth-curWidth)
			}
		}
	}

	return strings.Join(renderedRows, "\n")
}

func renderDynamicWrap(ctx widgets.Context, cfg *config.Config, termWidth int) []string {
	var rows []string
	var currentLine strings.Builder
	currentLineWidth := 0
	needSpace := false

	// Flatten all enabled widgets into an ordered flow
	var flatWidgets []config.WidgetConfig
	for _, row := range cfg.Rows {
		for _, w := range row {
			if w.Enabled {
				flatWidgets = append(flatWidgets, w)
			}
		}
	}

	for _, wCfg := range flatWidgets {
		if wCfg.Type == "row_break" {
			res := currentLine.String()
			if strings.TrimSpace(res) != "" {
				rows = append(rows, res)
			}
			currentLine.Reset()
			currentLineWidth = 0
			needSpace = false
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

		wWidth := lipgloss.Width(rendered)
		spacingWidth := 0
		if needSpace {
			spacingWidth = 1
		}

		// Check if widget exceeds available terminal width
		if currentLineWidth > 0 && (currentLineWidth+spacingWidth+wWidth) > termWidth {
			// Wrap to next line
			res := currentLine.String()
			if strings.TrimSpace(res) != "" {
				rows = append(rows, res)
			}
			currentLine.Reset()
			currentLineWidth = 0
			needSpace = false
			spacingWidth = 0
		}

		if needSpace {
			currentLine.WriteString(" ")
			currentLineWidth += 1
		}

		currentLine.WriteString(rendered)
		currentLineWidth += wWidth

		needSpace = !wCfg.Merge
	}

	// Append remaining buffer
	lastLine := currentLine.String()
	if strings.TrimSpace(lastLine) != "" {
		rows = append(rows, lastLine)
	}

	return rows
}

func renderManualRows(ctx widgets.Context, cfg *config.Config) []string {
	var rows []string

	for _, row := range cfg.Rows {
		var rowBuffer strings.Builder
		needSpace := false

		for _, wCfg := range row {
			if !wCfg.Enabled {
				continue
			}

			if wCfg.Type == "row_break" {
				res := rowBuffer.String()
				if strings.TrimSpace(res) != "" {
					rows = append(rows, res)
				}
				rowBuffer.Reset()
				needSpace = false
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
			needSpace = !wCfg.Merge
		}

		res := rowBuffer.String()
		if strings.TrimSpace(res) != "" {
			rows = append(rows, res)
		}
	}

	return rows
}
