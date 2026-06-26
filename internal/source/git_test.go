package source

import "testing"

func TestIsGitURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"https://github.com/user/repo.git", true},
		{"git@github.com:user/repo.git", true},
		{"git://example.com/repo.git", true},
		{"/local/path", false},
		{"./relative/path", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsGitURL(tt.input)
			if got != tt.want {
				t.Errorf("IsGitURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGitSourceNotOpened(t *testing.T) {
	gs := NewGitSource("https://github.com/user/repo.git")
	// Without Open(), operations should fail
	if _, err := gs.ReadDir("."); err == nil {
		t.Error("ReadDir on unopened source should fail")
	}
	if gs.Exists(".") {
		t.Error("Exists on unopened source should return false")
	}
	if err := gs.WriteDir(".", map[string][]byte{}); err == nil {
		t.Error("WriteDir on unopened source should fail")
	}
}

func TestGitSourceClose(t *testing.T) {
	gs := NewGitSource("https://github.com/user/repo.git")
	// Close without Open should be safe
	if err := gs.Close(); err != nil {
		t.Errorf("Close() on unopened source: %v", err)
	}
}
