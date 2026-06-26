package source

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// LocalSource implements Source backed by a local directory.
type LocalSource struct {
	root string // absolute path
}

// NewLocalSource creates a LocalSource rooted at the given directory.
func NewLocalSource(root string) (*LocalSource, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}
	return &LocalSource{root: abs}, nil
}

func (s *LocalSource) Root() string { return s.root }

func (s *LocalSource) ReadDir(relPath string) (map[string][]byte, error) {
	dir := filepath.Join(s.root, relPath)
	files := make(map[string][]byte)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return fmt.Errorf("relative path: %w", err)
		}
		// Use forward slashes for consistency
		rel = filepath.ToSlash(rel)
		files[rel] = data
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory not found: %s", relPath)
		}
		return nil, err
	}
	return files, nil
}

func (s *LocalSource) Exists(relPath string) bool {
	_, err := os.Stat(filepath.Join(s.root, relPath))
	return err == nil
}

func (s *LocalSource) WriteDir(relPath string, files map[string][]byte) error {
	dir := filepath.Join(s.root, relPath)
	for name, content := range files {
		fullPath := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			return fmt.Errorf("write %s: %w", fullPath, err)
		}
	}
	return nil
}

func (s *LocalSource) Close() error { return nil } // no-op
