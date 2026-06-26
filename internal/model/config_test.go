package model

import (
	"encoding/json"
	"testing"
)

func TestConfigJSONRoundTrip(t *testing.T) {
	c := Config{
		Source: "https://github.com/user/ai-components.git",
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Verify the JSON uses flat "source" key
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}
	if _, ok := raw["source"]; !ok {
		t.Error("JSON should contain top-level 'source' key")
	}
	if _, ok := raw["sourceUrl"]; ok {
		t.Error("JSON should NOT contain 'sourceUrl' key")
	}

	var got Config
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Source != c.Source {
		t.Errorf("Source: got %q, want %q", got.Source, c.Source)
	}
}
