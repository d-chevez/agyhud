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
	"row_break":          &RowBreakWidget{},
}

// CatalogCategory defines a grouped category of widgets for organized browsing.
type CatalogCategory struct {
	Name    string
	Widgets []CatalogItem
}

// CatalogItem describes an available widget in the catalog.
type CatalogItem struct {
	Type        string
	Name        string
	Description string
}

// CatalogCategories organizes all available widgets into logical functional groups.
var CatalogCategories = []CatalogCategory{
	{
		Name: "📁 Workspace & Session",
		Widgets: []CatalogItem{
			{Type: "workspace", Name: "Workspace", Description: "Current active workspace folder name"},
			{Type: "session_name", Name: "Session Name", Description: "Active conversation session title"},
			{Type: "session_id", Name: "Session ID", Description: "Session UUID identifier"},
			{Type: "agent_state", Name: "Agent State", Description: "Agent status badge (READY, RUNNING, etc.)"},
		},
	},
	{
		Name: "🌿 Git Telemetry",
		Widgets: []CatalogItem{
			{Type: "git_branch", Name: "Git Branch", Description: "Current active git branch name"},
			{Type: "git_status", Name: "Git Status", Description: "Dirty repo state indicator (*)"},
		},
	},
	{
		Name: "🧠 Model & AI Intelligence",
		Widgets: []CatalogItem{
			{Type: "model", Name: "Model", Description: "Active LLM model name (e.g. Gemini 2.5 Pro)"},
			{Type: "thinking_effort", Name: "Thinking Effort", Description: "Thinking budget effort level indicator"},
			{Type: "subagents", Name: "Subagents", Description: "Count of actively running subagents"},
		},
	},
	{
		Name: "📊 Context & Quota Usage",
		Widgets: []CatalogItem{
			{Type: "context_bar", Name: "Context Bar", Description: "Visual progress bar of context window usage"},
			{Type: "context_percentage", Name: "Context Percentage", Description: "Context window usage percent"},
			{Type: "tokens_total", Name: "Tokens Total", Description: "Total tokens consumed in current session"},
			{Type: "tokens_input", Name: "Tokens Input", Description: "Prompt input tokens count"},
			{Type: "tokens_output", Name: "Tokens Output", Description: "Completion tokens generated"},
			{Type: "session_usage", Name: "Session Usage", Description: "5-hour sliding quota usage percentage"},
			{Type: "reset_timer", Name: "Reset Timer", Description: "Time remaining until 5-hour quota resets"},
			{Type: "weekly_usage", Name: "Weekly Usage", Description: "Weekly quota usage percentage"},
		},
	},
	{
		Name: "📐 Layout & Spacers",
		Widgets: []CatalogItem{
			{Type: "row_break", Name: "Row Break", Description: "Forces a new line / row in HUD layout"},
			{Type: "separator", Name: "Separator / Spacer", Description: "Custom delimiter symbol (e.g. │, •, |)"},
			{Type: "custom_symbol", Name: "Custom Symbol", Description: "Custom decorative icon or glyph"},
		},
	},
}

// FlatCatalog returns all available widgets in order.
func FlatCatalog() []CatalogItem {
	var items []CatalogItem
	for _, cat := range CatalogCategories {
		items = append(items, cat.Widgets...)
	}
	return items
}

// AvailableWidgetTypes lists all supported widget identifiers.
var AvailableWidgetTypes = func() []string {
	var list []string
	for _, item := range FlatCatalog() {
		list = append(list, item.Type)
	}
	return list
}()

// DefaultLabel returns the user-friendly default label for a widget type (without underscores).
func DefaultLabel(widgetType string) string {
	switch widgetType {
	case "separator", "custom_symbol", "row_break", "git_status":
		return ""
	}
	parts := strings.Split(widgetType, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ") + ":"
}

func formatLabelValue(cfg config.WidgetConfig, value string, labelStyle lipgloss.Style, valStyle lipgloss.Style) string {
	if cfg.Bold {
		valStyle = valStyle.Bold(true)
		labelStyle = labelStyle.Bold(true)
	}
	if cfg.RawValue {
		cleanVal := strings.TrimSpace(value)
		prefix := strings.TrimSpace(cfg.RawPrefix)
		suffix := strings.TrimSpace(cfg.RawSuffix)
		if prefix != "" || suffix != "" {
			cleanVal = fmt.Sprintf("%s%s%s", prefix, cleanVal, suffix)
		}
		return valStyle.Render(cleanVal)
	}
	label := cfg.Label
	if label == "" {
		label = DefaultLabel(cfg.Type)
	}
	if label == "" {
		return valStyle.Render(value)
	}
	return fmt.Sprintf("%s %s", labelStyle.Render(label), valStyle.Render(value))
}

func formatTokenCount(n int64) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000.0)
	}
	if n >= 1000 {
		return fmt.Sprintf("%dk", n/1000)
	}
	return fmt.Sprintf("%d", n)
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
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	if cfg.Bold {
		style = style.Bold(true)
	}
	return style.Render(sym)
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

	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg, base, labelStyle, valStyle)
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
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg, ctx.GitInfo.Branch, labelStyle, valStyle)
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
	if len(title) > 28 {
		title = title[:25] + "..."
	}
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Text
	}
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg, title, labelStyle, valStyle)
}

// --- 6. Session ID Widget ---

type SessionIDWidget struct{}

func (w *SessionIDWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	id := ctx.Payload.ConversationID
	if id == "" {
		return ""
	}
	if len(id) > 8 {
		id = id[:8]
	}
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Dim
	}
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg, id, labelStyle, valStyle)
}

// --- 7. Agent State Widget ---

type AgentStateWidget struct{}

func (w *AgentStateWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	state := ctx.Payload.AgentState
	if state == "" {
		state = "READY"
	}

	color := cfg.Color
	sym := ""
	if ctx.Config.IconSet == config.IconSetClassic {
		sym = "[*]"
	}

	switch strings.ToUpper(state) {
	case "RUNNING", "ACTIVE":
		if color == "" {
			color = ctx.Config.Theme.Warning
		}
		if ctx.Config.IconSet == config.IconSetNerdFont {
			sym = "󱐌"
		}
	case "ERROR", "FAILED":
		if color == "" {
			color = ctx.Config.Theme.Danger
		}
		if ctx.Config.IconSet == config.IconSetNerdFont {
			sym = "󰅚"
		}
	default:
		if color == "" {
			color = ctx.Config.Theme.Success
		}
	}

	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))

	if cfg.RawValue {
		return formatLabelValue(cfg, strings.ToUpper(state), labelStyle, valStyle)
	}

	stateStr := fmt.Sprintf("%s %s", sym, strings.ToUpper(state))
	return formatLabelValue(cfg, stateStr, labelStyle, valStyle)
}

// --- 8. Model Widget ---

type ModelWidget struct{}

func (w *ModelWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	model := ctx.Payload.Model.DisplayName
	if model == "" {
		model = ctx.Payload.Model.ID
	}
	if model == "" {
		model = "Gemini"
	}
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Accent
	}
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg, model, labelStyle, valStyle)
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
		color = ctx.Config.Theme.Warning
	}
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))

	if cfg.RawValue {
		return formatLabelValue(cfg, effort, labelStyle, valStyle)
	}

	sym := "󰧑"
	if ctx.Config.IconSet == config.IconSetClassic {
		sym = "~"
	}
	effortStr := fmt.Sprintf("%s %s", sym, effort)
	return formatLabelValue(cfg, effortStr, labelStyle, valStyle)
}

// --- 10. Context Bar Widget ---

type ContextBarWidget struct{}

func (w *ContextBarWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	pct := ctx.Payload.ContextWindow.UsedPercentage
	width := 10
	filled := int(math.Round((pct / 100.0) * float64(width)))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	barColor := cfg.Color
	if barColor == "" {
		barColor = ctx.Config.Theme.BarFilled
		if pct > 85.0 {
			barColor = ctx.Config.Theme.Danger
		} else if pct > 65.0 {
			barColor = ctx.Config.Theme.Warning
		}
	}

	filledChar := "█"
	emptyChar := "░"
	if ctx.Config.IconSet == config.IconSetClassic {
		filledChar = "#"
		emptyChar = "-"
	}

	filledPart := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor)).Render(strings.Repeat(filledChar, filled))
	emptyPart := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.BarEmpty)).Render(strings.Repeat(emptyChar, width-filled))
	bar := filledPart + emptyPart

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	valStyle := lipgloss.NewStyle()
	return formatLabelValue(cfg, bar, labelStyle, valStyle)
}

// --- 11. Context Percentage Widget ---

type ContextPercentageWidget struct{}

func (w *ContextPercentageWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	pct := ctx.Payload.ContextWindow.UsedPercentage
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Text
		if pct > 85.0 {
			color = ctx.Config.Theme.Danger
		} else if pct > 65.0 {
			color = ctx.Config.Theme.Warning
		}
	}
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg, fmt.Sprintf("%.1f%%", pct), labelStyle, valStyle)
}

// --- 12. Tokens Total Widget ---

type TokensTotalWidget struct{}

func (w *TokensTotalWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	used := ctx.Payload.ContextWindow.TotalInputTokens + ctx.Payload.ContextWindow.TotalOutputTokens
	total := ctx.Payload.ContextWindow.ContextWindowSize
	if used == 0 && ctx.Payload.ContextWindow.CurrentUsage.InputTokens > 0 {
		used = ctx.Payload.ContextWindow.CurrentUsage.InputTokens + ctx.Payload.ContextWindow.CurrentUsage.OutputTokens
	}
	rawTokens := fmt.Sprintf("%s/%s", formatTokenCount(used), formatTokenCount(total))
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Dim
	}
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	if cfg.RawValue {
		return formatLabelValue(cfg, rawTokens, labelStyle, valStyle)
	}
	return formatLabelValue(cfg, "("+rawTokens+")", labelStyle, valStyle)
}

// --- 13. Tokens Input Widget ---

type TokensInputWidget struct{}

func (w *TokensInputWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	n := ctx.Payload.ContextWindow.TotalInputTokens
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Text
	}
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg, formatTokenCount(n), labelStyle, valStyle)
}

// --- 14. Tokens Output Widget ---

type TokensOutputWidget struct{}

func (w *TokensOutputWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	n := ctx.Payload.ContextWindow.TotalOutputTokens
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Text
	}
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg, formatTokenCount(n), labelStyle, valStyle)
}

// --- 15. Session Usage (5-Hour Quota) Widget ---

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

	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg, fmt.Sprintf("%.0f%%", pct), labelStyle, valStyle)
}

// --- 16. Reset Timer Widget ---

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
	rawTime := fmt.Sprintf("%dm", minutes)
	if hours > 0 {
		rawTime = fmt.Sprintf("%dh%02dm", hours, minutes)
	}

	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Dim
	}
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	if cfg.RawValue {
		return formatLabelValue(cfg, rawTime, labelStyle, valStyle)
	}
	return formatLabelValue(cfg, "("+rawTime+")", labelStyle, valStyle)
}

// --- 17. Weekly Usage Widget ---

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
	return formatLabelValue(cfg, fmt.Sprintf("%.0f%%", pct), labelStyle, valStyle)
}

// --- 18. Subagents Widget ---

type SubagentsWidget struct{}

func (w *SubagentsWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	color := cfg.Color
	if color == "" {
		color = ctx.Config.Theme.Accent
	}
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(cfg.Bold)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ctx.Config.Theme.Dim))
	return formatLabelValue(cfg, "0", labelStyle, valStyle)
}

// --- 19. Separator Widget ---

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
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	if cfg.Bold {
		style = style.Bold(true)
	}
	return style.Render(sep)
}

// --- 20. Row Break Widget ---

type RowBreakWidget struct{}

func (w *RowBreakWidget) Render(ctx Context, cfg config.WidgetConfig) string {
	return ""
}
