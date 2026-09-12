package projectinit

import (
	"bytes"
	"errors"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// RT013.1: exercise the real migration coordinator with local command doubles;
// no regression test invokes a metered provider.
func TestMigrationAndConversionUseDifferentConfiguredAgents(t *testing.T) {
	project := t.TempDir()
	writeProjectTestFile(t, filepath.Join(project, "docs", "ACs.md"), "# Existing ACs\n")
	var calls int
	options := Options{Output: &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{},
		RunCommand: func(name string, args []string, directory string, input io.Reader, stdout, stderr io.Writer) error {
			calls++
			if directory != project {
				t.Fatalf("migration cwd = %q", directory)
			}
			var want []string
			var skill string
			switch calls {
			case 1:
				if name != "hermes" {
					t.Fatalf("ticket migration invoked %q", name)
				}
				want = []string{"chat", "--query-file", "-", "--oneshot", "--in", project, "--model", "bulk-model", "--provider", "bulk-provider"}
				skill = "$migrate-legacy-acs-to-sdlc-v1"
			case 2:
				if name != "codex" {
					t.Fatalf("AC conversion invoked %q", name)
				}
				want = []string{"exec", "--ephemeral", "--sandbox", "workspace-write", "--model", "definition-model", "-"}
				skill = "$convert-migrated-acs-to-org"
				writeProjectTestFile(t, filepath.Join(project, "docs", "ACs.org"), canonicalLegacyLedger(""))
			default:
				t.Fatalf("unexpected additional invocation %d", calls)
			}
			if !reflect.DeepEqual(args, want) {
				t.Fatalf("arguments = %#v, want %#v", args, want)
			}
			if input == nil {
				t.Fatal("migration prompt missing from stdin")
			}
			prompt, err := io.ReadAll(input)
			if err != nil {
				return err
			}
			if !strings.Contains(string(prompt), skill) {
				t.Fatalf("prompt lacks %s: %s", skill, prompt)
			}
			return nil
		}}
	err := prepareLegacyProject(options, project, map[string]string{
		"SDLC_AUDIT_HARNESS": "hermes", "SDLC_AUDIT_MODEL": "bulk-model", "SDLC_AUDIT_PROVIDER": "bulk-provider",
		"SDLC_SPEC_HARNESS": "codex", "SDLC_SPEC_MODEL": "definition-model", "SDLC_SPEC_PROVIDER": "ignored-native-provider",
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("agent calls = %d", calls)
	}
}

// RT013.2: a failed ticket migration must not advance to ledger conversion.
func TestMigrationAgentFailureStopsPreparation(t *testing.T) {
	project := t.TempDir()
	writeProjectTestFile(t, filepath.Join(project, "docs", "ACs.md"), "# Existing ACs\n")
	want := errors.New("native CLI failed")
	calls := 0
	err := prepareLegacyProject(Options{Output: &bytes.Buffer{}, RunCommand: func(string, []string, string, io.Reader, io.Writer, io.Writer) error {
		calls++
		return want
	}}, project, map[string]string{"SDLC_AUDIT_HARNESS": "hermes", "SDLC_AUDIT_PROVIDER": "nous"}, true)
	if !errors.Is(err, want) || calls != 1 {
		t.Fatalf("error = %v; calls = %d", err, calls)
	}
}

// RT013.3: only Hermes consumes a provider; native command policies stay enabled.
func TestDirectMigrationAgentArguments(t *testing.T) {
	for _, tc := range []struct {
		name, provider string
		args           []string
		wantError      bool
	}{
		{"claude", "ignored", []string{"--print", "--permission-mode", "acceptEdits", "--model", "writer-model"}, false},
		{"hermes", "", nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			err := runConfiguredMutationSkill(Options{Output: &bytes.Buffer{}, RunCommand: func(name string, args []string, directory string, input io.Reader, stdout, stderr io.Writer) error {
				calls++
				if name != tc.name || directory != "/project" || !reflect.DeepEqual(args, tc.args) {
					t.Fatalf("command = %s %v in %s", name, args, directory)
				}
				prompt, err := io.ReadAll(input)
				if err != nil {
					return err
				}
				if string(prompt) != "literal $(not-a-shell-command)" {
					t.Fatalf("prompt changed: %q", prompt)
				}
				return nil
			}}, "/project", map[string]string{"SDLC_SPEC_HARNESS": tc.name, "SDLC_SPEC_MODEL": "writer-model", "SDLC_SPEC_PROVIDER": tc.provider}, "SDLC_SPEC", "convert-migrated-acs-to-org", "literal $(not-a-shell-command)")
			if (err != nil) != tc.wantError {
				t.Fatalf("error = %v", err)
			}
			if tc.wantError && calls != 0 {
				t.Fatal("invalid provider launched a process")
			}
		})
	}
}
