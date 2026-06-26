package model

// ComponentType defines the type of an AI component.
type ComponentType string

const (
	TypeRule  ComponentType = "rule"
	TypeSkill ComponentType = "skill"
	TypeHook  ComponentType = "hook"
)

// SourceAgent identifies the Agent format of a component source.
type SourceAgent string

const (
	SourceClaude    SourceAgent = "claude"
	SourceCodeBuddy SourceAgent = "codebuddy"
)

// Component represents the metadata stored in .pbrew.json.
type Component struct {
	Name        string        `json:"name"`
	Version     string        `json:"version"`
	Type        ComponentType `json:"type"`
	Description string        `json:"description"`
	Hash        string        `json:"hash"`
	UpdatedAt   string        `json:"updatedAt,omitempty"`
	SourceAgent SourceAgent   `json:"sourceAgent,omitempty"`
}
