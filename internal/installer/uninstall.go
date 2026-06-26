package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/pbrew/internal/agent"
	"github.com/user/pbrew/internal/model"
	"github.com/user/pbrew/internal/platform"
)

// UninstallParams holds parameters for uninstalling a component.
type UninstallParams struct {
	ComponentName string
	Agent         string
	Scope         string
}

// Uninstall removes a previously installed component.
func Uninstall(params UninstallParams) error {
	// 1. Validate agent
	if err := agent.Validate(params.Agent); err != nil {
		return err
	}
	profile, _ := agent.Get(params.Agent)

	// 2. Determine project root if needed
	var projectRoot string
	var err error
	if params.Scope == "project" {
		projectRoot, err = platform.FindProjectRoot()
		if err != nil {
			return err
		}
	}

	// 3. Read install record
	recordPath, err := platform.RecordFilePath(profile, params.Scope, projectRoot)
	if err != nil {
		return err
	}

	store, err := loadRecordStore(recordPath)
	if err != nil {
		return err
	}

	// 4. Find the record
	var record *model.InstallRecord
	var recordIdx int
	for i, r := range store.Records {
		if r.Name == params.ComponentName && r.Agent == params.Agent && r.Scope == params.Scope {
			record = &store.Records[i]
			recordIdx = i
			break
		}
	}
	if record == nil {
		return fmt.Errorf("component %s is not installed for agent %s in scope %s", params.ComponentName, params.Agent, params.Scope)
	}

	// 5. Remove files based on component type
	agentRoot, err := platform.AgentRoot(profile, params.Scope, projectRoot)
	if err != nil {
		return err
	}

	switch record.Type {
	case model.TypeRule:
		if err := removeRuleFiles(agentRoot, record); err != nil {
			return err
		}
	case model.TypeSkill:
		if err := removeSkillDir(agentRoot, record); err != nil {
			return err
		}
	}

	// 6. Remove record entry
	store.Records = append(store.Records[:recordIdx], store.Records[recordIdx+1:]...)

	// 7. Write back record store
	if err := saveRecordStore(recordPath, store); err != nil {
		return fmt.Errorf("update record: %w", err)
	}

	return nil
}

// removeRuleFiles deletes individual rule files and cleans up empty directories.
func removeRuleFiles(agentRoot string, record *model.InstallRecord) error {
	for _, f := range record.Files {
		fullPath := filepath.Join(agentRoot, f.RelPath)
		if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", fullPath, err)
		}
		// Clean up empty parent directories
		cleanEmptyDirs(filepath.Dir(fullPath))
	}
	return nil
}

// removeSkillDir deletes the entire skill installation subdirectory.
func removeSkillDir(agentRoot string, record *model.InstallRecord) error {
	if len(record.Files) == 0 {
		return nil
	}
	// Derive the skill directory from the first file path
	// e.g., "skills/commit/SKILL.md" → "skills/commit"
	firstRelPath := record.Files[0].RelPath
	skillDir := filepath.Join(agentRoot, filepath.Dir(firstRelPath))

	if err := os.RemoveAll(skillDir); err != nil {
		return fmt.Errorf("remove skill directory %s: %w", skillDir, err)
	}
	// Clean up empty parent directories
	cleanEmptyDirs(filepath.Dir(skillDir))
	return nil
}

// cleanEmptyDirs removes empty directories walking up from the given path.
func cleanEmptyDirs(dir string) {
	for {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			return
		}
		parent := filepath.Dir(dir)
		os.Remove(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
}

// loadRecordStore reads and parses _config.json.
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
