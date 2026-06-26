package model

import (
	"encoding/json"
	"testing"
)

func TestComponentJSONRoundTrip(t *testing.T) {
	c := Component{
		Name:        "git-conventions",
		Version:     "1.0.0",
		Type:        TypeRule,
		Description: "Git commit and branch conventions",
		Hash:        "abc123",
		UpdatedAt:   "2026-01-01T00:00:00Z",
		SourceAgent: SourceClaude,
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Component
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Name != c.Name {
		t.Errorf("Name: got %q, want %q", got.Name, c.Name)
	}
	if got.Version != c.Version {
		t.Errorf("Version: got %q, want %q", got.Version, c.Version)
	}
	if got.Type != c.Type {
		t.Errorf("Type: got %q, want %q", got.Type, c.Type)
	}
	if got.SourceAgent != c.SourceAgent {
		t.Errorf("SourceAgent: got %q, want %q", got.SourceAgent, c.SourceAgent)
	}
}

func TestComponentTypeConstants(t *testing.T) {
	tests := []struct {
		ct   ComponentType
		want string
	}{
		{TypeRule, "rule"},
		{TypeSkill, "skill"},
		{TypeHook, "hook"},
	}
	for _, tt := range tests {
		if string(tt.ct) != tt.want {
			t.Errorf("got %q, want %q", tt.ct, tt.want)
		}
	}
}

func TestSourceAgentConstants(t *testing.T) {
	if string(SourceClaude) != "claude" {
		t.Errorf("SourceClaude: got %q, want %q", SourceClaude, "claude")
	}
	if string(SourceCodeBuddy) != "codebuddy" {
		t.Errorf("SourceCodeBuddy: got %q, want %q", SourceCodeBuddy, "codebuddy")
	}
}
