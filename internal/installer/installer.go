package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StateSnapshot records pre-installation statusLine configuration for safe rollback.
type StateSnapshot struct {
	StatusLineExisted  bool        `json:"statusLine_existed"`
	OriginalStatusLine interface{} `json:"original_statusLine"`
}

// HookStatus describes current integration state in settings.json.
type HookStatus struct {
	SettingsFound bool
	SettingsPath  string
	Active        bool
	Command       string
	Enabled       bool
	IsAgyhud      bool
}

// FindSettingsPath locates the active Antigravity CLI settings file.
func FindSettingsPath() (string, error) {
	if custom := os.Getenv("GEMINI_CLI_HOME"); custom != "" {
		return filepath.Join(custom, "settings.json"), nil
	}
	if custom := os.Getenv("ANTIGRAVITY_CONFIG_DIR"); custom != "" {
		return filepath.Join(custom, "settings.json"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to resolve user home directory: %w", err)
	}

	return filepath.Join(home, ".gemini", "antigravity-cli", "settings.json"), nil
}

// GetStatus inspects settings.json to determine if agyhud is configured.
func GetStatus() (HookStatus, error) {
	path, err := FindSettingsPath()
	if err != nil {
		return HookStatus{}, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return HookStatus{SettingsFound: false, SettingsPath: path}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return HookStatus{SettingsFound: true, SettingsPath: path}, err
	}

	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		return HookStatus{SettingsFound: true, SettingsPath: path}, err
	}

	st := HookStatus{SettingsFound: true, SettingsPath: path}
	rawStatus, ok := root["statusLine"].(map[string]interface{})
	if !ok {
		return st, nil
	}

	if cmd, ok := rawStatus["command"].(string); ok {
		st.Command = cmd
		if strings.Contains(strings.ToLower(cmd), "agyhud") {
			st.IsAgyhud = true
		}
	}
	if en, ok := rawStatus["enabled"].(bool); ok {
		st.Enabled = en
	}
	st.Active = st.Enabled && st.Command != ""

	return st, nil
}

// Install configures agyhud as the active statusLine hook in settings.json.
func Install(customBinary string, classic bool) (string, error) {
	settingsPath, err := FindSettingsPath()
	if err != nil {
		return "", err
	}

	settingsDir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	binPath := customBinary
	if binPath == "" {
		var execErr error
		binPath, execErr = os.Executable()
		if execErr != nil {
			return "", fmt.Errorf("failed to resolve current executable path: %w", execErr)
		}
	}
	binPath, _ = filepath.Abs(binPath)
	binPath = filepath.Clean(binPath)

	var commandStr string
	if strings.Contains(binPath, " ") {
		commandStr = fmt.Sprintf("\"%s\" render", binPath)
	} else {
		commandStr = fmt.Sprintf("%s render", binPath)
	}

	if classic {
		commandStr += " --classic"
	}

	// Read existing settings
	root := make(map[string]interface{})
	var originalStatusLine interface{}
	statusLineExisted := false

	if data, err := os.ReadFile(settingsPath); err == nil {
		_ = json.Unmarshal(data, &root)
		if orig, ok := root["statusLine"]; ok {
			statusLineExisted = true
			originalStatusLine = orig
		}
		// Create backup file if none exists
		bakPath := settingsPath + ".bak"
		if _, err := os.Stat(bakPath); os.IsNotExist(err) {
			_ = os.WriteFile(bakPath, data, 0644)
		}
	}

	// Save snapshot for atomic rollback
	snapshot := StateSnapshot{
		StatusLineExisted:  statusLineExisted,
		OriginalStatusLine: originalStatusLine,
	}
	snapshotPath := filepath.Join(settingsDir, "agyhud_installed_state.json")
	if snapBytes, err := json.MarshalIndent(snapshot, "", "  "); err == nil {
		_ = os.WriteFile(snapshotPath, snapBytes, 0644)
	}

	// Update statusLine configuration
	root["statusLine"] = map[string]interface{}{
		"type":    "command",
		"command": commandStr,
		"enabled": true,
	}

	// Write updated settings
	updatedBytes, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal settings JSON: %w", err)
	}

	if err := os.WriteFile(settingsPath, updatedBytes, 0644); err != nil {
		return "", fmt.Errorf("failed to write settings file: %w", err)
	}

	return commandStr, nil
}

// Uninstall restores the previous statusLine configuration safely.
func Uninstall() error {
	settingsPath, err := FindSettingsPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return nil
	}

	settingsDir := filepath.Dir(settingsPath)
	snapshotPath := filepath.Join(settingsDir, "agyhud_installed_state.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return err
	}

	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}

	// Check if state snapshot exists
	if snapData, err := os.ReadFile(snapshotPath); err == nil {
		var snap StateSnapshot
		if err := json.Unmarshal(snapData, &snap); err == nil {
			if snap.StatusLineExisted {
				root["statusLine"] = snap.OriginalStatusLine
			} else {
				delete(root, "statusLine")
			}
			_ = os.Remove(snapshotPath)
		}
	} else {
		// Fallback: remove or disable statusLine
		delete(root, "statusLine")
	}

	updatedBytes, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsPath, updatedBytes, 0644)
}
