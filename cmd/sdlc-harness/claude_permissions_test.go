// ABOUTME: Checks exact external evidence grants through the CLI start/resume boundary.
// ABOUTME: Uses a local executable fixture, not a metered Claude invocation.
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/tigger-developer/sdlc/internal/harness"
)

func TestW011CLIProjectsEvidencePermissionsAndRetainsSession(t *testing.T) {
	root := t.TempDir()
	external := filepath.Join(t.TempDir(), "authority.org")
	probe := filepath.Join(root, "settings.json")
	identity := filepath.Join(root, "identity")
	t.Setenv("PROBE_SETTINGS", probe)
	t.Setenv("PROBE_IDENTITY", identity)
	t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
	script := `#!/bin/sh
previous=
for argument in "$@"; do
    if [ "$previous" = "--settings" ]; then printf '%s' "$argument" > "$PROBE_SETTINGS"; fi
    if [ "$previous" = "--session-id" ] || [ "$previous" = "--resume" ]; then printf '%s' "$argument" > "$PROBE_IDENTITY"; fi
    previous="$argument"
done
printf '%s\n' '{"result":"GATE: definition\nREVISION: fixture\nVERDICT: PASS\n"}'
`
	// #nosec G306 -- this local process-boundary fixture must be executable.
	if err := os.WriteFile(filepath.Join(root, "claude"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(external, []byte("external authority"), 0o600); err != nil {
		t.Fatal(err)
	}
	registry := filepath.Join(root, "prompts.yaml")
	if err := os.WriteFile(registry, []byte("version: 1\ngates:\n  definition:\n    prompt: audit\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	record := filepath.Join(root, "audits.yaml")
	args := []string{"--phase", "audit", "--gate", "definition", "--harness", "claude", "--model", "fixture", "--project", root, "--global-config", filepath.Join(root, "absent.yaml"), "--audit-prompts", registry, "--audit-record", record, "--work-item", "W011-evidence", "--input", external}
	canonical, err := filepath.EvalSymlinks(external)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Read(/" + external + ")"}
	if canonical != external {
		want = append(want, "Read(/"+canonical+")")
	}
	var firstID string
	for _, action := range []string{"start", "resume"} {
		if err := run(append([]string{action}, args...), strings.NewReader("review"), &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(probe)
		if err != nil {
			t.Fatal(err)
		}
		var settings map[string]map[string][]string
		if err := json.Unmarshal(data, &settings); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(settings["permissions"]["allow"], want) {
			t.Fatalf("grants = %s, want %v", data, want)
		}
		entry, found, err := harness.ReadAuditEntry(record, "W011-evidence", "definition")
		if err != nil || !found {
			t.Fatalf("audit record missing: %v", err)
		}
		invoked, err := os.ReadFile(identity)
		if err != nil {
			t.Fatal(err)
		}
		if action == "start" {
			firstID = entry.SessionID
		}
		if firstID == "" || entry.SessionID != firstID || string(invoked) != firstID {
			t.Fatal("resume changed native identity")
		}
	}
}
