package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/pbrew/internal/model"
	"github.com/user/pbrew/internal/source"
)

const indexFile = "_index.json"
const metaFile = ".pbrew.json"

// ListComponents discovers available components from the data source.
// It first tries to read _index.json; falls back to recursive directory walk.
func ListComponents(src source.Source) ([]ComponentEntry, error) {
	if src.Exists(indexFile) {
		return listFromIndex(src)
	}
	return listFromWalk(src)
}

// FindComponent locates a single component by name.
func FindComponent(src source.Source, name string) (*ComponentEntry, error) {
	components, err := ListComponents(src)
	if err != nil {
		return nil, err
	}
	for _, c := range components {
		if c.Name == name {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("component not found: %s", name)
}

// LoadComponent reads the full component metadata and source files.
func LoadComponent(src source.Source, entry ComponentEntry) (*model.Component, map[string][]byte, error) {
	files, err := src.ReadDir(entry.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("read component dir %s: %w", entry.Path, err)
	}

	metaData, ok := files[metaFile]
	if !ok {
		return nil, nil, fmt.Errorf("component %s missing %s", entry.Name, metaFile)
	}

	var comp model.Component
	if err := json.Unmarshal(metaData, &comp); err != nil {
		return nil, nil, fmt.Errorf("parse %s for %s: %w", metaFile, entry.Name, err)
	}

	// Validate Rule component flatness
	if comp.Type == model.TypeRule {
		if err := validateRuleFlatness(entry.Name, files); err != nil {
			return nil, nil, err
		}
	}

	// Remove .pbrew.json from the file map (it's not installed)
	delete(files, metaFile)

	return &comp, files, nil
}

// listFromIndex reads _index.json and returns its component entries.
func listFromIndex(src source.Source) ([]ComponentEntry, error) {
	files, err := src.ReadDir(".")
	if err != nil {
		return nil, err
	}
	data, ok := files[indexFile]
	if !ok {
		return nil, fmt.Errorf("_index.json not found")
	}
	idx, err := ParseIndex(data)
	if err != nil {
		return nil, err
	}
	return idx.Components, nil
}

// listFromWalk recursively walks the source directory looking for .pbrew.json files.
func listFromWalk(src source.Source) ([]ComponentEntry, error) {
	var components []ComponentEntry
	root := src.Root()

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != metaFile {
			return nil
		}

		// Found a .pbrew.json — parse it for summary info
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		var comp model.Component
		if err := json.Unmarshal(data, &comp); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		compDir := filepath.Dir(path)
		relPath, err := filepath.Rel(root, compDir)
		if err != nil {
			return fmt.Errorf("relative path: %w", err)
		}
		relPath = filepath.ToSlash(relPath)

		components = append(components, ComponentEntry{
			Name:        comp.Name,
			Type:        string(comp.Type),
			Version:     comp.Version,
			Description: comp.Description,
			Path:        relPath,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return components, nil
}

// validateRuleFlatness ensures Rule component directories contain no subdirectories.
// Skill components are allowed to have subdirectories.
func validateRuleFlatness(name string, files map[string][]byte) error {
	for relPath := range files {
		dir := filepath.Dir(relPath)
		if dir != "." && dir != "" {
			return fmt.Errorf("rule component %q violates flatness constraint: contains subdirectory %q (file %q). Rule components must not nest subdirectories",
				name, dir, relPath)
		}
	}
	return nil
}

// IsIndexPath checks if a relative path points to the _index.json file.
func IsIndexPath(relPath string) bool {
	return filepath.ToSlash(relPath) == indexFile
}

// IsHiddenPath checks if a relative path starts with a dot-prefix directory
// (excluding .pbrew.json which is a valid metadata file).
func IsHiddenPath(relPath string) bool {
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	for _, p := range parts {
		if strings.HasPrefix(p, ".") && p != metaFile {
			return true
		}
	}
	return false
}
