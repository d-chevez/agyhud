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

// Context encapsulates runtime telemetry and configuration for rendering widgets.
type Context struct {
	Payload *payload.SessionPayload
	Config  *config.Config
	GitInfo git.Info
}

// Widget is the contract implemented by each atomic telemetry component.
type Widget interface {
	Render(ctx Context, cfg config.WidgetConfig) string
}

// Registry maps widget type identifiers to their renderer instances.
var Registry = map[string]Widget{
	"custom_symbol":      &CustomSymbolWidget{},
	"workspace":          &WorkspaceWidget{},
	"git_branch":         &GitBranchWidget{},
	"git_status":         &GitStatusWidget{},
	"session_name":       &SessionNameWidget{},
	"session_id":         &SessionIDWidget{},
	"agent_state":        &AgentStateWidget{},
	"model":              &ModelWidget{},
	"thinking_effort":    &ThinkingEffortWidget{},
	"context_bar":        &ContextBarWidget{},
	"context_percentage": &ContextPercentageWidget{},
	"tokens_total":       &TokensTotalWidget{},
	"tokens_input":       &TokensInputWidget{},
	"tokens_output":      &TokensOutputWidget{},
	"session_usage":      &SessionUsageWidget{},
	"reset_timer":        &ResetTimerWidget{},
	"weekly_usage":       &WeeklyUsageWidget{},
	"subagents":          &SubagentsWidget{},
	"separator":          &SeparatorWidget{},
}

// AvailableWidgetTypes lists all supported widget identifiers for the TUI catalog.
var AvailableWidgetTypes = []string{
	"custom_symbol",
	"workspace",
	"git_branch",
	"git_status",
	"session_name",
	"session_id",
	"agent_state",
	"model",
	"thinking_effort",
	"context_bar",
	"context_percentage",
	"tokens_total",
	"tokens_input",
	"tokens_output",
	"session_usage",
	"reset_timer",
	"weekly_usage",
	"subagents",
	"separator",
}

func formatLabelValue(label string, value string, rawValue bool, labelStyle lipgloss.Style, valStyle lipgloss.Style) string {
	if rawValue || label == "" {
		return valStyle.Render(value)
	}
	return fmt.Sprintf("%s %s", labelStyle.Render(label), valStyle.Render(value))
}

// --- 1. Custom Symbol Widget ---

type CustomSymbolWidget struct{}

func (w *CustomSymbolWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	sym := cfg.CustomSymbol
	if sym == "" {
		sym = "•"
	}
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Dim
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(sym)
}

// --- 2. Workspace Widget ---

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

	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Accent
	}

	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold || true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg.Label, base, cfg.RawValue, labelStyle, valStyle)
}

// --- 3. Git Branch Widget ---

type GitBranchWidget struct{}

func (w *GitBranchWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	if !ctx.GitInfo.IsRepo || ctx.GitInfo.Branch == "" {
		return ""
	}
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Success
	}
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold || true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg.Label, ctx.GitInfo.Branch, cfg.RawValue, labelStyle, valStyle)
}

// --- 4. Git Status (Dirty) Widget ---

type GitStatusWidget struct{}

func (w *GitStatusWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	if !ctx.GitInfo.IsRepo {
		return ""
	}
	if ctx.GitInfo.IsDirty {
		color := cfg.Color
		if color == "" {
			color = ctx.Config.Theme.Warning
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render("*")
	}
	return ""
}

// --- 5. Session Name Widget ---

type SessionNameWidget struct{}

func (w *SessionNameWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	title := ctx.Payload.ConversationTitle
	if title == "" {
		return ""
	}
	// Truncate long titles for terminal aesthetics
	if len(title) > 28 {
		title = title[:25] + "..."
	}
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Text
	}
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg.Label, title, cfg.RawValue, labelStyle, valStyle)
}

// --- 6. Session ID Widget ---

type SessionIDWidget struct{}

func (w *SessionIDWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	id := ctx.Payload.SessionID
	if len(id) > 8 {
		id = id[:8]
	}
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Dim
	}
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg.Label, id, cfg.RawValue, labelStyle, valStyle)
}

// --- 7. Agent State Widget ---

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
		color = "#bb9af7"
	default:
		icon = ""
		if isClassic {
			icon = "●"
		}
		label = "READY"
		color = ctx.Config.Theme.Success
	}

	if cfg.Color != "" {
		color = cfg.Color
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true)
	if cfg.RawValue {
		return style.Render(label)
	}
	return fmt.Sprintf("%s %s", icon, style.Render(label))
}

// --- 8. Model Widget ---

type ModelWidget struct{}

func (w *ModelWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	name := ctx.Payload.Model.DisplayName
	if name == "" {
		name = ctx.Payload.Model.ID
	}
	if name == "" {
		name = "Gemini"
	}

	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Text
	}

	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold || true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg.Label, name, cfg.RawValue, labelStyle, valStyle)
}

// --- 9. Thinking Effort Widget ---

type ThinkingEffortWidget struct{}

func (w *ThinkingEffortWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	effort := ctx.Payload.Model.Effort
	if effort == "" {
		return ""
	}
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Dim
	}
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Italic(true)
	return style.Render(fmt.Sprintf("[%s]", effort))
}

// --- 10. Context Bar Widget ---

type ContextBarWidget struct{}

func (w *ContextBarWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	pct := ctx.Payload.ContextWindow.UsedPercentage
	if pct <= 0 && ctx.Payload.ContextWindow.ContextWindowSize > 0 {
		totalTokens := float64(ctx.Payload.ContextWindow.TotalInputTokens + ctx.Payload.ContextWindow.TotalOutputTokens)
		pct = (totalTokens / float64(ctx.Payload.ContextWindow.ContextWindowSize)) * 100.0
	}

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

	filledColor := ctx.Config.Theme.BarFilled
	if pct > 80 {
		filledColor = ctx.Config.Theme.Danger
	} else if pct > 60 {
		filledColor = ctx.Config.Theme.Warning
	}
	if cfg.Color != "" {
		filledColor = cfg.Color
	}

	filledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(filledColor))
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.BarEmpty))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))

	bar := filledStyle.Render(strings.Repeat(barCharFilled, filledSegments)) +
		emptyStyle.Render(strings.Repeat(barCharEmpty, emptySegments))

	return formatLabelValue(cfg.Label, bar, cfg.RawValue, labelStyle, lipgloss.NewStyle())
}

// --- 11. Context Percentage Widget ---

type ContextPercentageWidget struct{}

func (w *ContextPercentageWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	pct := ctx.Payload.ContextWindow.UsedPercentage
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Text
		if pct > 80 {
			color = ctx.Config.Theme.Danger
		} else if pct > 60 {
			color = ctx.Config.Theme.Warning
		}
	}
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg.Label, fmt.Sprintf("%.1f%%", pct), cfg.RawValue, labelStyle, valStyle)
}

// --- 12. Tokens Total Widget ---

type TokensTotalWidget struct{}

func formatTokenCount(n int64) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000.0)
	}
	if n >= 1000 {
		return fmt.Sprintf("%dk", n/1000)
	}
	return fmt.Sprintf("%d", n)
}

func (w *TokensTotalWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	used := ctx.Payload.ContextWindow.TotalInputTokens + ctx.Payload.ContextWindow.TotalOutputTokens
	total := ctx.Payload.ContextWindow.ContextWindowSize
	if used == 0 && ctx.Payload.ContextWindow.CurrentUsage.InputTokens > 0 {
		used = ctx.Payload.ContextWindow.CurrentUsage.InputTokens + ctx.Payload.ContextWindow.CurrentUsage.OutputTokens
	}
	str := fmt.Sprintf("(%s/%s)", formatTokenCount(used), formatTokenCount(total))
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Dim
	}
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg.Label, str, cfg.RawValue, labelStyle, valStyle)
}

// --- 13. Tokens Input / Output Widgets ---

type TokensInputWidget struct{}

func (w *TokensInputWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	n := ctx.Payload.ContextWindow.TotalInputTokens
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Text))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg.Label, formatTokenCount(n), cfg.RawValue, labelStyle, valStyle)
}

type TokensOutputWidget struct{}

func (w *TokensOutputWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	n := ctx.Payload.ContextWindow.TotalOutputTokens
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Text))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg.Label, formatTokenCount(n), cfg.RawValue, labelStyle, valStyle)
}

// --- 14. Session Usage (5h) Widget ---

type SessionUsageWidget struct{}

func (w *SessionUsageWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	var entry payload.QuotaEntry
	if q, ok := ctx.Payload.Quota["gemini-5h"]; ok {
		entry = q
	} else if q, ok := ctx.Payload.Quota["3p-5h"]; ok {
		entry = q
	} else {
		return ""
	}

	pct := entry.RemainingFraction * 100.0
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Success
		if pct < 20 {
			color = ctx.Config.Theme.Danger
		} else if pct < 50 {
			color = ctx.Config.Theme.Warning
		}
	}

	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold || true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg.Label, fmt.Sprintf("%.0f%%", pct), cfg.RawValue, labelStyle, valStyle)
}

// --- 15. Reset Timer Widget ---

type ResetTimerWidget struct{}

func (w *ResetTimerWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	var entry payload.QuotaEntry
	if q, ok := ctx.Payload.Quota["gemini-5h"]; ok {
		entry = q
	} else if q, ok := ctx.Payload.Quota["3p-5h"]; ok {
		entry = q
	} else {
		return ""
	}

	if entry.ResetInSeconds <= 0 {
		return ""
	}

	dur := time.Duration(entry.ResetInSeconds) * time.Second
	hours := int(dur.Hours())
	minutes := int(dur.Minutes()) % 60
	str := fmt.Sprintf("(%dm)", minutes)
	if hours > 0 {
		str = fmt.Sprintf("(%dh%02dm)", hours, minutes)
	}

	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Dim
	}
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg.Label, str, cfg.RawValue, labelStyle, valStyle)
}

// --- 16. Weekly Usage Widget ---

type WeeklyUsageWidget struct{}

func (w *WeeklyUsageWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	var entry payload.QuotaEntry
	if q, ok := ctx.Payload.Quota["gemini-weekly"]; ok {
		entry = q
	} else if q, ok := ctx.Payload.Quota["3p-weekly"]; ok {
		entry = q
	} else {
		return ""
	}

	pct := entry.RemainingFraction * 100.0
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Success
		if pct < 20 {
			color = ctx.Config.Theme.Danger
		} else if pct < 50 {
			color = ctx.Config.Theme.Warning
		}
	}
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg.Label, fmt.Sprintf("%.0f%%", pct), cfg.RawValue, labelStyle, valStyle)
}

// --- 17. Subagents Widget ---

type SubagentsWidget struct{}

func (w *SubagentsWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	// Antigravity subagent metrics (default 0 if none active)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Accent)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg.Label, "0", cfg.RawValue, labelStyle, valStyle)
}

// --- 18. Separator Widget ---

type SeparatorWidget struct{}

func (w *SeparatorWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	sep := cfg.Separator
	if sep == "" {
		sep = "│"
	}
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Dim
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(sep)
}
