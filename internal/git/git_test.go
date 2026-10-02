package git_test

import (
	"os"
	"testing"

	"github.com/d-chevez/agyhud/internal/git"
)

func TestGetInfoCurrentRepo(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get wd: %v", err)
	}

	info := git.GetInfo(wd, 5, 50)
	if !info.IsRepo {
		t.Errorf("Expected current workspace to be recognized as git repo")
	}
	if info.Branch != "release/v0.1.0" {
		t.Errorf("Expected branch 'release/v0.1.0', got '%s'", info.Branch)
	}

	// Test cache hit
	infoCached := git.GetInfo(wd, 5, 50)
	if infoCached.Branch != info.Branch {
		t.Errorf("Expected cached branch to match")
	}
}
