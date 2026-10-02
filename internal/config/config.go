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

// ResponsiveConfig controls adaptive layouts based on terminal width.
type ResponsiveConfig struct {
	Enabled         bool `json:"enabled"`
	BreakpointWidth int  `json:"breakpoint_width"`
}

// WidgetConfig configures an individual widget instance in a row.
type WidgetConfig struct {
	Type      string            `json:"type"`      // e.g. "model", "git", "context_bar", "quota", "agent_state", "workspace"
	Enabled   bool              `json:"enabled"`   // whether this widget is rendered
	Padding   int               `json:"padding"`   // space padding around content
	Separator string            `json:"separator"` // separator character after widget
	Options   map[string]string `json:"options"`   // widget-specific overrides
}

// Config represents the complete root configuration for agyhud.
type Config struct {
	IconSet    IconSet          `json:"icon_set"`
	Theme      ThemeConfig      `json:"theme"`
	Git        GitConfig        `json:"git"`
	Responsive ResponsiveConfig `json:"responsive"`
	Rows       [][]WidgetConfig `json:"rows"`
}

// DefaultConfig returns the recommended modern production configuration.
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
			Enabled:         true,
			BreakpointWidth: 90,
		},
		Rows: [][]WidgetConfig{
			// Row 1: Workspace Context, Git Branch/Status, Agent Lifecycle State
			{
				{Type: "workspace", Enabled: true, Padding: 1, Separator: "│"},
				{Type: "git", Enabled: true, Padding: 1, Separator: "│"},
				{Type: "agent_state", Enabled: true, Padding: 1, Separator: ""},
			},
			// Row 2: Active Model, Context Window Bar, Quotas
			{
				{Type: "model", Enabled: true, Padding: 1, Separator: "│"},
				{Type: "context_bar", Enabled: true, Padding: 1, Separator: "│"},
				{Type: "quota", Enabled: true, Padding: 1, Separator: ""},
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

	// Validate bounds for Git configuration
	if cfg.Git.RefreshSeconds < 1 {
		cfg.Git.RefreshSeconds = 1
	} else if cfg.Git.RefreshSeconds > 10 {
		cfg.Git.RefreshSeconds = 10
	}
	if cfg.Git.TimeoutMs < 10 {
		cfg.Git.TimeoutMs = 10
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
