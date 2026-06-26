package model

// AgentProfile describes an Agent's directory conventions and format constraints.
type AgentProfile struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	DirName           string `json:"dirName"`
	RuleFileName      string `json:"ruleFileName"`
	LocalRuleFileName string `json:"localRuleFileName"`
	RulesDir          string `json:"rulesDir"`
	SkillsDir         string `json:"skillsDir"`
	SettingsFile      string `json:"settingsFile"`

	// ComponentDirs maps ComponentType to the subdirectory relative to AgentRoot.
	ComponentDirs map[string]string `json:"componentDirs"`

	// FMConstraints defines frontmatter constraints for each component type.
	FMConstraints FrontmatterConstraints `json:"fmConstraints"`
}

// FrontmatterConstraints defines frontmatter rules for an Agent.
type FrontmatterConstraints struct {
	// SupportedFields maps ComponentType to a list of allowed frontmatter field names.
	SupportedFields map[string][]string `json:"supportedFields"`

	// DefaultValues maps "componentType.fieldName" to the default value to add when missing.
	// Note: Under the current Rule adaptation strategy (frontmatter preserved as-is),
	// Rule-related defaults are not used; only Skill-related defaults may apply.
	DefaultValues map[string]any `json:"defaultValues,omitempty"`

	// FieldConstraints maps "componentType.fieldName" to constraints on that field.
	FieldConstraints map[string]FieldConstraint `json:"fieldConstraints,omitempty"`
}

// FieldConstraint defines a constraint on a frontmatter field value.
type FieldConstraint struct {
	MaxLength int    `json:"maxLength,omitempty"` // 0 = no limit
	Behavior  string `json:"behavior,omitempty"`  // semantic behavior rule description
}
