package publisher

import (
	"encoding/json"
	"strings"

	"github.com/user/pbrew/internal/model"
)

const metaFile = ".pbrew.json"

// detectComponentType determines whether a set of files represents a Rule or Skill.
// If SKILL.md exists, it's a skill; otherwise it's a rule.
func detectComponentType(files map[string][]byte) model.ComponentType {
	if _, ok := files["SKILL.md"]; ok {
		return model.TypeSkill
	}
	return model.TypeRule
}

// detectSourceAgent determines the source Agent format from the component files.
// It first checks for an existing .pbrew.json with a sourceAgent field.
// If not found, it infers from file content (looks for CodeBuddy-specific patterns).
// Defaults to "claude".
func detectSourceAgent(files map[string][]byte) string {
	// 1. Check existing .pbrew.json
	if data, ok := files[metaFile]; ok {
		var comp model.Component
		if err := json.Unmarshal(data, &comp); err == nil && comp.SourceAgent != "" {
			return string(comp.SourceAgent)
		}
	}

	// 2. Infer from content: scan .md files for CodeBuddy-specific patterns
	for name, content := range files {
		if !strings.HasSuffix(name, ".md") || name == metaFile {
			continue
		}
		s := string(content)
		if strings.Contains(s, ".codebuddy/") ||
			strings.Contains(s, "CODEBUDDY_SKILL_DIR") ||
			strings.Contains(s, "CODEBUDDY_PROJECT_DIR") {
			return "codebuddy"
		}
	}

	return "claude"
}

// detectDescription extracts a short description for the component metadata.
// For Rule: uses frontmatter "description" field if present.
// For Skill: uses SKILL.md frontmatter "description" field.
// Falls back to empty string.
func detectDescription(files map[string][]byte, compType model.ComponentType) string {
	var content []byte
	switch compType {
	case model.TypeSkill:
		content = files["SKILL.md"]
	default:
		// For rule, pick the first .md file that is not .pbrew.json
		for name, data := range files {
			if strings.HasSuffix(name, ".md") && name != metaFile {
				content = data
				break
			}
		}
	}

	if content == nil {
		return ""
	}

	// Look for description in frontmatter
	s := string(content)
	if !strings.HasPrefix(s, "---") {
		return ""
	}
	end := strings.Index(s[3:], "\n---")
	if end == -1 {
		return ""
	}
	fm := s[3 : end+3]
	for _, line := range strings.Split(fm, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "description:") {
			desc := strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			// Remove surrounding quotes if present
			desc = strings.Trim(desc, "\"'")
			return desc
		}
	}
	return ""
}
