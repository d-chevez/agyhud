package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// IconSet defines supported icon styles.
type IconSet string

const (
	IconSetNerdFont IconSet = "nerd_font"
	IconSetClassic  IconSet = "classic"
)

// LayoutMode controls whether widgets automatically wrap or follow rigid user rows.
type LayoutMode string

const (
	LayoutModeDynamic LayoutMode = "dynamic_wrap"
	LayoutModeManual  LayoutMode = "manual_rows"
)

// ThemeConfig defines the granular color palette used by lipgloss.
type ThemeConfig struct {
	Accent    string `json:"accent"`
	Dim       string `json:"dim"`
	Text      string `json:"text"`
	Success   string `json:"success"`
	Warning   string `json:"warning"`
	Danger    string `json:"danger"`
	BarFilled string `json:"bar_filled"`
	BarEmpty  string `json:"bar_empty"`
}

// GitConfig governs git inspection and caching behavior.
type GitConfig struct {
	RefreshSeconds int `json:"refresh_seconds"`
	TimeoutMs      int `json:"timeout_ms"`
}

// ResponsiveConfig controls adaptive layouts and terminal width wrapping.
type ResponsiveConfig struct {
	Mode            LayoutMode `json:"mode"`             // "dynamic_wrap" or "manual_rows"
	FullWidth       bool       `json:"full_width"`       // pad or right-align to full terminal width
	BreakpointWidth int        `json:"breakpoint_width"` // minimum fallback width
}

// WidgetConfig configures an atomic individual widget instance in a row.
type WidgetConfig struct {
	ID           string            `json:"id,omitempty"`
	Type         string            `json:"type"`                    // e.g. "custom_symbol", "model", "thinking_effort", "context_bar", etc.
	Enabled      bool              `json:"enabled"`                 // whether this widget is rendered
	Label        string            `json:"label,omitempty"`        // custom label prefix (e.g. "Model:", "Context:")
	RawValue     bool              `json:"raw_value,omitempty"`    // if true, omit label and render bare value
	RawPrefix    string            `json:"raw_prefix,omitempty"`   // opening character when raw (e.g. "[", "(")
	RawSuffix    string            `json:"raw_suffix,omitempty"`   // closing character when raw (e.g. "]", ")")
	Merge        bool              `json:"merge,omitempty"`        // if true, suppress trailing space to merge seamlessly with next widget
	Bold         bool              `json:"bold,omitempty"`         // apply bold styling
	Color        string            `json:"color,omitempty"`        // custom color override (hex `#7aa2f7` or theme token)
	CustomSymbol string            `json:"custom_symbol,omitempty"` // symbol character for "custom_symbol" type
	Padding      int               `json:"padding,omitempty"`      // explicit padding
	Separator    string            `json:"separator,omitempty"`    // symbol for "separator" type
	Options      map[string]string `json:"options,omitempty"`      // widget-specific overrides
}

// AntigravityDarkTheme is the default dark theme designed to match Google Antigravity CLI.
var AntigravityDarkTheme = ThemeConfig{
	Accent:    "#7aa2f7", // Soft Blue
	Dim:       "#565f89", // Muted Gray
	Text:      "#c0caf5", // Clean Text
	Success:   "#9ece6a", // Lime Green
	Warning:   "#e0af68", // Warm Amber
	Danger:    "#f7768e", // Coral Red
	BarFilled: "#7aa2f7", // Accent Blue
	BarEmpty:  "#24283b", // Dark Charcoal
}

// Config represents the complete root configuration for agyhud.
type Config struct {
	IconSet    IconSet          `json:"icon_set"`
	ThemeMode  string           `json:"theme_mode,omitempty"` // "default" or "custom"
	Theme      ThemeConfig      `json:"theme"`
	Git        GitConfig        `json:"git"`
	Responsive ResponsiveConfig `json:"responsive"`
	Rows       [][]WidgetConfig `json:"rows"`
}

// DefaultConfig returns the recommended modern production configuration using atomic widgets.
func DefaultConfig() *Config {
	return &Config{
		IconSet:   IconSetNerdFont,
		ThemeMode: "default",
		Theme:     AntigravityDarkTheme,
		Git: GitConfig{
			RefreshSeconds: 5,
			TimeoutMs:      25,
		},
		Responsive: ResponsiveConfig{
			Mode:            LayoutModeDynamic,
			FullWidth:       false,
			BreakpointWidth: 80,
		},
		Rows: [][]WidgetConfig{
			{
				{Type: "custom_symbol", CustomSymbol: "󰉋", Enabled: true},
				{Type: "workspace", Enabled: true},
				{Type: "separator", Separator: "│", Enabled: true},
				{Type: "custom_symbol", CustomSymbol: "", Enabled: true},
				{Type: "git_branch", Enabled: true, Merge: true},
				{Type: "git_status", Enabled: true},
				{Type: "separator", Separator: "│", Enabled: true},
				{Type: "session_name", Enabled: true},
				{Type: "separator", Separator: "│", Enabled: true},
				{Type: "agent_state", Enabled: true},
			},
			{
				{Type: "custom_symbol", CustomSymbol: "󰚩", Enabled: true},
				{Type: "model", Enabled: true},
				{Type: "thinking_effort", Enabled: true},
				{Type: "separator", Separator: "│", Enabled: true},
				{Type: "custom_symbol", CustomSymbol: "󱍏", Enabled: true},
				{Type: "context_bar", Enabled: true},
				{Type: "context_percentage", Enabled: true},
				{Type: "tokens_total", Enabled: true},
				{Type: "separator", Separator: "│", Enabled: true},
				{Type: "custom_symbol", CustomSymbol: "󰥔", Enabled: true},
				{Type: "session_usage", Enabled: true},
				{Type: "reset_timer", Enabled: true},
			},
		},
	}
}

// DefaultConfigPath resolves the standard cross-platform configuration file path.
func DefaultConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, hErr := os.UserHomeDir()
		if hErr != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configDir, "agyhud", "config.json"), nil
}

// Load reads the configuration from the given path or returns DefaultConfig if not found.
func Load(path string) (*Config, error) {
	if path == "" {
		var err error
		path, err = DefaultConfigPath()
		if err != nil {
			return DefaultConfig(), nil
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return DefaultConfig(), nil
	}

	if cfg.Git.RefreshSeconds < 1 {
		cfg.Git.RefreshSeconds = 1
	} else if cfg.Git.RefreshSeconds > 10 {
		cfg.Git.RefreshSeconds = 10
	}
	if cfg.Git.TimeoutMs < 10 {
		cfg.Git.TimeoutMs = 10
	}
	if cfg.Responsive.Mode == "" {
		cfg.Responsive.Mode = LayoutModeDynamic
	}
	cfg.Rows = NormalizeRows(cfg.Rows)

	return cfg, nil
}

// NormalizeRows flattens 2D rows into a single continuous sequence with row_break widgets.
func NormalizeRows(rows [][]WidgetConfig) [][]WidgetConfig {
	var flat []WidgetConfig
	for r, row := range rows {
		if r > 0 && len(flat) > 0 && len(row) > 0 {
			if flat[len(flat)-1].Type != "row_break" {
				flat = append(flat, WidgetConfig{Type: "row_break", Enabled: true})
			}
		}
		for _, w := range row {
			w.Enabled = true
			flat = append(flat, w)
		}
	}
	if len(flat) == 0 {
		return [][]WidgetConfig{{}}
	}
	return [][]WidgetConfig{flat}
}

// Save writes the configuration to disk formatted as indented JSON.
func Save(path string, cfg *Config) error {
	if path == "" {
		var err error
		path, err = DefaultConfigPath()
		if err != nil {
			return err
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
