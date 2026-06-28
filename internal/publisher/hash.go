package publisher

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

// computeHash computes a SHA-256 hash over all component files.
// Files are sorted by RelPath for deterministic output.
func computeHash(files map[string][]byte) string {
	// Sort file paths for deterministic ordering
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	h := sha256.New()
	for _, p := range paths {
		h.Write([]byte(p))
		h.Write(files[p])
	}
	return hex.EncodeToString(h.Sum(nil))
}
