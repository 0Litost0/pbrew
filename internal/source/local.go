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
		if err := atomicWriteFile(fullPath, content, 0644); err != nil {
			return fmt.Errorf("write %s: %w", fullPath, err)
		}
	}
	return nil
}

func (s *LocalSource) Close() error { return nil } // no-op

// atomicWriteFile writes data to path atomically using write-temp-then-rename.
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".pbrew-tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, perm); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}
