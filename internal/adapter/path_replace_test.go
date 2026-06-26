package adapter

import (
	"testing"
)

func TestReplaceContentClaudeToCodeBuddy(t *testing.T) {
	input := []byte("# Rule\n\nSee @.claude/rules/other.md for details.\nUse ${CLAUDE_SKILL_DIR}/run.sh\n")
	expected := []byte("# Rule\n\nSee @.codebuddy/rules/other.md for details.\nUse ${CODEBUDDY_SKILL_DIR}/run.sh\n")

	result := ReplaceContent(input, "claude", "codebuddy")
	if string(result) != string(expected) {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestReplaceContentCodeBuddyToClaude(t *testing.T) {
	input := []byte("# Rule\n\nSee @.codebuddy/rules/other.md\nUse ${CODEBUDDY_PROJECT_DIR}/tool.sh\n")
	expected := []byte("# Rule\n\nSee @.claude/rules/other.md\nUse ${CLAUDE_PROJECT_DIR}/tool.sh\n")

	result := ReplaceContent(input, "codebuddy", "claude")
	if string(result) != string(expected) {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestReplaceContentSameAgent(t *testing.T) {
	input := []byte("@.claude/rules/other.md ${CLAUDE_SKILL_DIR}")
	result := ReplaceContent(input, "claude", "claude")
	if string(result) != string(input) {
		t.Error("same agent should not modify content")
	}
}

func TestReplaceContentPreservesFrontmatter(t *testing.T) {
	input := []byte("---\nalwaysApply: true\npaths: .claude/rules/test.md\n---\n\nSee @.claude/rules/other.md\n")
	result := ReplaceContent(input, "claude", "codebuddy")

	content := string(result)
	// Frontmatter should be preserved (including .claude/ inside it)
	if !contains(content, "paths: .claude/rules/test.md") {
		t.Error("frontmatter paths should not be replaced")
	}
	// Body should be replaced
	if !contains(content, "@.codebuddy/rules/other.md") {
		t.Error("body @ reference should be replaced")
	}
}

func TestReplaceFilePaths(t *testing.T) {
	tests := []struct {
		path   string
		source string
		target string
		want   string
	}{
		{".claude/rules/test/rule.md", "claude", "codebuddy", ".codebuddy/rules/test/rule.md"},
		{".codebuddy/rules/test/rule.md", "codebuddy", "claude", ".claude/rules/test/rule.md"},
		{"rules/test.md", "claude", "codebuddy", "rules/test.md"},
	}
	for _, tt := range tests {
		got := ReplaceFilePaths(tt.path, tt.source, tt.target)
		if got != tt.want {
			t.Errorf("ReplaceFilePaths(%q, %q, %q) = %q, want %q", tt.path, tt.source, tt.target, got, tt.want)
		}
	}
}

func TestAllVariableReplacements(t *testing.T) {
	input := []byte("${CLAUDE_SKILL_DIR} ${CLAUDE_PROJECT_DIR} ${CLAUDE_PLUGIN_ROOT} ${CLAUDE_PLUGIN_DATA}")
	result := ReplaceContent(input, "claude", "codebuddy")
	content := string(result)

	if !contains(content, "${CODEBUDDY_SKILL_DIR}") {
		t.Error("CLAUDE_SKILL_DIR not replaced")
	}
	if !contains(content, "${CODEBUDDY_PROJECT_DIR}") {
		t.Error("CLAUDE_PROJECT_DIR not replaced")
	}
	if !contains(content, "${CODEBUDDY_PLUGIN_ROOT}") {
		t.Error("CLAUDE_PLUGIN_ROOT not replaced")
	}
	if !contains(content, "${CODEBUDDY_PLUGIN_DATA}") {
		t.Error("CLAUDE_PLUGIN_DATA not replaced")
	}
}
