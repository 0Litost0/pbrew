package adapter

import (
	"bytes"
	"strings"
)

// ReplacementRule defines a single string replacement pair.
type ReplacementRule struct {
	From string
	To   string
}

// replacementRules defines the path and variable replacements for each direction.
var replacementRules = map[string]map[string][]ReplacementRule{
	// Claude → CodeBuddy
	"claude->codebuddy": {
		"paths": {
			{".claude/", ".codebuddy/"},
			{"CLAUDE.md", "CODEBUDDY.md"},
			{"CLAUDE.local.md", "CODEBUDDY.local.md"},
		},
		"variables": {
			{"${CLAUDE_SKILL_DIR}", "${CODEBUDDY_SKILL_DIR}"},
			{"${CLAUDE_PROJECT_DIR}", "${CODEBUDDY_PROJECT_DIR}"},
			{"${CLAUDE_PLUGIN_ROOT}", "${CODEBUDDY_PLUGIN_ROOT}"},
			{"${CLAUDE_PLUGIN_DATA}", "${CODEBUDDY_PLUGIN_DATA}"},
		},
	},
	// CodeBuddy → Claude
	"codebuddy->claude": {
		"paths": {
			{".codebuddy/", ".claude/"},
			{"CODEBUDDY.md", "CLAUDE.md"},
			{"CODEBUDDY.local.md", "CLAUDE.local.md"},
		},
		"variables": {
			{"${CODEBUDDY_SKILL_DIR}", "${CLAUDE_SKILL_DIR}"},
			{"${CODEBUDDY_PROJECT_DIR}", "${CLAUDE_PROJECT_DIR}"},
			{"${CODEBUDDY_PLUGIN_ROOT}", "${CLAUDE_PLUGIN_ROOT}"},
			{"${CODEBUDDY_PLUGIN_DATA}", "${CLAUDE_PLUGIN_DATA}"},
		},
	},
}

// ReplaceContent applies path and variable replacements to file content.
// It operates on the body text only (outside frontmatter), preserving frontmatter as-is.
func ReplaceContent(content []byte, sourceAgent, targetAgent string) []byte {
	if sourceAgent == targetAgent {
		return content
	}

	key := sourceAgent + "->" + targetAgent
	rules, ok := replacementRules[key]
	if !ok {
		return content
	}

	// Split frontmatter and body
	_, body, hasFM, err := ParseFrontmatter(content)
	if err != nil {
		// If parsing fails, apply replacements to entire content
		return applyReplacements(content, rules)
	}

	if !hasFM {
		return applyReplacements(content, rules)
	}

	// Apply replacements only to the body portion
	replacedBody := applyReplacements(body, rules)

	// Reconstruct: original frontmatter + replaced body
	// Find the body start position in original content
	fmEnd := bytes.Index(content[len("---\n"):], []byte("\n---"))
	if fmEnd == -1 {
		return content
	}
	// Original frontmatter portion (including --- delimiters)
	fmPortion := content[:len("---\n")+fmEnd+len("\n---\n")]

	return append(fmPortion, replacedBody...)
}

// applyReplacements applies all replacement rules to content.
func applyReplacements(content []byte, rules map[string][]ReplacementRule) []byte {
	result := string(content)

	// Apply path replacements first (longer strings first to avoid partial matches)
	allRules := make([]ReplacementRule, 0, len(rules["paths"])+len(rules["variables"]))
	allRules = append(allRules, rules["paths"]...)
	allRules = append(allRules, rules["variables"]...)

	// Sort by length descending to avoid partial matches
	sortRulesByLength(allRules)

	for _, r := range allRules {
		result = strings.ReplaceAll(result, r.From, r.To)
	}

	return []byte(result)
}

// sortRulesByLength sorts replacement rules by From string length descending.
func sortRulesByLength(rules []ReplacementRule) {
	for i := 0; i < len(rules)-1; i++ {
		for j := i + 1; j < len(rules); j++ {
			if len(rules[j].From) > len(rules[i].From) {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}
}

// ReplaceFilePaths applies path replacements to file paths (not content).
// Used when determining installation target paths.
func ReplaceFilePaths(path, sourceAgent, targetAgent string) string {
	if sourceAgent == targetAgent {
		return path
	}

	key := sourceAgent + "->" + targetAgent
	rules, ok := replacementRules[key]
	if !ok {
		return path
	}

	result := path
	for _, r := range rules["paths"] {
		result = strings.ReplaceAll(result, r.From, r.To)
	}
	return result
}
