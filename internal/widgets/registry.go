package widgets

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/d-chevez/agyhud/internal/config"
	"github.com/d-chevez/agyhud/internal/git"
	"github.com/d-chevez/agyhud/internal/payload"
)

// Context encapsulates all runtime telemetry and configuration for rendering widgets.
type Context struct {
	Payload *payload.SessionPayload
	Config  *config.Config
	GitInfo git.Info
}

// Widget is the contract implemented by each telemetry component.
type Widget interface {
	Render(ctx Context, cfg config.WidgetConfig) string
}

// Registry maps widget type identifiers to their renderer instances.
var Registry = map[string]Widget{
	"workspace":   &WorkspaceWidget{},
	"git":         &GitWidget{},
	"agent_state": &AgentStateWidget{},
	"model":       &ModelWidget{},
	"context_bar": &ContextBarWidget{},
	"quota":       &QuotaWidget{},
}

// --- 1. Workspace Widget ---

type WorkspaceWidget struct{}

func (w *WorkspaceWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	dir := ctx.Payload.Workspace.CurrentDir
	if dir == "" {
		dir = ctx.Payload.CWD
	}
	base := filepath.Base(dir)
	if base == "." || base == "/" || base == "\\" {
		base = dir
	}

	icon := "󰉋"
	if ctx.Config.IconSet == config.IconSetClassic {
		icon = "DIR:"
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Accent)).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))

	return fmt.Sprintf("%s %s", dimStyle.Render(icon), style.Render(base))
}

// --- 2. Git Widget ---

type GitWidget struct{}

func (w *GitWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	if !ctx.GitInfo.IsRepo || ctx.GitInfo.Branch == "" {
		return ""
	}

	icon := ""
	if ctx.Config.IconSet == config.IconSetClassic {
		icon = "GIT:"
	}

	branchColor := ctx.Config.Theme.Success
	dirtyMarker := ""
	if ctx.GitInfo.IsDirty {
		branchColor = ctx.Config.Theme.Warning
		dirtyMarker = "*"
	}

	iconStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	branchStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(branchColor)).Bold(true)

	return fmt.Sprintf("%s %s", iconStyle.Render(icon), branchStyle.Render(ctx.GitInfo.Branch+dirtyMarker))
}

// --- 3. Agent State Widget ---

type AgentStateWidget struct{}

func (w *AgentStateWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	state := strings.ToLower(ctx.Payload.AgentState)
	if state == "" {
		state = "ready"
	}

	var icon, label, color string
	isClassic := ctx.Config.IconSet == config.IconSetClassic

	switch state {
	case "thinking":
		icon = "󰟷"
		if isClassic {
			icon = "◆"
		}
		label = "THINKING"
		color = ctx.Config.Theme.Warning
	case "working":
		icon = ""
		if isClassic {
			icon = "⚙"
		}
		label = "WORKING"
		color = ctx.Config.Theme.Accent
	case "tool":
		icon = ""
		if isClassic {
			icon = "🔧"
		}
		label = "TOOL"
		color = "#bb9af7" // Purple
	default: // ready / idle
		icon = ""
		if isClassic {
			icon = "●"
		}
		label = "READY"
		color = ctx.Config.Theme.Success
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true)
	return fmt.Sprintf("%s %s", icon, style.Render(label))
}

// --- 4. Model Widget ---

type ModelWidget struct{}

func (w *ModelWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	name := ctx.Payload.Model.DisplayName
	if name == "" {
		name = ctx.Payload.Model.ID
	}
	if name == "" {
		name = "Gemini"
	}

	icon := "󰚩"
	if ctx.Config.IconSet == config.IconSetClassic {
		icon = "AI:"
	}

	iconStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Text)).Bold(true)

	res := fmt.Sprintf("%s %s", iconStyle.Render(icon), nameStyle.Render(name))

	// If effort tier is set (e.g. "high", "medium"), render a subtle badge
	if ctx.Payload.Model.Effort != "" {
		effortStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ctx.Config.Theme.Dim)).
			Italic(true)
		res += fmt.Sprintf(" %s", effortStyle.Render(fmt.Sprintf("[%s]", ctx.Payload.Model.Effort)))
	}

	return res
}

// --- 5. Context Bar Widget ---

type ContextBarWidget struct{}

func (w *ContextBarWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	pct := ctx.Payload.ContextWindow.UsedPercentage
	if pct <= 0 && ctx.Payload.ContextWindow.ContextWindowSize > 0 {
		totalTokens := float64(ctx.Payload.ContextWindow.TotalInputTokens + ctx.Payload.ContextWindow.TotalOutputTokens)
		pct = (totalTokens / float64(ctx.Payload.ContextWindow.ContextWindowSize)) * 100.0
	}

	// Bar visualization (10 segments)
	totalSegments := 10
	filledSegments := int(math.Round((pct / 100.0) * float64(totalSegments)))
	if filledSegments > totalSegments {
		filledSegments = totalSegments
	}
	if filledSegments < 0 {
		filledSegments = 0
	}

	emptySegments := totalSegments - filledSegments

	barCharFilled := "█"
	barCharEmpty := "░"
	if ctx.Config.IconSet == config.IconSetClassic {
		barCharFilled = "="
		barCharEmpty = "-"
	}

	filledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.BarFilled))
	if pct > 80 {
		filledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Danger))
	} else if pct > 60 {
		filledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Warning))
	}

	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.BarEmpty))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Text))

	bar := filledStyle.Render(strings.Repeat(barCharFilled, filledSegments)) +
		emptyStyle.Render(strings.Repeat(barCharEmpty, emptySegments))

	icon := "󱍏"
	if ctx.Config.IconSet == config.IconSetClassic {
		icon = "CTX:"
	}

	return fmt.Sprintf("%s %s %s",
		dimStyle.Render(icon),
		bar,
		textStyle.Render(fmt.Sprintf("%.1f%%", pct)),
	)
}

// --- 6. Quota Widget ---

type QuotaWidget struct{}

func (w *QuotaWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	if len(ctx.Payload.Quota) == 0 {
		return ""
	}

	// Select primary quota (prefer 'gemini-5h' or first available)
	var qKey string
	if _, ok := ctx.Payload.Quota["gemini-5h"]; ok {
		qKey = "gemini-5h"
	} else if _, ok := ctx.Payload.Quota["3p-5h"]; ok {
		qKey = "3p-5h"
	} else {
		for k := range ctx.Payload.Quota {
			qKey = k
			break
		}
	}

	entry := ctx.Payload.Quota[qKey]
	pct := entry.RemainingFraction * 100.0

	// Format countdown if reset_in_seconds is available
	countdown := ""
	if entry.ResetInSeconds > 0 {
		dur := time.Duration(entry.ResetInSeconds) * time.Second
		hours := int(dur.Hours())
		minutes := int(dur.Minutes()) % 60
		if hours > 0 {
			countdown = fmt.Sprintf(" (%dh%02dm)", hours, minutes)
		} else {
			countdown = fmt.Sprintf(" (%dm)", minutes)
		}
	}

	icon := "󰥔"
	if ctx.Config.IconSet == config.IconSetClassic {
		icon = "QUOTA:"
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	pctColor := ctx.Config.Theme.Success
	if pct < 20 {
		pctColor = ctx.Config.Theme.Danger
	} else if pct < 50 {
		pctColor = ctx.Config.Theme.Warning
	}
	pctStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(pctColor)).Bold(true)

	label := "5h"
	if strings.Contains(qKey, "weekly") {
		label = "wk"
	}

	return fmt.Sprintf("%s %s:%s%s",
		dimStyle.Render(icon),
		dimStyle.Render(label),
		pctStyle.Render(fmt.Sprintf("%.0f%%", pct)),
		dimStyle.Render(countdown),
	)
}
