package adapter

import "gopkg.in/yaml.v3"

// ClaudeAdapter implements Adapter for Claude Code format.
type ClaudeAdapter struct{}

func (a *ClaudeAdapter) AgentID() string { return "claude" }

// ToCanonical converts Claude-format files to Canonical.
// For Rule components: files are passed through as-is (no frontmatter parsing).
// For Skill components: SKILL.md frontmatter is parsed.
func (a *ClaudeAdapter) ToCanonical(files []CanonicalFile, meta CanonicalMeta) (*CanonicalComponent, error) {
	comp := &CanonicalComponent{Meta: meta}

	switch meta.Type {
	case "rule":
		// Rule: preserve files as-is, no frontmatter parsing
		comp.Rule = &CanonicalRule{}
		comp.Files = files
	case "skill":
		skill, processedFiles, err := parseSkillFrontmatter(files, "claude")
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

// FromCanonical converts Canonical to Claude-format files.
// For Rule components: content is output as-is (only body replacements applied).
// For Skill components: SKILL.md frontmatter is written with Claude-supported fields.
func (a *ClaudeAdapter) FromCanonical(canonical *CanonicalComponent) ([]CanonicalFile, error) {
	switch canonical.Meta.Type {
	case "rule":
		return ruleFromCanonical("claude", canonical)
	case "skill":
		return skillToClaude(canonical)
	default:
		return canonical.Files, nil
	}
}

// skillToClaude writes SKILL.md frontmatter with all Claude-supported fields.
func skillToClaude(canonical *CanonicalComponent) ([]CanonicalFile, error) {
	if canonical.Skill == nil {
		return canonical.Files, nil
	}

	result := make([]CanonicalFile, 0, len(canonical.Files))
	for _, f := range canonical.Files {
		if f.RelPath == "SKILL.md" {
			// Apply path/variable replacements to SKILL.md body
			replacedBody := ReplaceContent(f.Content, canonical.Meta.SourceAgent, "claude")
			newContent, err := writeSkillFrontmatter(replacedBody, canonical.Skill, "claude")
			if err != nil {
				return nil, err
			}
			result = append(result, CanonicalFile{RelPath: f.RelPath, Content: newContent})
		} else {
			// Apply path/variable replacements to non-SKILL.md files
			replaced := ReplaceContent(f.Content, canonical.Meta.SourceAgent, "claude")
			result = append(result, CanonicalFile{RelPath: f.RelPath, Content: replaced})
		}
	}
	return result, nil
}

// parseSkillFrontmatter extracts skill metadata from SKILL.md frontmatter.
func parseSkillFrontmatter(files []CanonicalFile, sourceAgent string) (*CanonicalSkill, []CanonicalFile, error) {
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
			if v, ok := GetStringField(fmNode, "when_to_use"); ok {
				skill.WhenToUse = v
			}
			if v, ok := GetStringSliceField(fmNode, "allowed-tools"); ok {
				skill.AllowedTools = v
			}
			if v, ok := GetStringSliceField(fmNode, "disallowed-tools"); ok {
				skill.DisallowedTools = v
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

		// Reconstruct SKILL.md with body only (frontmatter will be re-written on output)
		if hasFM {
			f.Content = body
		}
		processedFiles = append(processedFiles, f)
	}

	return skill, processedFiles, nil
}

// writeSkillFrontmatter writes skill metadata as YAML frontmatter on SKILL.md.
func writeSkillFrontmatter(body []byte, skill *CanonicalSkill, targetAgent string) ([]byte, error) {
	// Build YAML mapping node
	root := &yaml.Node{Kind: yaml.MappingNode}

	SetStringField(root, "description", skill.Description)

	if targetAgent == "claude" && skill.WhenToUse != "" {
		SetStringField(root, "when_to_use", skill.WhenToUse)
	}

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

	if targetAgent == "claude" && len(skill.DisallowedTools) > 0 {
		seq := &yaml.Node{Kind: yaml.SequenceNode}
		for _, t := range skill.DisallowedTools {
			seq.Content = append(seq.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: t})
		}
		root.Content = append(root.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "disallowed-tools"},
			seq,
		)
	}

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

	return SerializeFrontmatter(root, body)
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// ruleFromCanonical handles Rule component output — content as-is with body replacements.
func ruleFromCanonical(targetAgent string, canonical *CanonicalComponent) ([]CanonicalFile, error) {
	result := make([]CanonicalFile, 0, len(canonical.Files))
	for _, f := range canonical.Files {
		replaced := ReplaceContent(f.Content, canonical.Meta.SourceAgent, targetAgent)
		result = append(result, CanonicalFile{RelPath: f.RelPath, Content: replaced})
	}
	return result, nil
}
