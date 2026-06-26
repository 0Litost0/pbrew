package adapter

import (
	"testing"
)

func TestParseFrontmatter_WithFrontmatter(t *testing.T) {
	input := []byte("---\npaths: src/api/**/*.ts\n---\n# Rules\n")
	fm, body, hasFM, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatal(err)
	}
	if !hasFM {
		t.Error("should detect frontmatter")
	}
	if fm == nil {
		t.Fatal("frontmatter node should not be nil")
	}
	if fm.Kind != 4 { // yaml.MappingNode
		t.Errorf("expected MappingNode (4), got %d", fm.Kind)
	}
	if len(body) == 0 {
		t.Error("body should not be empty")
	}
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	input := []byte("# Just a rule\nNo frontmatter here.\n")
	fm, body, hasFM, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatal(err)
	}
	if hasFM {
		t.Error("should not detect frontmatter")
	}
	if fm != nil {
		t.Error("frontmatter node should be nil")
	}
	if string(body) != string(input) {
		t.Error("body should equal input when no frontmatter")
	}
}

func TestParseFrontmatter_GetStringField(t *testing.T) {
	input := []byte("---\npaths: src/api/**/*.ts\ndescription: API rules\n---\n# Body\n")
	fm, _, _, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := GetStringField(fm, "paths"); !ok || v != "src/api/**/*.ts" {
		t.Errorf("paths: got %q, ok=%v", v, ok)
	}
	if v, ok := GetStringField(fm, "description"); !ok || v != "API rules" {
		t.Errorf("description: got %q, ok=%v", v, ok)
	}
}

func TestParseFrontmatter_GetStringSliceField(t *testing.T) {
	input := []byte("---\nallowed-tools:\n  - Bash\n  - Read\n---\n# Body\n")
	fm, _, _, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatal(err)
	}
	v, ok := GetStringSliceField(fm, "allowed-tools")
	if !ok {
		t.Error("should find allowed-tools")
	}
	if len(v) != 2 || v[0] != "Bash" || v[1] != "Read" {
		t.Errorf("got %v", v)
	}
}

func TestSerializeFrontmatter_RoundTrip(t *testing.T) {
	input := []byte("---\npaths: src/**/*.ts\ndescription: Test\n---\n# Body\n")
	fm, body, hasFM, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatal(err)
	}
	if !hasFM {
		t.Fatal("expected frontmatter")
	}

	output, err := SerializeFrontmatter(fm, body)
	if err != nil {
		t.Fatal(err)
	}

	// Re-parse and verify
	fm2, body2, hasFM2, err := ParseFrontmatter(output)
	if err != nil {
		t.Fatal(err)
	}
	if !hasFM2 {
		t.Error("round-trip should preserve frontmatter")
	}
	if v, ok := GetStringField(fm2, "paths"); !ok || v != "src/**/*.ts" {
		t.Errorf("paths after round-trip: got %q", v)
	}
	if string(body2) != string(body) {
		t.Errorf("body mismatch after round-trip:\ngot:  %q\nwant: %q", body2, body)
	}
}
