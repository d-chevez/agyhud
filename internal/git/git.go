package git

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Info holds Git telemetry for a workspace.
type Info struct {
	Branch  string `json:"branch"`
	IsDirty bool   `json:"is_dirty"`
	IsRepo  bool   `json:"is_repo"`
}

type cachedGit struct {
	Info      Info      `json:"info"`
	Timestamp time.Time `json:"timestamp"`
}

// GetInfo returns the Git status for the given directory using the hybrid strategy and TTL cache.
func GetInfo(dir string, refreshSec int, timeoutMs int) Info {
	if dir == "" {
		return Info{}
	}

	gitDir := findGitDir(dir)
	if gitDir == "" {
		return Info{IsRepo: false}
	}

	cacheFile := getCachePath(gitDir)

	// Check cache validity
	if cached, ok := readCache(cacheFile, refreshSec); ok {
		return cached
	}

	// 1. Ultra-fast direct file read for Branch (< 0.2ms)
	branch := readBranchFromHead(gitDir)
	if branch == "" {
		branch = "HEAD"
	}

	// 2. Controlled subprocess for dirty status with strict timeout
	isDirty := checkDirtyWithTimeout(dir, timeoutMs)

	info := Info{
		Branch:  branch,
		IsDirty: isDirty,
		IsRepo:  true,
	}

	// Persist to ephemeral cache
	writeCache(cacheFile, info)

	return info
}

func findGitDir(startDir string) string {
	curr := startDir
	for {
		gitPath := filepath.Join(curr, ".git")
		fi, err := os.Stat(gitPath)
		if err == nil {
			if fi.IsDir() {
				return gitPath
			}
			// Handle git submodules or worktrees
			content, err := os.ReadFile(gitPath)
			if err == nil {
				text := strings.TrimSpace(string(content))
				if strings.HasPrefix(text, "gitdir:") {
					relPath := strings.TrimSpace(strings.TrimPrefix(text, "gitdir:"))
					if filepath.IsAbs(relPath) {
						return relPath
					}
					return filepath.Clean(filepath.Join(curr, relPath))
				}
			}
		}

		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}
	return ""
}

func readBranchFromHead(gitDir string) string {
	headPath := filepath.Join(gitDir, "HEAD")
	data, err := os.ReadFile(headPath)
	if err != nil {
		return ""
	}

	line := strings.TrimSpace(string(data))
	if strings.HasPrefix(line, "ref: refs/heads/") {
		return strings.TrimPrefix(line, "ref: refs/heads/")
	}
	if len(line) >= 7 {
		return line[:7] // Detached commit
	}
	return line
}

func checkDirtyWithTimeout(dir string, timeoutMs int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "-C", dir, "status", "--porcelain", "-uno")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return false
	}

	return strings.TrimSpace(out.String()) != ""
}

func getCachePath(gitDir string) string {
	h := sha256.Sum256([]byte(gitDir))
	hashStr := hex.EncodeToString(h[:8])
	cacheDir := filepath.Join(os.TempDir(), "agyhud_cache")
	_ = os.MkdirAll(cacheDir, 0755)
	return filepath.Join(cacheDir, "git_"+hashStr+".json")
}

func readCache(cachePath string, refreshSec int) (Info, bool) {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return Info{}, false
	}

	var entry cachedGit
	if err := json.Unmarshal(data, &entry); err != nil {
		return Info{}, false
	}

	if time.Since(entry.Timestamp) <= time.Duration(refreshSec)*time.Second {
		return entry.Info, true
	}

	return Info{}, false
}

func writeCache(cachePath string, info Info) {
	entry := cachedGit{
		Info:      info,
		Timestamp: time.Now(),
	}
	if data, err := json.Marshal(entry); err == nil {
		_ = os.WriteFile(cachePath, data, 0644)
	}
}
