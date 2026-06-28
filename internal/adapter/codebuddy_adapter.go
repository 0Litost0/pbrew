package adapter

import "gopkg.in/yaml.v3"

// CodeBuddyAdapter implements Adapter for CodeBuddy Code format.
type CodeBuddyAdapter struct{}

func (a *CodeBuddyAdapter) AgentID() string { return "codebuddy" }

// ToCanonical converts CodeBuddy-format files to Canonical.
// For Rule components: files are passed through as-is (no frontmatter parsing).
// For Skill components: SKILL.md frontmatter is parsed (without claude-specific fields).
func (a *CodeBuddyAdapter) ToCanonical(files []CanonicalFile, meta CanonicalMeta) (*CanonicalComponent, error) {
	comp := &CanonicalComponent{Meta: meta}

	switch meta.Type {
	case "rule":
		// Rule: preserve files as-is
		comp.Rule = &CanonicalRule{}
		comp.Files = files
	case "skill":
		skill, processedFiles, err := parseCodeBuddySkillFrontmatter(files)
		if err != nil {
			return nil, err
		}
		comp.Skill = skill
		comp.Files = processedFiles
	default:
		comp.Files = files
	}

	return comp, nil
}

// FromCanonical converts Canonical to CodeBuddy-format files.
// For Rule components: content output as-is (only body replacements applied).
// For Skill components: SKILL.md frontmatter written with CodeBuddy-supported fields only.
func (a *CodeBuddyAdapter) FromCanonical(canonical *CanonicalComponent) ([]CanonicalFile, error) {
	switch canonical.Meta.Type {
	case "rule":
		return ruleFromCanonical("codebuddy", canonical)
	case "skill":
		return skillToCodeBuddy(canonical)
	default:
		return canonical.Files, nil
	}
}

// parseCodeBuddySkillFrontmatter extracts skill metadata from CodeBuddy SKILL.md.
// CodeBuddy does NOT support: when_to_use, disallowed-tools.
func parseCodeBuddySkillFrontmatter(files []CanonicalFile) (*CanonicalSkill, []CanonicalFile, error) {
	skill := &CanonicalSkill{}
	processedFiles := make([]CanonicalFile, 0, len(files))

	for _, f := range files {
		if f.RelPath != "SKILL.md" {
			processedFiles = append(processedFiles, f)
			continue
		}

		fmNode, body, hasFM, err := ParseFrontmatter(f.Content)
		if err != nil {
			return nil, nil, err
		}

		if hasFM && fmNode != nil {
			if v, ok := GetStringField(fmNode, "description"); ok {
				skill.Description = v
			}
			// CodeBuddy does NOT have when_to_use or disallowed-tools
			if v, ok := GetStringSliceField(fmNode, "allowed-tools"); ok {
				skill.AllowedTools = v
			}
			if v, ok := GetStringField(fmNode, "disable-model-invocation"); ok {
				skill.DisableModelInvocation = v == "true"
			}
			if v, ok := GetStringField(fmNode, "user-invocable"); ok {
				skill.UserInvocable = v == "true"
			}
			if v, ok := GetStringField(fmNode, "context"); ok {
				skill.Context = v
			}
			if v, ok := GetStringField(fmNode, "agent"); ok {
				skill.Agent = v
			}
			if v, ok := GetStringField(fmNode, "model"); ok {
				skill.Model = v
			}
		}

		if hasFM {
			f.Content = body
		}
		processedFiles = append(processedFiles, f)
	}

	return skill, processedFiles, nil
}

// skillToCodeBuddy writes SKILL.md frontmatter with CodeBuddy-supported fields only.
// Skips: when_to_use, disallowed-tools
// Truncates: description to 1024 characters
func skillToCodeBuddy(canonical *CanonicalComponent) ([]CanonicalFile, error) {
	if canonical.Skill == nil {
		return canonical.Files, nil
	}

	skill := canonical.Skill
	desc := skill.Description
	if len(desc) > 1024 {
		desc = desc[:1024]
	}

	result := make([]CanonicalFile, 0, len(canonical.Files))
	for _, f := range canonical.Files {
		if f.RelPath == "SKILL.md" {
			// Build YAML frontmatter node for CodeBuddy
			root := &yaml.Node{Kind: yaml.MappingNode}

			SetStringField(root, "description", desc)
			// Skip when_to_use (CodeBuddy doesn't support it)

			if len(skill.AllowedTools) > 0 {
				seq := &yaml.Node{Kind: yaml.SequenceNode}
				for _, t := range skill.AllowedTools {
					seq.Content = append(seq.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: t})
				}
				root.Content = append(root.Content,
					&yaml.Node{Kind: yaml.ScalarNode, Value: "allowed-tools"},
					seq,
				)
			}
			// Skip disallowed-tools (CodeBuddy doesn't support it)

			SetStringField(root, "disable-model-invocation", boolStr(skill.DisableModelInvocation))
			SetStringField(root, "user-invocable", boolStr(skill.UserInvocable))

			if skill.Context != "" {
				SetStringField(root, "context", skill.Context)
			}
			if skill.Agent != "" {
				SetStringField(root, "agent", skill.Agent)
			}
			if skill.Model != "" {
				SetStringField(root, "model", skill.Model)
			}

			// Apply path/variable replacements to SKILL.md body
			replacedBody := ReplaceContent(f.Content, canonical.Meta.SourceAgent, "codebuddy")
			newContent, err := SerializeFrontmatter(root, replacedBody)
			if err != nil {
				return nil, err
			}
			result = append(result, CanonicalFile{RelPath: f.RelPath, Content: newContent})
		} else {
			replaced := ReplaceContent(f.Content, canonical.Meta.SourceAgent, "codebuddy")
			result = append(result, CanonicalFile{RelPath: f.RelPath, Content: replaced})
		}
	}
	return result, nil
}
