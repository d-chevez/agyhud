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
	Merge        bool              `json:"merge,omitempty"`        // if true, suppress trailing space to merge seamlessly with next widget
	Bold         bool              `json:"bold,omitempty"`         // apply bold styling
	Color        string            `json:"color,omitempty"`        // custom color override (hex `#7aa2f7` or theme token)
	CustomSymbol string            `json:"custom_symbol,omitempty"` // symbol character for "custom_symbol" type
	Padding      int               `json:"padding,omitempty"`      // explicit padding
	Separator    string            `json:"separator,omitempty"`    // symbol for "separator" type
	Options      map[string]string `json:"options,omitempty"`      // widget-specific overrides
}

// Config represents the complete root configuration for agyhud.
type Config struct {
	IconSet    IconSet          `json:"icon_set"`
	Theme      ThemeConfig      `json:"theme"`
	Git        GitConfig        `json:"git"`
	Responsive ResponsiveConfig `json:"responsive"`
	Rows       [][]WidgetConfig `json:"rows"`
}

// DefaultConfig returns the recommended modern production configuration using atomic widgets.
func DefaultConfig() *Config {
	return &Config{
		IconSet: IconSetNerdFont,
		Theme: ThemeConfig{
			Accent:    "#7aa2f7", // Soft Blue
			Dim:       "#565f89", // Muted Gray
			Text:      "#c0caf5", // Clean Text
			Success:   "#9ece6a", // Lime Green
			Warning:   "#e0af68", // Warm Amber
			Danger:    "#f7768e", // Coral Red
			BarFilled: "#7aa2f7", // Accent Blue
			BarEmpty:  "#24283b", // Dark Charcoal
		},
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
			// Row 1: Workspace, Git branch with dirty status, Session title, Agent state
			{
				{Type: "custom_symbol", CustomSymbol: "󰉋", Enabled: true, Merge: false},
				{Type: "workspace", Label: "", RawValue: true, Enabled: true, Merge: false},
				{Type: "separator", Separator: "│", Enabled: true, Merge: false},
				{Type: "custom_symbol", CustomSymbol: "", Enabled: true, Merge: false},
				{Type: "git_branch", Label: "", RawValue: true, Enabled: true, Merge: true},
				{Type: "git_status", Enabled: true, Merge: false},
				{Type: "separator", Separator: "│", Enabled: true, Merge: false},
				{Type: "session_name", Label: "Session:", RawValue: false, Enabled: true, Merge: false},
				{Type: "separator", Separator: "│", Enabled: true, Merge: false},
				{Type: "agent_state", Enabled: true, Merge: false},
			},
			// Row 2: Model, thinking effort, context bar, tokens, 5h quota, reset timer
			{
				{Type: "custom_symbol", CustomSymbol: "󰚩", Enabled: true, Merge: false},
				{Type: "model", Label: "Model:", RawValue: false, Enabled: true, Merge: false},
				{Type: "thinking_effort", Enabled: true, Merge: false},
				{Type: "separator", Separator: "│", Enabled: true, Merge: false},
				{Type: "custom_symbol", CustomSymbol: "󱍏", Enabled: true, Merge: false},
				{Type: "context_bar", Label: "Context:", RawValue: false, Enabled: true, Merge: false},
				{Type: "context_percentage", RawValue: true, Enabled: true, Merge: false},
				{Type: "tokens_total", RawValue: true, Enabled: true, Merge: false},
				{Type: "separator", Separator: "│", Enabled: true, Merge: false},
				{Type: "custom_symbol", CustomSymbol: "󰥔", Enabled: true, Merge: false},
				{Type: "session_usage", Label: "5h:", RawValue: false, Enabled: true, Merge: false},
				{Type: "reset_timer", Enabled: true, Merge: false},
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

	return cfg, nil
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
