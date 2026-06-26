package agent

import (
	"testing"
)

func TestGet(t *testing.T) {
	tests := []struct {
		id      string
		wantOk  bool
		wantDir string
	}{
		{"claude", true, ".claude"},
		{"codebuddy", true, ".codebuddy"},
		{"unknown", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			profile, ok := Get(tt.id)
			if ok != tt.wantOk {
				t.Fatalf("Get(%q): ok = %v, want %v", tt.id, ok, tt.wantOk)
			}
			if !ok {
				return
			}
			if profile.DirName != tt.wantDir {
				t.Errorf("DirName: got %q, want %q", profile.DirName, tt.wantDir)
			}
		})
	}
}

func TestList(t *testing.T) {
	profiles := List()
	if len(profiles) < 2 {
		t.Errorf("List() returned %d profiles, want at least 2", len(profiles))
	}

	seen := make(map[string]bool)
	for _, p := range profiles {
		if seen[p.ID] {
			t.Errorf("duplicate profile ID: %q", p.ID)
		}
		seen[p.ID] = true
	}
	if !seen["claude"] || !seen["codebuddy"] {
		t.Error("List() should contain both 'claude' and 'codebuddy'")
	}
}

func TestValidate(t *testing.T) {
	if err := Validate("claude"); err != nil {
		t.Errorf("Validate('claude'): unexpected error: %v", err)
	}
	if err := Validate("codebuddy"); err != nil {
		t.Errorf("Validate('codebuddy'): unexpected error: %v", err)
	}
	if err := Validate("unknown"); err == nil {
		t.Error("Validate('unknown'): expected error, got nil")
	}
}

func TestCodeBuddyNoRuleDefaults(t *testing.T) {
	// Verify that codebuddy AgentProfile has no rule-related DefaultValues
	// (per the new Rule frontmatter preservation strategy)
	profile, ok := Get("codebuddy")
	if !ok {
		t.Fatal("codebuddy profile not found")
	}

	for key := range profile.FMConstraints.DefaultValues {
		if len(key) >= 4 && key[:4] == "rule" {
			t.Errorf("codebuddy should not have rule-related DefaultValues, found key: %q", key)
		}
	}

	for key := range profile.FMConstraints.FieldConstraints {
		if len(key) >= 4 && key[:4] == "rule" {
			t.Errorf("codebuddy should not have rule-related FieldConstraints, found key: %q", key)
		}
	}
}

func TestCodeBuddySkillDescriptionMaxLength(t *testing.T) {
	profile, ok := Get("codebuddy")
	if !ok {
		t.Fatal("codebuddy profile not found")
	}

	fc, exists := profile.FMConstraints.FieldConstraints["skill.description"]
	if !exists {
		t.Fatal("codebuddy should have skill.description FieldConstraint")
	}
	if fc.MaxLength != 1024 {
		t.Errorf("skill.description MaxLength: got %d, want 1024", fc.MaxLength)
	}
}

func TestClaudeSkillDescriptionNoLimit(t *testing.T) {
	profile, ok := Get("claude")
	if !ok {
		t.Fatal("claude profile not found")
	}

	fc, exists := profile.FMConstraints.FieldConstraints["skill.description"]
	if !exists {
		// MaxLength: 0 means no limit, check that it's defined with 0
		t.Fatal("claude should have skill.description FieldConstraint")
	}
	if fc.MaxLength != 0 {
		t.Errorf("claude skill.description MaxLength: got %d, want 0 (no limit)", fc.MaxLength)
	}
}
