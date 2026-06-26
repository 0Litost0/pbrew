package installer

import (
	"fmt"

	"github.com/user/pbrew/internal/agent"
	"github.com/user/pbrew/internal/model"
	"github.com/user/pbrew/internal/platform"
	"github.com/user/pbrew/internal/registry"
	"github.com/user/pbrew/internal/source"

	"github.com/Masterminds/semver/v3"
)

// UpdateParams holds parameters for updating components.
type UpdateParams struct {
	ComponentNames []string // empty means all (when All is true)
	Agent          string
	Scope          string
	All            bool
	Source         source.Source
}

// Update compares installed versions with source versions and re-installs if newer.
func Update(params UpdateParams) error {
	// 1. Validate agent
	if err := agent.Validate(params.Agent); err != nil {
		return err
	}
	profile, _ := agent.Get(params.Agent)

	// 2. Determine project root if needed
	var projectRoot string
	var err error
	if params.Scope == "project" {
		projectRoot, err = platform.FindProjectRoot()
		if err != nil {
			return err
		}
	}

	// 3. Read install records
	recordPath, err := platform.RecordFilePath(profile, params.Scope, projectRoot)
	if err != nil {
		return err
	}

	store, err := loadRecordStore(recordPath)
	if err != nil {
		return err
	}

	// 4. Filter records by agent/scope/name
	var targets []model.InstallRecord
	for _, r := range store.Records {
		if r.Agent != params.Agent || r.Scope != params.Scope {
			continue
		}
		if !params.All {
			found := false
			for _, name := range params.ComponentNames {
				if r.Name == name {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		targets = append(targets, r)
	}

	if len(targets) == 0 {
		fmt.Println("No components to update.")
		return nil
	}

	// 5. Check each target for updates
	anyUpdated := false
	for _, r := range targets {
		entry, err := registry.FindComponent(params.Source, r.Name)
		if err != nil {
			fmt.Printf("Warning: component %s not found in source, skipping\n", r.Name)
			continue
		}

		cmp, err := semverCompare(r.Version, entry.Version)
		if err != nil {
			fmt.Printf("Warning: version comparison failed for %s: %v\n", r.Name, err)
			continue
		}

		if cmp >= 0 {
			fmt.Printf("%s: already up-to-date (%s)\n", r.Name, r.Version)
			continue
		}

		// 6. Re-install (will overwrite via Force=true)
		fmt.Printf("%s: updating %s → %s\n", r.Name, r.Version, entry.Version)
		if err := Install(InstallParams{
			ComponentName: r.Name,
			Agent:         params.Agent,
			Scope:         params.Scope,
			Source:        params.Source,
			Force:         true,
		}); err != nil {
			fmt.Printf("Error updating %s: %v\n", r.Name, err)
			continue
		}
		anyUpdated = true
	}

	if !anyUpdated {
		fmt.Println("All components are up-to-date.")
	}

	return nil
}

// semverCompare compares two semantic versions.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
func semverCompare(a, b string) (int, error) {
	va, err := semver.NewVersion(a)
	if err != nil {
		return 0, fmt.Errorf("parse version %q: %w", a, err)
	}
	vb, err := semver.NewVersion(b)
	if err != nil {
		return 0, fmt.Errorf("parse version %q: %w", b, err)
	}

	if va.LessThan(vb) {
		return -1, nil
	}
	if va.GreaterThan(vb) {
		return 1, nil
	}
	return 0, nil
}
