package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/d-chevez/agyhud/internal/config"
)

var themePresets = []struct {
	name  string
	theme config.ThemeConfig
}{
	{"Tokyo Night (Modern Soft)", config.ThemeConfig{Accent: "#7aa2f7", Dim: "#565f89", Text: "#c0caf5", Success: "#9ece6a", Warning: "#e0af68", Danger: "#f7768e", BarFilled: "#7aa2f7", BarEmpty: "#24283b"}},
	{"Catppuccin Mocha (Pastel)", config.ThemeConfig{Accent: "#89b4fa", Dim: "#6c7086", Text: "#cdd6f4", Success: "#a6e3a1", Warning: "#f9e2af", Danger: "#f38ba8", BarFilled: "#89b4fa", BarEmpty: "#313244"}},
	{"Gruvbox (Warm Retro)", config.ThemeConfig{Accent: "#83a598", Dim: "#928374", Text: "#ebdbb2", Success: "#b8bb26", Warning: "#fabd2f", Danger: "#fb4934", BarFilled: "#83a598", BarEmpty: "#3c3836"}},
	{"Nord (Arctic Minimal)", config.ThemeConfig{Accent: "#88c0d0", Dim: "#4c566a", Text: "#eceff4", Success: "#a3be8c", Warning: "#ebcb8b", Danger: "#bf616a", BarFilled: "#88c0d0", BarEmpty: "#3b4252"}},
	{"Cyberpunk (High Neon)", config.ThemeConfig{Accent: "#00ffff", Dim: "#711c91", Text: "#f8f8f2", Success: "#05ffa1", Warning: "#ffe600", Danger: "#ff0055", BarFilled: "#00ffff", BarEmpty: "#130924"}},
}

func (m *Model) renderAppearanceTab() string {
	var items []string

	items = append(items, "── PRESET PALETTES (Press Enter to apply preset) ──")
	for _, p := range themePresets {
		accentSwatch := lipgloss.NewStyle().Foreground(lipgloss.Color(p.theme.Accent)).Render("■")
		successSwatch := lipgloss.NewStyle().Foreground(lipgloss.Color(p.theme.Success)).Render("■")
		warnSwatch := lipgloss.NewStyle().Foreground(lipgloss.Color(p.theme.Warning)).Render("■")
		dangerSwatch := lipgloss.NewStyle().Foreground(lipgloss.Color(p.theme.Danger)).Render("■")
		items = append(items, fmt.Sprintf("%-32s %s%s%s%s", p.name, accentSwatch, successSwatch, warnSwatch, dangerSwatch))
	}

	items = append(items, "── GRANULAR THEME PALETTE (Press Enter to edit hex) ──")
	for _, f := range m.colorFields {
		val := f.get(&m.config.Theme)
		swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(val)).Render("■■■")
		items = append(items, fmt.Sprintf("%-16s %-10s %s", f.label+":", val, swatch))
	}

	return m.renderStructuredList(items)
}
