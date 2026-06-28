package publisher

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/pbrew/internal/model"
	"github.com/user/pbrew/internal/registry"
	"github.com/user/pbrew/internal/source"
)

func TestComputeHash(t *testing.T) {
	files := map[string][]byte{
		"a.md": []byte("hello"),
		"b.md": []byte("world"),
	}
	h1 := computeHash(files)

	// Same content in different order should produce same hash
	files2 := map[string][]byte{
		"b.md": []byte("world"),
		"a.md": []byte("hello"),
	}
	h2 := computeHash(files2)

	if h1 != h2 {
		t.Errorf("hash should be deterministic regardless of map iteration order")
	}

	// Different content should produce different hash
	files3 := map[string][]byte{
		"a.md": []byte("hello"),
		"b.md": []byte("different"),
	}
	h3 := computeHash(files3)
	if h1 == h3 {
		t.Error("different content should produce different hash")
	}
}

func TestDetectComponentType(t *testing.T) {
	tests := []struct {
		name string
		files map[string][]byte
		want model.ComponentType
	}{
		{
			name: "skill_has_skill_md",
			files: map[string][]byte{
				"SKILL.md": []byte("---\n---\n# Skill"),
				"scripts/hello.sh": []byte("#!/bin/bash"),
			},
			want: model.TypeSkill,
		},
		{
			name: "rule_no_skill_md",
			files: map[string][]byte{
				"my-rule.md": []byte("# My Rule"),
			},
			want: model.TypeRule,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectComponentType(tt.files)
			if got != tt.want {
				t.Errorf("detectComponentType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectSourceAgent(t *testing.T) {
	tests := []struct {
		name string
		files map[string][]byte
		want string
	}{
		{
			name: "from_pbrew_json",
			files: map[string][]byte{
				".pbrew.json": []byte(`{"name":"x","sourceAgent":"codebuddy"}`),
				"rule.md":     []byte("# Rule"),
			},
			want: "codebuddy",
		},
		{
			name: "from_content_codebuddy",
			files: map[string][]byte{
				"rule.md": []byte("See @.codebuddy/rules/api.md\n${CODEBUDDY_SKILL_DIR}/run.sh"),
			},
			want: "codebuddy",
		},
		{
			name: "default_claude",
			files: map[string][]byte{
				"rule.md": []byte("# Simple rule"),
			},
			want: "claude",
		},
		{
			name: "pbrew_json_takes_priority",
			files: map[string][]byte{
				".pbrew.json": []byte(`{"name":"x","sourceAgent":"claude"}`),
				"rule.md":     []byte("See @.codebuddy/rules/api.md"),
			},
			want: "claude",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectSourceAgent(tt.files)
			if got != tt.want {
				t.Errorf("detectSourceAgent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectDescription(t *testing.T) {
	tests := []struct {
		name string
		files map[string][]byte
		compType model.ComponentType
		want string
	}{
		{
			name: "skill_from_frontmatter",
			files: map[string][]byte{
				"SKILL.md": []byte("---\ndescription: My skill desc\n---\n# Skill"),
			},
			compType: model.TypeSkill,
			want:     "My skill desc",
		},
		{
			name: "rule_from_frontmatter",
			files: map[string][]byte{
				"rule.md": []byte("---\ndescription: Rule description\nalwaysApply: true\n---\n# Rule"),
			},
			compType: model.TypeRule,
			want:     "Rule description",
		},
		{
			name: "no_frontmatter",
			files: map[string][]byte{
				"rule.md": []byte("# Plain rule"),
			},
			compType: model.TypeRule,
			want:     "",
		},
		{
			name: "no_description_field",
			files: map[string][]byte{
				"rule.md": []byte("---\npaths: src/**/*.ts\n---\n# Rule"),
			},
			compType: model.TypeRule,
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectDescription(tt.files, tt.compType)
			if got != tt.want {
				t.Errorf("detectDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveVersion(t *testing.T) {
	// Create a mock source with a Rule component
	dir := t.TempDir()
	setupMockRepo(dir, "existing-rule", "1.2.3", "rule", "An existing rule")

	src, err := source.NewLocalSource(dir)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		compName string
		explicit string
		want    string
		wantErr bool
	}{
		{
			name:     "new_component_default",
			compName: "new-rule",
			explicit: "",
			want:     "0.1.0",
		},
		{
			name:     "existing_component_auto_bump",
			compName: "existing-rule",
			explicit: "",
			want:     "1.2.4",
		},
		{
			name:     "explicit_new_version",
			compName: "existing-rule",
			explicit: "2.0.0",
			want:     "2.0.0",
		},
		{
			name:     "same_version_fails",
			compName: "existing-rule",
			explicit: "1.2.3",
			wantErr:  true,
		},
		{
			name:     "invalid_version_fails",
			compName: "new-rule",
			explicit: "not-a-version",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveVersion(tt.compName, tt.explicit, src)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("resolveVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPublishNewComponent(t *testing.T) {
	// Set up an empty data source
	repoDir := t.TempDir()
	// Create an empty _index.json
	idx := registry.Index{Version: 1, Components: nil}
	idxData, _ := json.MarshalIndent(idx, "", "  ")
	os.MkdirAll(repoDir, 0755)
	os.WriteFile(filepath.Join(repoDir, "_index.json"), idxData, 0644)

	src, err := source.NewLocalSource(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Create a local component directory with a Rule
	compDir := t.TempDir()
	ruleContent := "---\npaths: src/**/*.ts\n---\n# My Rule\n\nFollow these conventions.\n"
	os.WriteFile(filepath.Join(compDir, "my-rule.md"), []byte(ruleContent), 0644)

	_, err = Publish(PublishParams{
		Name:   "my-rule",
		Dir:    compDir,
		Source: src,
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	// Verify files were written
	rulePath := filepath.Join(repoDir, "rules", "my-rule", "my-rule.md")
	if _, err := os.Stat(rulePath); err != nil {
		t.Errorf("expected rule file at %s: %v", rulePath, err)
	}

	metaPath := filepath.Join(repoDir, "rules", "my-rule", ".pbrew.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("expected .pbrew.json: %v", err)
	}

	var comp model.Component
	if err := json.Unmarshal(metaData, &comp); err != nil {
		t.Fatalf("parse .pbrew.json: %v", err)
	}
	if comp.Name != "my-rule" {
		t.Errorf("name = %q, want %q", comp.Name, "my-rule")
	}
	if comp.Version != "0.1.0" {
		t.Errorf("version = %q, want %q", comp.Version, "0.1.0")
	}
	if comp.Type != model.TypeRule {
		t.Errorf("type = %q, want %q", comp.Type, model.TypeRule)
	}
	if comp.Hash == "" {
		t.Error("hash should not be empty")
	}

	// Verify _index.json was updated
	idxData2, _ := os.ReadFile(filepath.Join(repoDir, "_index.json"))
	var idx2 registry.Index
	if err := json.Unmarshal(idxData2, &idx2); err != nil {
		t.Fatalf("parse _index.json: %v", err)
	}
	found := false
	for _, c := range idx2.Components {
		if c.Name == "my-rule" {
			found = true
			if c.Version != "0.1.0" {
				t.Errorf("index version = %q, want %q", c.Version, "0.1.0")
			}
		}
	}
	if !found {
		t.Error("my-rule not found in _index.json")
	}
}

func TestPublishExistingComponentAutoBump(t *testing.T) {
	repoDir := t.TempDir()
	setupMockRepo(repoDir, "my-skill", "0.5.2", "skill", "A skill")

	src, err := source.NewLocalSource(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Create a local Skill directory
	compDir := t.TempDir()
	skillContent := "---\ndescription: A skill\nallowed-tools: [Bash]\n---\n# My Skill\n"
	os.WriteFile(filepath.Join(compDir, "SKILL.md"), []byte(skillContent), 0644)

	_, err = Publish(PublishParams{
		Name:   "my-skill",
		Dir:    compDir,
		Source: src,
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	// Verify version was bumped
	metaPath := filepath.Join(repoDir, "skills", "my-skill", ".pbrew.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read .pbrew.json: %v", err)
	}
	var comp model.Component
	if err := json.Unmarshal(metaData, &comp); err != nil {
		t.Fatalf("parse .pbrew.json: %v", err)
	}
	if comp.Version != "0.5.3" {
		t.Errorf("version = %q, want %q", comp.Version, "0.5.3")
	}
}

func TestPublishSameVersionFails(t *testing.T) {
	repoDir := t.TempDir()
	setupMockRepo(repoDir, "my-rule", "1.0.0", "rule", "A rule")

	src, err := source.NewLocalSource(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	compDir := t.TempDir()
	os.WriteFile(filepath.Join(compDir, "my-rule.md"), []byte("# Rule"), 0644)

	_, err = Publish(PublishParams{
		Name:    "my-rule",
		Dir:     compDir,
		Version: "1.0.0",
		Source:  src,
	})
	if err == nil {
		t.Fatal("expected error when publishing same version, got nil")
	}
}

func TestPublishWithExplicitVersion(t *testing.T) {
	repoDir := t.TempDir()
	setupMockRepo(repoDir, "my-rule", "1.0.0", "rule", "A rule")

	src, err := source.NewLocalSource(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	compDir := t.TempDir()
	os.WriteFile(filepath.Join(compDir, "my-rule.md"), []byte("# Rule"), 0644)

	_, err = Publish(PublishParams{
		Name:    "my-rule",
		Dir:     compDir,
		Version: "2.0.0",
		Source:  src,
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	metaPath := filepath.Join(repoDir, "rules", "my-rule", ".pbrew.json")
	metaData, _ := os.ReadFile(metaPath)
	var comp model.Component
	json.Unmarshal(metaData, &comp)
	if comp.Version != "2.0.0" {
		t.Errorf("version = %q, want %q", comp.Version, "2.0.0")
	}
}

// --- helpers ---

// setupMockRepo creates a minimal data source directory with a component and _index.json.
func setupMockRepo(dir, name, version, compType, desc string) {
	typeDir := compType + "s" // "rules" or "skills"
	compDir := filepath.Join(dir, typeDir, name)
	os.MkdirAll(compDir, 0755)

	// Create .pbrew.json
	comp := model.Component{
		Name:        name,
		Version:     version,
		Type:        model.ComponentType(compType),
		Description: desc,
		Hash:        "existinghash",
		UpdatedAt:   "2025-01-01T00:00:00Z",
		SourceAgent: model.SourceClaude,
	}
	metaData, _ := json.MarshalIndent(comp, "", "  ")
	os.WriteFile(filepath.Join(compDir, ".pbrew.json"), metaData, 0644)

	// Create a dummy content file
	if compType == "skill" {
		os.WriteFile(filepath.Join(compDir, "SKILL.md"), []byte("---\ndescription: "+desc+"\n---\n# "+name), 0644)
	} else {
		os.WriteFile(filepath.Join(compDir, name+".md"), []byte("# "+name+"\n"+desc), 0644)
	}

	// Create _index.json
	path := typeDir + "/" + name
	idx := registry.Index{
		Version: 1,
		Components: []registry.ComponentEntry{
			{Name: name, Type: compType, Version: version, Description: desc, Path: path},
		},
	}
	idxData, _ := json.MarshalIndent(idx, "", "  ")
	os.WriteFile(filepath.Join(dir, "_index.json"), idxData, 0644)
}

// setupMockRepo creates a component with .pbrew.json but no _index.json variant.
// Used by TestResolveVersion which only needs registry.FindComponent to work.
func setupMockRepoNoIndex(dir, name, version, compType, desc string) {
	typeDir := compType + "s"
	compDir := filepath.Join(dir, typeDir, name)
	os.MkdirAll(compDir, 0755)

	comp := model.Component{
		Name:        name,
		Version:     version,
		Type:        model.ComponentType(compType),
		Description: desc,
		Hash:        "hash",
	}
	metaData, _ := json.MarshalIndent(comp, "", "  ")
	os.WriteFile(filepath.Join(compDir, ".pbrew.json"), metaData, 0644)

	if compType == "skill" {
		os.WriteFile(filepath.Join(compDir, "SKILL.md"), []byte("# Skill"), 0644)
	} else {
		os.WriteFile(filepath.Join(compDir, name+".md"), []byte("# Rule"), 0644)
	}
}
