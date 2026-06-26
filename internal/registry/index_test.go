package registry

import (
	"testing"
)

func TestParseIndex(t *testing.T) {
	data := []byte(`{
		"version": 1,
		"components": [
			{"name":"git-conventions","type":"rule","version":"1.0.0","description":"Git conventions","path":"rules/git-conventions"},
			{"name":"commit","type":"skill","version":"0.2.0","description":"Commit messages","path":"skills/commit"}
		]
	}`)

	idx, err := ParseIndex(data)
	if err != nil {
		t.Fatalf("ParseIndex error: %v", err)
	}
	if idx.Version != 1 {
		t.Errorf("Version: got %d, want 1", idx.Version)
	}
	if len(idx.Components) != 2 {
		t.Fatalf("Components: got %d, want 2", len(idx.Components))
	}
	if idx.Components[0].Name != "git-conventions" {
		t.Errorf("Components[0].Name: got %q, want %q", idx.Components[0].Name, "git-conventions")
	}
}

func TestParseIndex_InvalidVersion(t *testing.T) {
	data := []byte(`{"version": 99, "components": []}`)
	_, err := ParseIndex(data)
	if err == nil {
		t.Error("expected error for unsupported version")
	}
}

func TestParseIndex_InvalidJSON(t *testing.T) {
	_, err := ParseIndex([]byte(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
