package model

import (
	"encoding/json"
	"testing"
)

func TestRecordStoreJSONRoundTrip(t *testing.T) {
	rs := RecordStore{
		Version: 1,
		Records: []InstallRecord{
			{
				Name:    "git-conventions",
				Version: "1.0.0",
				Type:    TypeRule,
				Agent:   "codebuddy",
				Scope:   "project",
				Source:  "https://github.com/user/repo.git",
				Files: []InstalledFile{
					{RelPath: "rules/git-conventions/git-conventions.md"},
				},
			},
		},
	}

	data, err := json.Marshal(rs)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got RecordStore
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Version != rs.Version {
		t.Errorf("Version: got %d, want %d", got.Version, rs.Version)
	}
	if len(got.Records) != 1 {
		t.Fatalf("Records: got %d, want 1", len(got.Records))
	}
	r := got.Records[0]
	if r.Name != "git-conventions" {
		t.Errorf("Name: got %q, want %q", r.Name, "git-conventions")
	}
	if len(r.Files) != 1 || r.Files[0].RelPath != "rules/git-conventions/git-conventions.md" {
		t.Errorf("Files: got %v, want one file with correct path", r.Files)
	}
}
