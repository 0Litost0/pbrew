package registry

import (
	"encoding/json"
	"fmt"
)

// Index represents the _index.json root structure.
type Index struct {
	Version    int              `json:"version"`
	Components []ComponentEntry `json:"components"`
}

// ComponentEntry is a summary record in _index.json.
type ComponentEntry struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Path        string `json:"path"`
}

// ParseIndex parses _index.json content.
func ParseIndex(data []byte) (*Index, error) {
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parse _index.json: %w", err)
	}
	if idx.Version != 1 {
		return nil, fmt.Errorf("unsupported _index.json version: %d", idx.Version)
	}
	return &idx, nil
}
