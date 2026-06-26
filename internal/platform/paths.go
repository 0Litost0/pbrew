package platform

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/pbrew/internal/model"
)

// FindProjectRoot walks up from the current working directory to find the
// nearest directory containing a .git/ entry, which marks the project root.
func FindProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	dir := cwd
	for {
		if info, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			if info.IsDir() {
				return dir, nil
			}
			// .git could be a file (worktree), still valid
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not a git repository (or any parent): %s", cwd)
		}
		dir = parent
	}
}

// AgentRoot returns the root directory for a given Agent and scope.
// For user scope: ~/.<agent-dir>/
// For project scope: <project-root>/<agent-dir>/
func AgentRoot(profile model.AgentProfile, scope, projectRoot string) (string, error) {
	switch scope {
	case "user":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get home directory: %w", err)
		}
		return filepath.Join(home, profile.DirName), nil
	case "project":
		if projectRoot == "" {
			return "", fmt.Errorf("project root is required for project scope")
		}
		return filepath.Join(projectRoot, profile.DirName), nil
	default:
		return "", fmt.Errorf("unsupported scope: %s", scope)
	}
}

// InstallTargetDir returns the installation target directory for a component type.
// For rule: <agent-root>/rules/<component-name>/
// For skill: <agent-root>/skills/<component-name>/
func InstallTargetDir(profile model.AgentProfile, scope, projectRoot, componentType, componentName string) (string, error) {
	root, err := AgentRoot(profile, scope, projectRoot)
	if err != nil {
		return "", err
	}

	subDir, ok := profile.ComponentDirs[componentType]
	if !ok || subDir == "" {
		return "", fmt.Errorf("component type %q has no directory mapping for agent %q", componentType, profile.ID)
	}

	return filepath.Join(root, subDir, componentName), nil
}

// RecordFilePath returns the path to the _config.json installation record file.
// Always: <agent-root>/.pbrew/_config.json
func RecordFilePath(profile model.AgentProfile, scope, projectRoot string) (string, error) {
	root, err := AgentRoot(profile, scope, projectRoot)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, ".pbrew", "_config.json"), nil
}
