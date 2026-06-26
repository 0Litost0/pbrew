package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/pbrew/internal/adapter"
	"github.com/user/pbrew/internal/agent"
	"github.com/user/pbrew/internal/model"
	"github.com/user/pbrew/internal/platform"
	"github.com/user/pbrew/internal/registry"
	"github.com/user/pbrew/internal/source"
)

// InstallParams holds parameters for installing a component.
type InstallParams struct {
	ComponentName string
	Agent         string
	Scope         string
	Source        source.Source
	Force         bool
}

// Install installs a component to the target Agent directory.
func Install(params InstallParams) error {
	// 1. Validate agent
	if err := agent.Validate(params.Agent); err != nil {
		return err
	}
	profile, _ := agent.Get(params.Agent)

	// 2. Find and load component
	entry, err := registry.FindComponent(params.Source, params.ComponentName)
	if err != nil {
		return err
	}
	comp, files, err := registry.LoadComponent(params.Source, *entry)
	if err != nil {
		return fmt.Errorf("load component %s: %w", params.ComponentName, err)
	}

	// 3. Determine project root if needed
	var projectRoot string
	if params.Scope == "project" {
		projectRoot, err = platform.FindProjectRoot()
		if err != nil {
			return err
		}
	}

	// 4. Transform files
	sourceAgent := string(comp.SourceAgent)
	if sourceAgent == "" {
		sourceAgent = "claude" // Default
	}

	// Convert map[string][]byte to []CanonicalFile
	canonicalFiles := make([]adapter.CanonicalFile, 0, len(files))
	for name, content := range files {
		canonicalFiles = append(canonicalFiles, adapter.CanonicalFile{
			RelPath: name,
			Content: content,
		})
	}

	transformed, err := adapter.Transform(sourceAgent, params.Agent, canonicalFiles, adapter.CanonicalMeta{
		Name:        comp.Name,
		Version:     comp.Version,
		Type:        string(comp.Type),
		Description: comp.Description,
		SourceAgent: sourceAgent,
		Hash:        comp.Hash,
	})
	if err != nil {
		return fmt.Errorf("transform component: %w", err)
	}

	// 5. Calculate target directory
	targetDir, err := platform.InstallTargetDir(profile, params.Scope, projectRoot, string(comp.Type), comp.Name)
	if err != nil {
		return err
	}

	// 6. Check conflicts and write files
	installedFiles, err := writeFiles(targetDir, transformed, params.Force)
	if err != nil {
		return err
	}

	// 7. Update install record
	recordPath, err := platform.RecordFilePath(profile, params.Scope, projectRoot)
	if err != nil {
		return err
	}
	if err := updateRecord(recordPath, model.InstallRecord{
		Name:    comp.Name,
		Version: comp.Version,
		Type:    comp.Type,
		Agent:   params.Agent,
		Scope:   params.Scope,
		Source:  params.Source.Root(),
		Files:   installedFiles,
	}); err != nil {
		return fmt.Errorf("update install record: %w", err)
	}

	return nil
}

// writeFiles writes transformed files to the target directory.
// Returns the list of installed file relative paths for recording.
func writeFiles(targetDir string, files []adapter.CanonicalFile, force bool) ([]model.InstalledFile, error) {
	var installed []model.InstalledFile

	for _, f := range files {
		fullPath := filepath.Join(targetDir, f.RelPath)

		// Check if file already exists
		if _, err := os.Stat(fullPath); err == nil && !force {
			fmt.Printf("File already exists: %s\n", fullPath)
			fmt.Print("Overwrite? (y/N): ")
			var answer string
			fmt.Scanln(&answer)
			if answer != "y" && answer != "Y" {
				fmt.Printf("Skipping %s\n", f.RelPath)
				continue
			}
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return nil, fmt.Errorf("create directory: %w", err)
		}

		// Write file atomically
		if err := atomicWriteFile(fullPath, f.Content, 0644); err != nil {
			return nil, fmt.Errorf("write %s: %w", f.RelPath, err)
		}

		// Record relative path from Agent root
		relFromAgentRoot := filepath.Join(
			filepath.Base(filepath.Dir(targetDir)), // e.g., "rules" or "skills"
			filepath.Base(targetDir),               // e.g., "git-conventions"
			f.RelPath,                              // e.g., "git-conventions.md"
		)
		relFromAgentRoot = filepath.ToSlash(relFromAgentRoot)
		installed = append(installed, model.InstalledFile{RelPath: relFromAgentRoot})
	}

	return installed, nil
}

// updateRecord reads, updates, and writes the _config.json using atomic write.
func updateRecord(path string, record model.InstallRecord) error {
	var store model.RecordStore

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read record: %w", err)
	}
	if err == nil {
		if err := json.Unmarshal(data, &store); err != nil {
			return fmt.Errorf("parse record: %w", err)
		}
	}
	if store.Version == 0 {
		store.Version = 1
	}

	// Update or append record
	found := false
	for i, r := range store.Records {
		if r.Name == record.Name && r.Agent == record.Agent && r.Scope == record.Scope {
			store.Records[i] = record
			found = true
			break
		}
	}
	if !found {
		store.Records = append(store.Records, record)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create record directory: %w", err)
	}

	outData, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal record: %w", err)
	}

	return atomicWriteFile(path, outData, 0644)
}

// atomicWriteFile writes data to path atomically using write-temp-then-rename.
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
