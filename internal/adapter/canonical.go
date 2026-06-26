package adapter

// CanonicalComponent represents a component in the standard intermediate format.
// It is independent of any specific Agent format.
type CanonicalComponent struct {
	Meta  CanonicalMeta  `json:"meta"`
	Rule  *CanonicalRule `json:"rule,omitempty"`
	Skill *CanonicalSkill `json:"skill,omitempty"`
	Hook  *CanonicalHook `json:"hook,omitempty"`
	Files []CanonicalFile `json:"files"`
}

type CanonicalMeta struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Type        string `json:"type"`        // "rule" | "skill" | "hook"
	Description string `json:"description"`
	SourceAgent string `json:"sourceAgent"` // Source Agent identifier
	Hash        string `json:"hash"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

// CanonicalRule holds the semantic information of a Rule component.
// Note: Under the current adaptation strategy (Rule frontmatter preserved as-is),
// AlwaysApply and Paths are NOT populated during Rule adaptation.
// These fields are reserved for future iterations.
type CanonicalRule struct {
	// alwaysApply: true = always loaded, false = conditionally triggered
	AlwaysApply bool     `json:"alwaysApply"`
	// File glob patterns that trigger the rule; nil = always apply
	Paths      []string `json:"paths,omitempty"`
}

// CanonicalSkill holds the semantic information of a Skill component.
type CanonicalSkill struct {
	Description            string   `json:"description"`
	WhenToUse              string   `json:"whenToUse,omitempty"`
	AllowedTools           []string `json:"allowedTools,omitempty"`
	DisallowedTools        []string `json:"disallowedTools,omitempty"`
	DisableModelInvocation bool     `json:"disableModelInvocation"`
	UserInvocable          bool     `json:"userInvocable"`
	Context                string   `json:"context,omitempty"` // "fork" | ""
	Agent                  string   `json:"agent,omitempty"`
	Model                  string   `json:"model,omitempty"`  // "default" | "lite" | "reasoning"
}

// CanonicalHook holds the semantic information of a Hook component.
// Reserved for future iteration — not used in MVP.
type CanonicalHook struct {
	Events []CanonicalHookEvent `json:"events"`
}

type CanonicalHookEvent struct {
	EventName string             `json:"eventName"`
	Matcher   string             `json:"matcher,omitempty"`
	Handlers  []CanonicalHandler `json:"handlers"`
}

type CanonicalHandler struct {
	Type         string            `json:"type"`                  // "command" | "prompt" | "http" | "agent"
	Command      string            `json:"command,omitempty"`     // type=command
	Prompt       string            `json:"prompt,omitempty"`      // type=prompt / agent
	URL          string            `json:"url,omitempty"`         // type=http
	Method       string            `json:"method,omitempty"`      // type=http
	Headers      map[string]string `json:"headers,omitempty"`     // type=http
	Model        string            `json:"model,omitempty"`       // type=prompt / agent
	Timeout      int               `json:"timeout,omitempty"`     // seconds
	If           string            `json:"if,omitempty"`          // permission rule filter
	Once         bool              `json:"once,omitempty"`
	Async        bool              `json:"async,omitempty"`
	AsyncRewake  bool              `json:"asyncRewake,omitempty"`
}

// CanonicalFile represents a single file within a component.
type CanonicalFile struct {
	// Relative path from the component root, using forward slashes
	RelPath string `json:"relPath"`
	// File content
	Content []byte `json:"content"`
}
