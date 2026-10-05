package installer_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/d-chevez/agyhud/internal/installer"
)

func TestInstallAndUninstallCycle(t *testing.T) {
	tempHome, err := os.MkdirTemp("", "agyhud-install-test")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)

	// Direct installer to temp directory via environment override
	settingsDir := filepath.Join(tempHome, ".gemini", "antigravity-cli")
	_ = os.MkdirAll(settingsDir, 0755)
	t.Setenv("GEMINI_CLI_HOME", settingsDir)
	t.Setenv("APPDATA", tempHome)
	t.Setenv("XDG_CONFIG_HOME", tempHome)

	settingsFile := filepath.Join(settingsDir, "settings.json")
	initialContent := []byte(`{ "colorScheme": "tokyo night" }`)
	if err := os.WriteFile(settingsFile, initialContent, 0644); err != nil {
		t.Fatalf("Failed to write initial settings: %v", err)
	}

	// 1. Install
	mockBin := "C:/Develop/Personal/agyhud/agyhud.exe"
	cmdStr, err := installer.Install(mockBin, false)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if !strings.Contains(cmdStr, "render") {
		t.Errorf("Expected command string to contain 'render', got %s", cmdStr)
	}

	// Verify default config was created
	expectedCfg := filepath.Join(tempHome, "agyhud", "config.json")
	if _, err := os.Stat(expectedCfg); os.IsNotExist(err) {
		t.Errorf("Expected default config to be created at %s", expectedCfg)
	}

	// Verify settings.json updated
	status, err := installer.GetStatus()
	if err != nil {
		t.Fatalf("Failed to check status: %v", err)
	}
	if !status.IsAgyhud || !status.Active {
		t.Errorf("Expected agyhud to be active after install")
	}

	// 2. Uninstall
	if err := installer.Uninstall(); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	// Verify statusLine removed cleanly and original settings preserved
	uninstalledData, _ := os.ReadFile(settingsFile)
	var root map[string]interface{}
	_ = json.Unmarshal(uninstalledData, &root)

	if _, ok := root["statusLine"]; ok {
		t.Errorf("Expected statusLine to be removed after uninstall")
	}
	if root["colorScheme"] != "tokyo night" {
		t.Errorf("Expected colorScheme 'tokyo night' to be preserved")
	}
}
