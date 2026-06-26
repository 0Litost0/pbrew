package platform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/pbrew/internal/model"
)

func TestFindProjectRoot(t *testing.T) {
	// Create a fake git repo structure
	repoDir := t.TempDir()
	gitDir := filepath.Join(repoDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a subdirectory to test upward search
	subDir := filepath.Join(repoDir, "src", "pkg")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Change cwd to subdirectory
	origDir, _ := os.Getwd()
	os.Chdir(subDir)
	defer os.Chdir(origDir)

	root, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot() error: %v", err)
	}
	if root != repoDir {
		t.Errorf("got %q, want %q", root, repoDir)
	}
}

func TestFindProjectRoot_NotGitRepo(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	_, err := FindProjectRoot()
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestAgentRoot(t *testing.T) {
	profile := model.AgentProfile{DirName: ".claude"}

	// User scope
	root, err := AgentRoot(profile, "user", "")
	if err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".claude")
	if root != expected {
		t.Errorf("user scope: got %q, want %q", root, expected)
	}

	// Project scope
	root, err = AgentRoot(profile, "project", "/my/project")
	if err != nil {
		t.Fatal(err)
	}
	if root != "/my/project/.claude" {
		t.Errorf("project scope: got %q, want %q", root, "/my/project/.claude")
	}

	// Invalid scope
	_, err = AgentRoot(profile, "invalid", "")
	if err == nil {
		t.Error("expected error for invalid scope")
	}
}

func TestInstallTargetDir(t *testing.T) {
	profile := model.AgentProfile{
		DirName: ".claude",
		ComponentDirs: map[string]string{
			"rule":  "rules",
			"skill": "skills",
		},
	}

	tests := []struct {
		componentType string
		componentName string
		want          string
	}{
		{"rule", "git-conventions", "/proj/.claude/rules/git-conventions"},
		{"skill", "commit", "/proj/.claude/skills/commit"},
	}

	for _, tt := range tests {
		t.Run(tt.componentType, func(t *testing.T) {
			got, err := InstallTargetDir(profile, "project", "/proj", tt.componentType, tt.componentName)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInstallTargetDir_UnknownType(t *testing.T) {
	profile := model.AgentProfile{
		DirName:        ".claude",
		ComponentDirs:  map[string]string{},
	}
	_, err := InstallTargetDir(profile, "project", "/proj", "unknown", "test")
	if err == nil {
		t.Error("expected error for unknown component type")
	}
}

func TestRecordFilePath(t *testing.T) {
	profile := model.AgentProfile{DirName: ".codebuddy"}
	path, err := RecordFilePath(profile, "project", "/proj")
	if err != nil {
		t.Fatal(err)
	}
	expected := "/proj/.codebuddy/.pbrew/_config.json"
	if path != expected {
		t.Errorf("got %q, want %q", path, expected)
	}
}
