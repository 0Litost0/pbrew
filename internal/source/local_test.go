package source

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalSourceReadDir(t *testing.T) {
	dir := t.TempDir()
	// Create test files
	if err := os.MkdirAll(filepath.Join(dir, "rules", "test-rule"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rules", "test-rule", "test.md"), []byte("# Test Rule"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rules", "test-rule", ".pbrew.json"), []byte(`{"name":"test-rule"}`), 0644); err != nil {
		t.Fatal(err)
	}

	ls, err := NewLocalSource(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer ls.Close()

	files, err := ls.ReadDir("rules/test-rule")
	if err != nil {
		t.Fatalf("ReadDir error: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
	if string(files["test.md"]) != "# Test Rule" {
		t.Errorf("test.md content mismatch: got %q", files["test.md"])
	}
}

func TestLocalSourceReadDir_NotFound(t *testing.T) {
	ls, err := NewLocalSource(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer ls.Close()

	_, err = ls.ReadDir("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestLocalSourceExists(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0644)

	ls, err := NewLocalSource(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer ls.Close()

	if !ls.Exists("a.txt") {
		t.Error("a.txt should exist")
	}
	if ls.Exists("b.txt") {
		t.Error("b.txt should not exist")
	}
}

func TestLocalSourceWriteDir(t *testing.T) {
	dir := t.TempDir()
	ls, err := NewLocalSource(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer ls.Close()

	files := map[string][]byte{
		"rule.md":     []byte("# Rule"),
		".pbrew.json": []byte(`{"name":"test"}`),
	}
	if err := ls.WriteDir("rules/test", files); err != nil {
		t.Fatalf("WriteDir error: %v", err)
	}

	// Verify written content
	got, err := os.ReadFile(filepath.Join(dir, "rules", "test", "rule.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "# Rule" {
		t.Errorf("rule.md: got %q, want %q", got, "# Rule")
	}
}

func TestLocalSourceRoot(t *testing.T) {
	ls, err := NewLocalSource(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer ls.Close()

	if ls.Root() == "" {
		t.Error("Root() should not be empty")
	}
}
