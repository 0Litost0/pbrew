package source

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GitSource implements Source backed by a Git repository.
// It clones the repo to a temporary directory on Open and cleans up on Close.
type GitSource struct {
	url     string // Git repository URL
	tempDir string // temporary clone directory
	keep    bool   // if true, Close will NOT remove tempDir
}

// NewGitSource creates a GitSource for the given repository URL.
// The actual clone happens in Open; this constructor just records the URL.
func NewGitSource(url string) *GitSource {
	return &GitSource{url: url}
}

// Open clones the repository into a temporary directory using shallow clone.
func (s *GitSource) Open() error {
	tmpDir, err := os.MkdirTemp("", "pbrew-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

	cmd := exec.Command("git", "clone", "--depth", "1", s.url, tmpDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("git clone failed: %w\n%s", err, output)
	}

	s.tempDir = tmpDir
	return nil
}

// Keep marks the temporary directory so it will NOT be removed on Close.
// Call this after a successful publish so the user can cd in and git push.
func (s *GitSource) Keep() {
	s.keep = true
}

func (s *GitSource) Root() string { return s.tempDir }

func (s *GitSource) ReadDir(relPath string) (map[string][]byte, error) {
	if s.tempDir == "" {
		return nil, fmt.Errorf("git source not opened yet")
	}
	// Delegate to LocalSource which shares the same ReadDir logic
	ls := &LocalSource{root: s.tempDir}
	return ls.ReadDir(relPath)
}

func (s *GitSource) Exists(relPath string) bool {
	if s.tempDir == "" {
		return false
	}
	ls := &LocalSource{root: s.tempDir}
	return ls.Exists(relPath)
}

func (s *GitSource) WriteDir(relPath string, files map[string][]byte) error {
	if s.tempDir == "" {
		return fmt.Errorf("git source not opened yet")
	}
	ls := &LocalSource{root: s.tempDir}
	return ls.WriteDir(relPath, files)
}

// Close removes the temporary clone directory, unless Keep() was called.
func (s *GitSource) Close() error {
	if s.tempDir == "" {
		return nil
	}
	if s.keep {
		// Preserve tempDir so Root() still returns the path after Close.
		return nil
	}
	err := os.RemoveAll(s.tempDir)
	s.tempDir = ""
	return err
}

// IsGitURL checks whether a string looks like a Git URL.
func IsGitURL(s string) bool {
	return !filepath.IsAbs(s) && isGitURL(s)
}

func isGitURL(s string) bool {
	if len(s) >= 4 && s[:4] == "git@" {
		return true
	}
	if len(s) >= 5 && s[:5] == "https" {
		return true
	}
	if len(s) >= 6 && s[:6] == "git://" {
		return true
	}
	if len(s) >= 4 && s[len(s)-4:] == ".git" {
		return true
	}
	return false
}

// OpenSource is a convenience function that detects the source type
// (local directory vs Git URL) and returns the appropriate Source.
// For Git URLs, it also calls Open() to perform the clone.
func OpenSource(source string) (Source, error) {
	if IsGitURL(source) {
		gs := NewGitSource(source)
		if err := gs.Open(); err != nil {
			return nil, err
		}
		return gs, nil
	}
	return NewLocalSource(source)
}
