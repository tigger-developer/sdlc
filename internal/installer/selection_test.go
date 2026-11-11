package installer

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestAgentSelectionUsesLegacyConfigurationBeforeFirstMigration(t *testing.T) {
	home := t.TempDir()
	installRoot := filepath.Join(t.TempDir(), "sdlc")
	for _, name := range []string{".codex", ".claude", ".hermes"} {
		if err := os.MkdirAll(filepath.Join(home, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeFixtureFile(t, filepath.Join(home, ".agents", "sdlc.yaml"), "version: 3\nactive-harness:\n  - codex\n  - hermes\n")

	var output bytes.Buffer
	selected, err := selectInteractiveAgents(home, installRoot, bufio.NewReader(strings.NewReader("")), &output)
	if err != nil || !reflect.DeepEqual(selected, []string{"codex", "hermes"}) {
		t.Fatalf("legacy selection = %v, err = %v", selected, err)
	}
	if !strings.Contains(output.String(), filepath.Join(home, ".agents", "sdlc.yaml")) || strings.Contains(output.String(), "Install the SDLC for which agents?") {
		t.Fatalf("legacy selection prompted: %q", output.String())
	}

	writeFixtureFile(t, filepath.Join(installRoot, "sdlc.yaml"), "version: 3\nactive-harness:\n  - claude\n")
	output.Reset()
	selected, err = selectInteractiveAgents(home, installRoot, bufio.NewReader(strings.NewReader("")), &output)
	if err != nil || !reflect.DeepEqual(selected, []string{"claude"}) {
		t.Fatalf("canonical selection = %v, err = %v", selected, err)
	}
}
