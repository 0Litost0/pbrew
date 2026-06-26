package model

// InstalledFile records the relative path of a single installed file
// (relative to the Agent root directory).
type InstalledFile struct {
	RelPath string `json:"relPath"`
}

// HookHandlerRecord records the location of a single installed Hook handler
// for precise removal during uninstall.
type HookHandlerRecord struct {
	EventName    string `json:"eventName"`
	Matcher      string `json:"matcher"`
	HandlerIndex int    `json:"handlerIndex"`
}

// InstallRecord records information about a single installed component.
type InstallRecord struct {
	Name         string              `json:"name"`
	Version      string              `json:"version"`
	Type         ComponentType       `json:"type"`
	Agent        string              `json:"agent"`
	Scope        string              `json:"scope"`
	Source       string              `json:"source"`
	Files        []InstalledFile     `json:"files,omitempty"`
	HookHandlers []HookHandlerRecord `json:"hookHandlers,omitempty"`
}

// RecordStore represents the full structure of .pbrew/_config.json.
type RecordStore struct {
	Version int             `json:"version"`
	Records []InstallRecord `json:"records"`
}
