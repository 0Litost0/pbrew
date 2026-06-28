package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/pbrew/internal/source"
)

func fixtureRepo(t *testing.T) string {
	t.Helper()
	repo, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixtures", "repo"))
	if err != nil {
		t.Fatal(err)
	}
	return repo
}

func TestListComponents_FromIndex(t *testing.T) {
	ls, err := source.NewLocalSource(fixtureRepo(t))
	if err != nil {
		t.Fatal(err)
	}

	components, err := ListComponents(ls)
	if err != nil {
		t.Fatalf("ListComponents error: %v", err)
	}
	if len(components) < 2 {
		t.Errorf("expected at least 2 components, got %d", len(components))
	}

	names := make(map[string]bool)
	for _, c := range components {
		names[c.Name] = true
	}
	if !names["git-conventions"] {
		t.Error("missing git-conventions component")
	}
	if !names["commit"] {
		t.Error("missing commit component")
	}
}

func TestListComponents_FromWalk(t *testing.T) {
	repo := t.TempDir()
	setupRepoWithoutIndex(t, repo)

	ls, err := source.NewLocalSource(repo)
	if err != nil {
		t.Fatal(err)
	}

	components, err := ListComponents(ls)
	if err != nil {
		t.Fatalf("ListComponents error: %v", err)
	}
	if len(components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(components))
	}
	if components[0].Name != "test-rule" {
		t.Errorf("component name: got %q, want %q", components[0].Name, "test-rule")
	}
}

func TestFindComponent(t *testing.T) {
	ls, err := source.NewLocalSource(fixtureRepo(t))
	if err != nil {
		t.Fatal(err)
	}

	entry, err := FindComponent(ls, "git-conventions")
	if err != nil {
		t.Fatalf("FindComponent error: %v", err)
	}
	if entry.Name != "git-conventions" {
		t.Errorf("Name: got %q, want %q", entry.Name, "git-conventions")
	}
	if entry.Type != "rule" {
		t.Errorf("Type: got %q, want %q", entry.Type, "rule")
	}
}

func TestFindComponent_NotFound(t *testing.T) {
	ls, err := source.NewLocalSource(fixtureRepo(t))
	if err != nil {
		t.Fatal(err)
	}

	_, err = FindComponent(ls, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent component")
	}
}

func TestLoadComponent(t *testing.T) {
	ls, err := source.NewLocalSource(fixtureRepo(t))
	if err != nil {
		t.Fatal(err)
	}

	entry, err := FindComponent(ls, "git-conventions")
	if err != nil {
		t.Fatal(err)
	}

	comp, files, err := LoadComponent(ls, *entry)
	if err != nil {
		t.Fatalf("LoadComponent error: %v", err)
	}
	if comp.Name != "git-conventions" {
		t.Errorf("Name: got %q, want %q", comp.Name, "git-conventions")
	}
	// .pbrew.json should be included in files (reserialized metadata)
	if _, ok := files[".pbrew.json"]; !ok {
		t.Error(".pbrew.json should be in returned files")
	}
	// git-conventions.md should be present
	if _, ok := files["git-conventions.md"]; !ok {
		t.Error("git-conventions.md should be in returned files")
	}
}

func TestLoadComponent_RuleFlatnessViolation(t *testing.T) {
	repo := t.TempDir()
	setupRuleWithSubdir(t, repo)

	ls, err := source.NewLocalSource(repo)
	if err != nil {
		t.Fatal(err)
	}

	components, err := ListComponents(ls)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = LoadComponent(ls, components[0])
	if err == nil {
		t.Error("expected error for Rule with nested subdirectory")
	}
}

func TestValidateRuleFlatness(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string][]byte
		wantErr bool
	}{
		{
			name: "flat_rule",
			files: map[string][]byte{
				"rule.md":     []byte("# Rule"),
				".pbrew.json": []byte("{}"),
			},
			wantErr: false,
		},
		{
			name: "nested_rule",
			files: map[string][]byte{
				"docs/extra.md": []byte("nested"),
				".pbrew.json":   []byte("{}"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRuleFlatness("test", tt.files)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRuleFlatness() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsHiddenPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{".pbrew.json", false},
		{"rules/test.md", false},
		{".git/config", true},
		{".hidden/file.md", true},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsHiddenPath(tt.path); got != tt.want {
				t.Errorf("IsHiddenPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// Test helpers

func setupRepoWithoutIndex(t *testing.T, dir string) {
	t.Helper()
	ruleDir := filepath.Join(dir, "rules", "test-rule")
	if err := os.MkdirAll(ruleDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(ruleDir, ".pbrew.json"), `{"name":"test-rule","version":"0.1.0","type":"rule","description":"A test rule","hash":"aaa"}`)
	writeTestFile(t, filepath.Join(ruleDir, "test-rule.md"), "# Test Rule\n")
}

func setupRuleWithSubdir(t *testing.T, dir string) {
	t.Helper()
	ruleDir := filepath.Join(dir, "rules", "bad-rule")
	docsDir := filepath.Join(ruleDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(ruleDir, ".pbrew.json"), `{"name":"bad-rule","version":"0.1.0","type":"rule","description":"A bad rule","hash":"bbb"}`)
	writeTestFile(t, filepath.Join(docsDir, "extra.md"), "# Extra\n")
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
