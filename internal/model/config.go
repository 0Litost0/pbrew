package model

// Config represents the global configuration stored in ~/.pbrew/config.json.
type Config struct {
	Source string `json:"source"` // Data source URL (Git) or local directory path
}
