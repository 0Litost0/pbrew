package publisher

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/pbrew/internal/model"
	"github.com/user/pbrew/internal/source"
)

// TestPublishCodeBuddyFormat verifies that a CodeBuddy-format Rule is correctly
// transformed to Claude format when published.
func TestPublishCodeBuddyFormat(t *testing.T) {
	repoDir := t.TempDir()
	src, err := source.NewLocalSource(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Create a CodeBuddy-format Rule component
	compDir := t.TempDir()
	// .pbrew.json indicates sourceAgent is codebuddy
	comp := model.Component{
		Name:        "cb-rule",
		Version:     "0.3.0",
		Type:        model.TypeRule,
		Description: "A codebuddy rule",
		Hash:        "abc",
		SourceAgent: model.SourceCodeBuddy,
	}
	metaData, _ := json.MarshalIndent(comp, "", "  ")
	os.WriteFile(filepath.Join(compDir, ".pbrew.json"), metaData, 0644)

	ruleContent := "---\nalwaysApply: true\n---\n\n# CB Rule\n\nSee @.codebuddy/rules/api.md\nRun ${CODEBUDDY_SKILL_DIR}/check.sh\n"
	os.WriteFile(filepath.Join(compDir, "cb-rule.md"), []byte(ruleContent), 0644)

	_, err = Publish(PublishParams{
		Name:   "cb-rule",
		Dir:    compDir,
		Source: src,
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	// Verify the rule was written with Claude format paths (transformed)
	written, _ := os.ReadFile(filepath.Join(repoDir, "rules", "cb-rule", "cb-rule.md"))
	s := string(written)

	// Body should have Claude paths (transformed from CodeBuddy)
	if !contains(s, "@.claude/rules/api.md") {
		t.Errorf("expected .claude/ path in body after transformation, got: %s", s)
	}
	if !contains(s, "${CLAUDE_SKILL_DIR}/check.sh") {
		t.Errorf("expected CLAUDE_SKILL_DIR variable in body after transformation, got: %s", s)
	}

	// Frontmatter should be preserved as-is (Rule policy)
	if !contains(s, "alwaysApply: true") {
		t.Error("frontmatter should be preserved as-is")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
