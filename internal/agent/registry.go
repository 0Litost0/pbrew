package agent

import (
	"fmt"

	"github.com/user/pbrew/internal/model"
)

// builtIn defines the built-in Agent profiles.
var builtIn = map[string]model.AgentProfile{
	"claude": {
		ID:                "claude",
		DisplayName:       "Claude Code",
		DirName:           ".claude",
		RuleFileName:      "CLAUDE.md",
		LocalRuleFileName: "CLAUDE.local.md",
		RulesDir:          "rules",
		SkillsDir:         "skills",
		SettingsFile:      "settings.json",
		ComponentDirs: map[string]string{
			"rule":  "rules",
			"skill": "skills",
			"hook":  "", // Claude hooks live in settings.json, no standalone directory
		},
		FMConstraints: model.FrontmatterConstraints{
			SupportedFields: map[string][]string{
				"rule":  {"paths"},
				"skill": {"description", "when_to_use", "allowed-tools", "disallowed-tools",
					"disable-model-invocation", "user-invocable", "context", "agent", "model"},
			},
			FieldConstraints: map[string]model.FieldConstraint{
				"skill.description": {MaxLength: 0}, // no limit
			},
		},
	},
	"codebuddy": {
		ID:                "codebuddy",
		DisplayName:       "CodeBuddy Code",
		DirName:           ".codebuddy",
		RuleFileName:      "CODEBUDDY.md",
		LocalRuleFileName: "CODEBUDDY.local.md",
		RulesDir:          "rules",
		SkillsDir:         "skills",
		SettingsFile:      "settings.json",
		ComponentDirs: map[string]string{
			"rule":  "rules",
			"skill": "skills",
			"hook":  "", // CodeBuddy hooks live in settings.json, no standalone directory
		},
		FMConstraints: model.FrontmatterConstraints{
			SupportedFields: map[string][]string{
				"rule":  {"description", "alwaysApply", "enabled", "paths", "updatedAt", "provider"},
				"skill": {"description", "allowed-tools",
					"disable-model-invocation", "user-invocable", "context", "agent", "model"},
			},
			FieldConstraints: map[string]model.FieldConstraint{
				"skill.description": {MaxLength: 1024},
			},
		},
	},
}

// Get returns the AgentProfile for the given agent ID.
func Get(id string) (model.AgentProfile, bool) {
	profile, ok := builtIn[id]
	return profile, ok
}

// List returns all built-in Agent profiles.
func List() []model.AgentProfile {
	profiles := make([]model.AgentProfile, 0, len(builtIn))
	for _, p := range builtIn {
		profiles = append(profiles, p)
	}
	return profiles
}

// Validate checks whether the given agent ID is a known built-in Agent.
func Validate(id string) error {
	if _, ok := builtIn[id]; !ok {
		return fmt.Errorf("unsupported agent: %s", id)
	}
	return nil
}
