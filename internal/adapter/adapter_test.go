package adapter

import (
	"testing"
)

func TestNewAdapter(t *testing.T) {
	tests := []struct {
		id      string
		wantErr bool
	}{
		{"claude", false},
		{"codebuddy", false},
		{"unknown", true},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			a, err := NewAdapter(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAdapter(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
			if err == nil && a.AgentID() != tt.id {
				t.Errorf("AgentID() = %q, want %q", a.AgentID(), tt.id)
			}
		})
	}
}

func TestTransformIdentity(t *testing.T) {
	files := []CanonicalFile{
		{RelPath: "rule.md", Content: []byte("# Test\nSee @.claude/rules/other.md\n")},
	}
	meta := CanonicalMeta{Name: "test", Type: "rule", SourceAgent: "claude"}

	result, err := Transform("claude", "claude", files, meta)
	if err != nil {
		t.Fatal(err)
	}
	if string(result[0].Content) != string(files[0].Content) {
		t.Errorf("identity transform should not modify content")
	}
}

func TestTransformRuleClaudeToCodeBuddy(t *testing.T) {
	files := []CanonicalFile{
		{RelPath: "my-rule.md", Content: []byte("---\nalwaysApply: true\n---\n\nSee @.claude/rules/other.md\nRun ${CLAUDE_SKILL_DIR}/check.sh\n")},
	}
	meta := CanonicalMeta{Name: "test", Type: "rule", SourceAgent: "claude"}

	result, err := Transform("claude", "codebuddy", files, meta)
	if err != nil {
		t.Fatal(err)
	}

	content := string(result[0].Content)

	// frontmatter should be preserved as-is
	if !contains(content, "alwaysApply: true") {
		t.Error("frontmatter should be preserved")
	}

	// body path should be replaced
	if !contains(content, "@.codebuddy/rules/other.md") {
		t.Errorf("expected .codebuddy/ path in body, got: %s", content)
	}

	// body variable should be replaced
	if !contains(content, "${CODEBUDDY_SKILL_DIR}/check.sh") {
		t.Errorf("expected CODEBUDDY variable in body, got: %s", content)
	}
}

func TestTransformRuleCodeBuddyToClaude(t *testing.T) {
	files := []CanonicalFile{
		{RelPath: "my-rule.md", Content: []byte("---\ndescription: Test\nalwaysApply: false\nenabled: true\n---\n\nSee @.codebuddy/rules/other.md\nRun ${CODEBUDDY_SKILL_DIR}/check.sh\n")},
	}
	meta := CanonicalMeta{Name: "test", Type: "rule", SourceAgent: "codebuddy"}

	result, err := Transform("codebuddy", "claude", files, meta)
	if err != nil {
		t.Fatal(err)
	}

	content := string(result[0].Content)

	// frontmatter should be preserved as-is (including codebuddy-specific fields)
	if !contains(content, "alwaysApply: false") {
		t.Error("frontmatter alwaysApply should be preserved")
	}
	if !contains(content, "enabled: true") {
		t.Error("frontmatter enabled should be preserved")
	}

	// body path should be replaced
	if !contains(content, "@.claude/rules/other.md") {
		t.Errorf("expected .claude/ path in body, got: %s", content)
	}

	// body variable should be replaced
	if !contains(content, "${CLAUDE_SKILL_DIR}/check.sh") {
		t.Errorf("expected CLAUDE variable in body, got: %s", content)
	}
}

func TestTransformSkillClaudeToCodeBuddy(t *testing.T) {
	skillContent := []byte("---\ndescription: Generate commit messages\nwhen_to_use: When you need a commit message\ndisallowed-tools: [Edit]\nallowed-tools: [Bash, Read]\n---\n\n# Commit Skill\n\nGenerate a message.")
	files := []CanonicalFile{
		{RelPath: "SKILL.md", Content: skillContent},
	}
	meta := CanonicalMeta{Name: "commit", Type: "skill", SourceAgent: "claude"}

	result, err := Transform("claude", "codebuddy", files, meta)
	if err != nil {
		t.Fatal(err)
	}

	content := string(result[0].Content)

	// description should be present
	if !contains(content, "description:") {
		t.Error("description field should be present")
	}

	// when_to_use should NOT be present (CodeBuddy doesn't support)
	if contains(content, "when_to_use:") {
		t.Error("when_to_use should be removed for CodeBuddy")
	}

	// disallowed-tools should NOT be present (CodeBuddy doesn't support)
	if contains(content, "disallowed-tools:") {
		t.Error("disallowed-tools should be removed for CodeBuddy")
	}

	// allowed-tools should be present
	if !contains(content, "allowed-tools:") {
		t.Error("allowed-tools should be present")
	}
}

func TestTransformSkillDescriptionTruncation(t *testing.T) {
	longDesc := makeLongString(2000)
	skillContent := []byte("---\ndescription: " + longDesc + "\n---\n\n# Skill")
	files := []CanonicalFile{
		{RelPath: "SKILL.md", Content: skillContent},
	}
	meta := CanonicalMeta{Name: "test", Type: "skill", SourceAgent: "claude"}

	result, err := Transform("claude", "codebuddy", files, meta)
	if err != nil {
		t.Fatal(err)
	}

	content := string(result[0].Content)
	// The description in the output should be <= 1024 chars when serialized
	if contains(content, longDesc) {
		t.Error("description should be truncated to 1024 chars for CodeBuddy")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func makeLongString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a' + byte(i%26)
	}
	return string(b)
}
