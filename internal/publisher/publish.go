package publisher

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/user/pbrew/internal/adapter"
	"github.com/user/pbrew/internal/agent"
	"github.com/user/pbrew/internal/model"
	"github.com/user/pbrew/internal/registry"
	"github.com/user/pbrew/internal/source"
)

// PublishResult holds the result of a successful publish operation.
type PublishResult struct {
	Name    string // Published component name
	Version string // Published version
	Path    string // Relative path in data source (e.g. "rules/my-rule")
}

// PublishParams holds the parameters for publishing a component.
type PublishParams struct {
	Name    string        // Component name (required)
	Dir     string        // Local directory containing component files (required)
	Version string        // Explicit version (optional; auto-incremented if empty)
	Source  source.Source // Configured data source
}

// componentNameRe defines the allowed characters for component names.
// Only lowercase letters, digits, hyphens, and underscores are permitted.
var componentNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

// Publish publishes a component to the data source.
func Publish(params PublishParams) (*PublishResult, error) {
	// 1. Validate component name (guard against path traversal)
	if !componentNameRe.MatchString(params.Name) {
		return nil, fmt.Errorf("invalid component name %q: must match %s (no path separators, uppercase, or special characters)", params.Name, componentNameRe.String())
	}

	// 2. Load the component directory
	localSrc, err := source.NewLocalSource(params.Dir)
	if err != nil {
		return nil, fmt.Errorf("open component directory: %w", err)
	}
	rawFiles, err := localSrc.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("read component directory: %w", err)
	}

	// 3. Validate directory is not empty (excluding .pbrew.json)
	nonMetaCount := 0
	for name := range rawFiles {
		if name != metaFile {
			nonMetaCount++
		}
	}
	if nonMetaCount == 0 {
		return nil, fmt.Errorf("component directory %q contains no files to publish", params.Dir)
	}

	// 4. Detect sourceAgent and component type
	srcAgent := detectSourceAgent(rawFiles)
	compType := detectComponentType(rawFiles)
	desc := detectDescription(rawFiles, compType)

	// 5. Remove .pbrew.json from the file set (it will be regenerated)
	delete(rawFiles, metaFile)

	// 6. Convert files to CanonicalFile slice
	canonicalFiles := make([]adapter.CanonicalFile, 0, len(rawFiles))
	for name, content := range rawFiles {
		canonicalFiles = append(canonicalFiles, adapter.CanonicalFile{
			RelPath: name,
			Content: content,
		})
	}

	// 7. Transform to repo default format (Claude)
	transformed, err := adapter.Transform(srcAgent, "claude", canonicalFiles, adapter.CanonicalMeta{
		Name:        params.Name,
		Type:        string(compType),
		SourceAgent: srcAgent,
	})
	if err != nil {
		return nil, fmt.Errorf("transform component: %w", err)
	}

	// 8. Convert transformed files back to map for hash and WriteDir
	transformedMap := make(map[string][]byte, len(transformed))
	for _, f := range transformed {
		transformedMap[f.RelPath] = f.Content
	}

	// 9. Determine version
	version, err := resolveVersion(params.Name, params.Version, params.Source)
	if err != nil {
		return nil, err
	}

	// 10. Compute hash
	hash := computeHash(transformedMap)

	// 11. Generate .pbrew.json metadata
	now := time.Now().UTC().Format(time.RFC3339)
	comp := model.Component{
		Name:        params.Name,
		Version:     version,
		Type:        compType,
		Description: desc,
		Hash:        hash,
		UpdatedAt:   now,
		SourceAgent: model.SourceAgent(srcAgent),
	}
	metaData, err := json.MarshalIndent(comp, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal .pbrew.json: %w", err)
	}
	transformedMap[metaFile] = metaData

	// 12. Determine write path in data source (data-driven via AgentProfile)
	claudeProfile, _ := agent.Get("claude")
	relPath := claudeProfile.ComponentDirs[string(compType)]
	if relPath == "" {
		return nil, fmt.Errorf("component type %q has no directory mapping for agent %q", compType, "claude")
	}
	relPath = relPath + "/" + params.Name // forward-slash for consistency

	// 13. Write component files to data source (atomically)
	if err := params.Source.WriteDir(relPath, transformedMap); err != nil {
		return nil, fmt.Errorf("write component to source: %w", err)
	}

	// 14. Update _index.json
	if err := updateIndex(params.Source, registry.ComponentEntry{
		Name:        params.Name,
		Type:        string(compType),
		Version:     version,
		Description: desc,
		Path:        relPath,
	}); err != nil {
		return nil, fmt.Errorf("update _index.json: %w", err)
	}

	return &PublishResult{
		Name:    params.Name,
		Version: version,
		Path:    relPath,
	}, nil
}

// resolveVersion determines the version for the component being published.
// If explicit version is provided, it is used directly (after checking for duplicates).
// Otherwise, it auto-increments: 0.1.0 for new components, patch+1 for existing.
// In all cases, if the resolved version already exists in the source, an error is returned.
func resolveVersion(name, explicitVersion string, src source.Source) (string, error) {
	existing, findErr := registry.FindComponent(src, name)
	componentExists := findErr == nil

	if explicitVersion != "" {
		// Validate the explicit version is valid semver
		if _, err := semver.NewVersion(explicitVersion); err != nil {
			return "", fmt.Errorf("invalid version %q: %w", explicitVersion, err)
		}
		// Check if same name+version already exists
		if componentExists && existing.Version == explicitVersion {
			return "", fmt.Errorf("component %q version %s already exists in source. Specify a different version with --version", name, explicitVersion)
		}
		return explicitVersion, nil
	}

	// Auto-increment
	if !componentExists {
		return "0.1.0", nil
	}

	// Bump patch version
	v, err := semver.NewVersion(existing.Version)
	if err != nil {
		return "", fmt.Errorf("parse existing version %q: %w", existing.Version, err)
	}
	newV := v.IncPatch()
	resolved := newV.String()

	// Guard: verify the auto-incremented version doesn't collide with an existing one
	// (possible if user previously published a higher version explicitly)
	check, checkErr := registry.FindComponent(src, name)
	if checkErr == nil && check.Version == resolved {
		return "", fmt.Errorf("auto-incremented version %s for component %q already exists in source. Specify a version explicitly with --version", resolved, name)
	}

	return resolved, nil
}

// updateIndex reads, updates, and writes _index.json in the data source root.
func updateIndex(src source.Source, entry registry.ComponentEntry) error {
	const indexFile = "_index.json"

	var idx registry.Index

	rootFiles, err := src.ReadDir(".")
	if err == nil {
		if data, ok := rootFiles[indexFile]; ok {
			parsed, parseErr := registry.ParseIndex(data)
			if parseErr == nil {
				idx = *parsed
			}
			// If parsing fails, start fresh — _index.json will be regenerated
		}
	}
	// If ReadDir fails (e.g. empty source), start with empty index

	// Update or append entry
	found := false
	for i, c := range idx.Components {
		if c.Name == entry.Name {
			idx.Components[i] = entry
			found = true
			break
		}
	}
	if !found {
		idx.Components = append(idx.Components, entry)
	}
	if idx.Version == 0 {
		idx.Version = 1
	}

	idxData, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal _index.json: %w", err)
	}

	// Write _index.json atomically
	files := map[string][]byte{
		indexFile: idxData,
	}
	return src.WriteDir(".", files)
}

// componentPathInSource returns the relative path for a component in the data source,
// driven by the target Agent's ComponentDirs profile.
func componentPathInSource(agentID, compType, name string) (string, error) {
	profile, ok := agent.Get(agentID)
	if !ok {
		return "", fmt.Errorf("unsupported agent: %s", agentID)
	}
	dir, ok := profile.ComponentDirs[compType]
	if !ok || dir == "" {
		return "", fmt.Errorf("component type %q has no directory mapping for agent %q", compType, agentID)
	}
	// Use forward slashes for cross-platform consistency
	return strings.Join([]string{dir, name}, "/"), nil
}
