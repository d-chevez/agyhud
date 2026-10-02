package payload

import (
	"encoding/json"
	"io"
)

// ModelInfo captures the active LLM attributes.
type ModelInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Effort      string `json:"effort"`
}

// WorkspaceInfo contains directory context paths.
type WorkspaceInfo struct {
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

// CurrentUsage tracks turn-specific token consumption.
type CurrentUsage struct {
	InputTokens              int64 `json:"input_tokens"`
	OutputTokens             int64 `json:"output_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
}

// ContextWindow captures token metrics and percentages.
type ContextWindow struct {
	TotalInputTokens    int64        `json:"total_input_tokens"`
	TotalOutputTokens   int64        `json:"total_output_tokens"`
	ContextWindowSize   int64        `json:"context_window_size"`
	UsedPercentage      float64      `json:"used_percentage"`
	RemainingPercentage float64      `json:"remaining_percentage"`
	CurrentUsage        CurrentUsage `json:"current_usage"`
}

// QuotaEntry represents a single rate-limit period.
type QuotaEntry struct {
	RemainingFraction float64 `json:"remaining_fraction"`
	ResetTime         string  `json:"reset_time"`
	ResetInSeconds    int64   `json:"reset_in_seconds"`
}

// SandboxInfo contains sandbox execution settings.
type SandboxInfo struct {
	Enabled bool `json:"enabled"`
}

// SessionPayload represents the complete JSON telemetry sent by Antigravity CLI.
type SessionPayload struct {
	CWD               string                `json:"cwd"`
	SessionID         string                `json:"session_id"`
	ConversationID    string                `json:"conversation_id"`
	ConversationTitle string                `json:"conversation_title"`
	TranscriptPath    string                `json:"transcript_path"`
	Model             ModelInfo             `json:"model"`
	Workspace         WorkspaceInfo         `json:"workspace"`
	Version           string                `json:"version"`
	ContextWindow     ContextWindow         `json:"context_window"`
	Exceeds200kTokens bool                  `json:"exceeds_200k_tokens"`
	Product           string                `json:"product"`
	Quota             map[string]QuotaEntry `json:"quota"`
	AgentState        string                `json:"agent_state"`
	Sandbox           SandboxInfo           `json:"sandbox"`
	PlanTier          string                `json:"plan_tier"`
	Email             string                `json:"email"`
	TerminalWidth     int                   `json:"terminal_width"`
}

// Parse reads JSON data from an io.Reader and unmarshals it into SessionPayload.
func Parse(r io.Reader) (*SessionPayload, error) {
	var p SessionPayload
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}
