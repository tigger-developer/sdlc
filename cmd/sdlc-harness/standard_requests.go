// ABOUTME: Serves installed audit standards on demand to tools-disabled providers.
// ABOUTME: Restricts requests to the captured, hash-verified SDLC inventory.
package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tigger-developer/sdlc/internal/harness"
)

const maximumStandardFetches = 4
const standardRequestPrefix = "STANDARDS-REQUEST: "

func (r recoveryRun) standardInventory() (map[string]string, []string, string, error) {
	standards := map[string]string{}
	var projectFiles []string
	var listing strings.Builder
	for _, file := range r.evidence.Files {
		relative, err := filepath.Rel(r.standardsRoot, file.Path)
		if err != nil {
			return nil, nil, "", err
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			projectFiles = append(projectFiles, file.Path)
			continue
		}
		name := filepath.ToSlash(relative)
		standards[name] = file.Path
		fmt.Fprintf(&listing, "- %s sha256:%s\n", name, file.SHA256)
	}
	return standards, projectFiles, listing.String(), nil
}

func (r recoveryRun) requestedStandardPaths(response string, already map[string]bool) ([]string, bool, error) {
	if !strings.HasPrefix(response, standardRequestPrefix) {
		return nil, false, nil
	}
	standards, _, _, err := r.standardInventory()
	if err != nil {
		return nil, true, err
	}
	var names []string
	if err := json.Unmarshal([]byte(strings.TrimPrefix(response, standardRequestPrefix)), &names); err != nil {
		return nil, true, fmt.Errorf("invalid standards request: %w", err)
	}
	if len(names) == 0 || len(names) > len(standards) {
		return nil, true, fmt.Errorf("standards request must name between one and %d documents", len(standards))
	}
	paths := make([]string, 0, len(names))
	batch := map[string]bool{}
	for _, name := range names {
		path, allowed := standards[name]
		if !allowed || already[name] || batch[name] {
			return nil, true, fmt.Errorf("unavailable or repeated audit standard %q", name)
		}
		paths = append(paths, path)
		batch[name] = true
	}
	return paths, true, nil
}

func standardsRequestIncident(harnessName, sessionID string, err error) error {
	return &harness.Incident{Kind: "response-malformed", Harness: harnessName, SessionID: sessionID, Err: err}
}
