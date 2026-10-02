package tui

import (
	"github.com/d-chevez/agyhud/internal/payload"
)

// GetSamplePayload returns a realistic Antigravity session payload for the TUI live preview.
func GetSamplePayload() *payload.SessionPayload {
	return &payload.SessionPayload{
		CWD:               "C:\\Develop\\Personal\\agyhud",
		SessionID:         "demo-session-preview",
		ConversationID:    "demo-conversation",
		ConversationTitle: "Interactive Statusline Preview",
		Model: payload.ModelInfo{
			ID:          "gemini-3.8-flash",
			DisplayName: "Gemini 3.8 Flash",
			Effort:      "high",
		},
		Workspace: payload.WorkspaceInfo{
			CurrentDir: "agyhud",
			ProjectDir: "agyhud",
		},
		Version: "1.2.6",
		ContextWindow: payload.ContextWindow{
			TotalInputTokens:    186520,
			TotalOutputTokens:   76141,
			ContextWindowSize:   1048576,
			UsedPercentage:      25.0,
			RemainingPercentage: 75.0,
			CurrentUsage: payload.CurrentUsage{
				InputTokens:  3018,
				OutputTokens: 531,
			},
		},
		AgentState: "working",
		Quota: map[string]payload.QuotaEntry{
			"gemini-5h": {
				RemainingFraction: 0.88,
				ResetInSeconds:    13837,
			},
		},
		TerminalWidth: 120,
	}
}
