package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNonExistent(t *testing.T) {
	// Temporarily override home to a temp dir so config doesn't collide
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Source != "" {
		t.Errorf("expected empty Source, got %q", cfg.Source)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	cfg.Source = "https://github.com/test/repo.git"

	w := Wrap(cfg)
	if err := w.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify file exists
	expectedPath := filepath.Join(tmpDir, ".pbrew", "config.json")
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	// Reload and verify
	cfg2, err := Load()
	if err != nil {
		t.Fatalf("Load() after save error: %v", err)
	}
	if cfg2.Source != cfg.Source {
		t.Errorf("Source: got %q, want %q", cfg2.Source, cfg.Source)
	}
}

func TestSetAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	if err := Set("source", "/local/path"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	val, ok := Get(cfg, "source")
	if !ok {
		t.Error("Get(source) should return ok=true")
	}
	if val != "/local/path" {
		t.Errorf("Get(source): got %q, want %q", val, "/local/path")
	}
}

func TestSetUnknownKey(t *testing.T) {
	err := Set("nonexistent", "value")
	if err == nil {
		t.Error("Set with unknown key should return error")
	}
}

func TestGetUnknownKey(t *testing.T) {
	cfg, _ := Load()
	_, ok := Get(cfg, "nonexistent")
	if ok {
		t.Error("Get with unknown key should return ok=false")
	}
}

func TestAtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.json")

	data := []byte(`{"source":"test"}`)
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
	entries, _ := os.ReadDir(tmpDir)
	for _, e := range entries {
		if e.Name() != "test.json" {
			t.Errorf("unexpected file: %s", e.Name())
		}
	}
}
