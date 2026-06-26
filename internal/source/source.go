package source

// Source abstracts access to a component data source (local directory or Git repo).
type Source interface {
	// Root returns the absolute path to the source root directory.
	Root() string

	// ReadDir reads all files under the given relative path recursively.
	// Returns a map of filename → content for all files found.
	// Skips directories (only returns files).
	ReadDir(relPath string) (map[string][]byte, error)

	// Exists checks whether a file or directory exists at the given relative path.
	Exists(relPath string) bool

	// WriteDir writes files under the given relative path (used by publish).
	// Creates intermediate directories as needed.
	WriteDir(relPath string, files map[string][]byte) error

	// Close releases any resources held by the source (e.g. temp directories).
	Close() error
}
