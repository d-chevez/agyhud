package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/d-chevez/agyhud/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.IconSet != config.IconSetNerdFont {
		t.Errorf("Expected IconSetNerdFont, got %s", cfg.IconSet)
	}
	if len(cfg.Rows) != 2 {
		t.Errorf("Expected 2 rows in default layout, got %d", len(cfg.Rows))
	}
	if cfg.Git.RefreshSeconds != 5 {
		t.Errorf("Expected 5 seconds git refresh, got %d", cfg.Git.RefreshSeconds)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agyhud-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	cfg := config.DefaultConfig()
	cfg.IconSet = config.IconSetClassic
	cfg.Git.RefreshSeconds = 8

	if err := config.Save(configPath, cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loaded, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loaded.IconSet != config.IconSetClassic {
		t.Errorf("Expected IconSetClassic, got %s", loaded.IconSet)
	}
	if loaded.Git.RefreshSeconds != 8 {
		t.Errorf("Expected 8s refresh, got %d", loaded.Git.RefreshSeconds)
	}
}
