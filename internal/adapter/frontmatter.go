package adapter

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseFrontmatter splits a markdown file into frontmatter YAML and body.
// Returns (frontmatterNode, body, hasFrontmatter).
// If no frontmatter is found, frontmatterNode is nil and body is the full content.
func ParseFrontmatter(content []byte) (frontmatter *yaml.Node, body []byte, hasFrontmatter bool, err error) {
	// Find opening ---
	idx := bytes.IndexByte(content, '\n')
	if idx == -1 {
		// Single line, no frontmatter
		return nil, content, false, nil
	}

	firstLine := strings.TrimSpace(string(content[:idx]))
	if firstLine != "---" {
		return nil, content, false, nil
	}

	// Find closing ---
	rest := content[idx+1:]
	closeIdx := bytes.Index(rest, []byte("\n---"))
	if closeIdx == -1 {
		return nil, content, false, nil
	}

	fmBytes := rest[:closeIdx]
	bodyStart := closeIdx + 4 // skip \n---
	if bodyStart < len(rest) && rest[bodyStart] == '\n' {
		bodyStart++
	}
	body = rest[bodyStart:]

	var node yaml.Node
	if err := yaml.Unmarshal(fmBytes, &node); err != nil {
		return nil, nil, false, fmt.Errorf("parse frontmatter YAML: %w", err)
	}

	// yaml.Unmarshal returns a DocumentNode wrapping the actual content
	var contentNode *yaml.Node
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		contentNode = node.Content[0]
	} else {
		contentNode = &node
	}

	return contentNode, body, true, nil
}

// SerializeFrontmatter reconstructs a markdown file from a frontmatter node and body.
// If frontmatter is nil, returns body as-is.
func SerializeFrontmatter(frontmatter *yaml.Node, body []byte) ([]byte, error) {
	if frontmatter == nil {
		return body, nil
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")

	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(frontmatter); err != nil {
		return nil, fmt.Errorf("serialize frontmatter: %w", err)
	}
	// yaml.Encoder adds a trailing newline via the document separator "---\n",
	// we need to close the encoder to flush, then trim the trailing "---"
	// that the yaml encoder adds as a document separator
	encoder.Close()

	output := buf.Bytes()
	// Remove the trailing document separator "---\n" that yaml encoder adds
	if bytes.HasSuffix(output, []byte("---\n")) {
		output = output[:len(output)-4]
	}

	buf = *bytes.NewBuffer(output)
	// Ensure single newline before closing ---
	if !bytes.HasSuffix(buf.Bytes(), []byte("\n")) {
		buf.WriteByte('\n')
	}
	buf.WriteString("---\n")

	if len(body) > 0 {
		buf.Write(body)
	}

	return buf.Bytes(), nil
}

// GetStringField retrieves a string value from a YAML mapping node.
func GetStringField(node *yaml.Node, key string) (string, bool) {
	if node.Kind != yaml.MappingNode {
		return "", false
	}
	for i := 0; i < len(node.Content)-1; i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1].Value, true
		}
	}
	return "", false
}

// SetStringField sets a string value in a YAML mapping node.
// If the key already exists, updates the value. Otherwise, appends it.
func SetStringField(node *yaml.Node, key, value string) {
	if node.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i < len(node.Content)-1; i += 2 {
		if node.Content[i].Value == key {
			node.Content[i+1].Value = value
			return
		}
	}
	// Append new key-value pair
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Value: value},
	)
}

// DeleteField removes a key from a YAML mapping node.
func DeleteField(node *yaml.Node, key string) {
	if node.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i < len(node.Content)-1; i += 2 {
		if node.Content[i].Value == key {
			node.Content = append(node.Content[:i], node.Content[i+2:]...)
			return
		}
	}
}

// GetStringSliceField retrieves a string slice value from a YAML mapping node.
func GetStringSliceField(node *yaml.Node, key string) ([]string, bool) {
	if node.Kind != yaml.MappingNode {
		return nil, false
	}
	for i := 0; i < len(node.Content)-1; i += 2 {
		if node.Content[i].Value == key && node.Content[i+1].Kind == yaml.SequenceNode {
			var items []string
			seq := node.Content[i+1]
			for _, item := range seq.Content {
				items = append(items, item.Value)
			}
			return items, true
		}
	}
	return nil, false
}
