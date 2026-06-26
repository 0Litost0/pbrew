package adapter

import (
	"fmt"
)

// Adapter implements bidirectional conversion between an Agent's file format and Canonical.
type Adapter interface {
	// FromCanonical converts Canonical to the target Agent's file format.
	FromCanonical(canonical *CanonicalComponent) ([]CanonicalFile, error)

	// ToCanonical converts the source Agent's file format to Canonical.
	ToCanonical(files []CanonicalFile, meta CanonicalMeta) (*CanonicalComponent, error)

	// AgentID returns the Agent identifier this adapter handles.
	AgentID() string
}

// NewAdapter creates an Adapter for the given agent ID.
func NewAdapter(agentID string) (Adapter, error) {
	switch agentID {
	case "claude":
		return &ClaudeAdapter{}, nil
	case "codebuddy":
		return &CodeBuddyAdapter{}, nil
	default:
		return nil, fmt.Errorf("unsupported agent: %s", agentID)
	}
}

// Transform executes the full conversion pipeline: source → Canonical → target.
// If sourceAgent == targetAgent, it returns the original files unchanged (identity transform).
func Transform(sourceAgent, targetAgent string, files []CanonicalFile, meta CanonicalMeta) ([]CanonicalFile, error) {
	// Identity transform: same source and target
	if sourceAgent == targetAgent {
		return files, nil
	}

	// Source → Canonical
	srcAdapter, err := NewAdapter(sourceAgent)
	if err != nil {
		return nil, fmt.Errorf("source adapter: %w", err)
	}
	canonical, err := srcAdapter.ToCanonical(files, meta)
	if err != nil {
		return nil, fmt.Errorf("to canonical: %w", err)
	}

	// Canonical → Target
	tgtAdapter, err := NewAdapter(targetAgent)
	if err != nil {
		return nil, fmt.Errorf("target adapter: %w", err)
	}
	result, err := tgtAdapter.FromCanonical(canonical)
	if err != nil {
		return nil, fmt.Errorf("from canonical: %w", err)
	}

	return result, nil
}
