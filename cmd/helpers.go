package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/pbrew/internal/model"
	"github.com/user/pbrew/internal/platform"
)

// findProjectRoot wraps platform.FindProjectRoot with a friendlier error message.
func findProjectRoot() (string, error) {
	root, err := platform.FindProjectRoot()
	if err != nil {
		return "", fmt.Errorf("cannot determine project root (not in a git repository). Use --scope user instead")
	}
	return root, nil
}

// agentRoot returns the Agent root directory for the given profile and scope.
func agentRoot(profile model.AgentProfile, scope, projectRoot string) (string, error) {
	return platform.AgentRoot(profile, scope, projectRoot)
}

// recordFilePath returns the _config.json path for the given profile and scope.
func recordFilePath(profile model.AgentProfile, scope, projectRoot string) (string, error) {
	return platform.RecordFilePath(profile, scope, projectRoot)
}

// loadRecordStore reads _config.json, returning an empty store if the file doesn't exist.
func loadRecordStore(path string) (*model.RecordStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &model.RecordStore{Version: 1}, nil
		}
		return nil, fmt.Errorf("read record: %w", err)
	}
	var store model.RecordStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("parse record: %w", err)
	}
	return &store, nil
}

// saveRecordStore writes _config.json atomically.
func saveRecordStore(path string, store *model.RecordStore) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create record directory: %w", err)
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal record: %w", err)
	}
	return atomicWriteFile(path, data, 0644)
}

// atomicWriteFile is duplicated here to avoid circular import with installer package.
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".pbrew-tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, perm); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}
