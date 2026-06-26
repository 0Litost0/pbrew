package installer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/pbrew/internal/model"
)

func TestAtomicWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	data := []byte(`{"key":"value"}`)

	if err := atomicWriteFile(path, data, 0644); err != nil {
		t.Fatalf("atomicWriteFile: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("content mismatch: got %q, want %q", got, data)
	}

	// No leftover temp files
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.Name() != "test.json" {
			t.Errorf("unexpected file: %s", e.Name())
		}
	}
}

func TestAtomicWriteFile_Overwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")

	if err := atomicWriteFile(path, []byte("first"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := atomicWriteFile(path, []byte("second"), 0644); err != nil {
		t.Fatal(err)
	}

	got, _ := os.ReadFile(path)
	if string(got) != "second" {
		t.Errorf("got %q, want %q", got, "second")
	}
}

func TestUpdateRecord(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".pbrew", "_config.json")

	r1 := model.InstallRecord{
		Name:    "rule-a",
		Version: "1.0.0",
		Type:    model.TypeRule,
		Agent:   "codebuddy",
		Scope:   "project",
		Source:  "test",
		Files:   []model.InstalledFile{{RelPath: "rules/rule-a/rule-a.md"}},
	}

	if err := updateRecord(path, r1); err != nil {
		t.Fatalf("updateRecord first: %v", err)
	}

	// Read back
	store, err := loadRecordStore(path)
	if err != nil {
		t.Fatalf("loadRecordStore: %v", err)
	}
	if len(store.Records) != 1 || store.Records[0].Name != "rule-a" {
		t.Errorf("unexpected records: %+v", store.Records)
	}

	// Update same record (new version)
	r2 := model.InstallRecord{
		Name:    "rule-a",
		Version: "1.1.0",
		Type:    model.TypeRule,
		Agent:   "codebuddy",
		Scope:   "project",
		Source:  "test",
		Files:   []model.InstalledFile{{RelPath: "rules/rule-a/rule-a.md"}},
	}
	if err := updateRecord(path, r2); err != nil {
		t.Fatalf("updateRecord second: %v", err)
	}

	store2, _ := loadRecordStore(path)
	if len(store2.Records) != 1 {
		t.Errorf("expected 1 record, got %d", len(store2.Records))
	}
	if store2.Records[0].Version != "1.1.0" {
		t.Errorf("version: got %q, want %q", store2.Records[0].Version, "1.1.0")
	}

	// Add a second record
	r3 := model.InstallRecord{
		Name:    "skill-b",
		Version: "0.1.0",
		Type:    model.TypeSkill,
		Agent:   "codebuddy",
		Scope:   "project",
		Source:  "test",
	}
	if err := updateRecord(path, r3); err != nil {
		t.Fatalf("updateRecord third: %v", err)
	}

	store3, _ := loadRecordStore(path)
	if len(store3.Records) != 2 {
		t.Errorf("expected 2 records, got %d", len(store3.Records))
	}
}

func TestSemverCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"1.2.0", "1.1.9", 1},
		{"0.1.0", "0.2.0", -1},
	}
	for _, tt := range tests {
		result, err := semverCompare(tt.a, tt.b)
		if err != nil {
			t.Errorf("semverCompare(%q, %q): unexpected error: %v", tt.a, tt.b, err)
			continue
		}
		if result != tt.want {
			t.Errorf("semverCompare(%q, %q) = %d, want %d", tt.a, tt.b, result, tt.want)
		}
	}
}

func TestCleanEmptyDirs(t *testing.T) {
	dir := t.TempDir()
	// Create a nested empty directory structure
	inner := filepath.Join(dir, "a", "b", "c")
	os.MkdirAll(inner, 0755)

	cleanEmptyDirs(filepath.Join(dir, "a", "b", "c"))

	// All empty dirs should be cleaned
	if _, err := os.Stat(filepath.Join(dir, "a")); err == nil {
		t.Error("expected 'a' directory to be removed")
	}
}

func TestCleanEmptyDirs_StopsAtNonEmpty(t *testing.T) {
	dir := t.TempDir()
	inner := filepath.Join(dir, "a", "b", "c")
	os.MkdirAll(inner, 0755)
	// Place a file in 'a' to prevent it from being removed
	os.WriteFile(filepath.Join(dir, "a", "keep.txt"), []byte("data"), 0644)

	cleanEmptyDirs(filepath.Join(dir, "a", "b", "c"))

	// 'a' should survive (has keep.txt), but 'b' should be gone
	if _, err := os.Stat(filepath.Join(dir, "a")); err != nil {
		t.Error("'a' directory should still exist (non-empty)")
	}
	if _, err := os.Stat(filepath.Join(dir, "a", "b")); err == nil {
		t.Error("'b' directory should be removed")
	}
}

func TestSaveAndLoadRecordStore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "_config.json")

	store := &model.RecordStore{
		Version: 1,
		Records: []model.InstallRecord{
			{Name: "test", Version: "0.1.0", Type: model.TypeRule, Agent: "claude", Scope: "project", Source: "local"},
		},
	}

	if err := saveRecordStore(path, store); err != nil {
		t.Fatal(err)
	}

	// Verify JSON structure
	data, _ := os.ReadFile(path)
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)
	if raw["version"] != float64(1) {
		t.Errorf("version field: got %v, want 1", raw["version"])
	}

	// Load back
	got, err := loadRecordStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Records) != 1 || got.Records[0].Name != "test" {
		t.Errorf("unexpected loaded records: %+v", got.Records)
	}
}
