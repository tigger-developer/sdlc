// ABOUTME: Verifies evidence-scoped Claude permission settings at the invocation boundary.
// ABOUTME: Tests local argument construction only, never a metered provider.
package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

func TestW011ClaudeGrantsExactEvidenceReadsOnStartAndResume(t *testing.T) {
	root := t.TempDir()
	realDir := filepath.Join(root, "real")
	aliasDir := filepath.Join(root, "alias")
	if err := os.Mkdir(realDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realDir, aliasDir); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(aliasDir, "a file [draft]*.org")
	if err := os.WriteFile(file, []byte("evidence"), 0o600); err != nil {
		t.Fatal(err)
	}
	canonical, err := filepath.EvalSymlinks(file)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Read(/" + filepath.ToSlash(filepath.Dir(file)) + "/a file \\[draft\\]\\*.org)"}
	if canonical == file {
		t.Fatal("fixture did not create an aliased path")
	}
	want = append(want, "Read(/"+filepath.ToSlash(filepath.Dir(canonical))+"/a file \\[draft\\]\\*.org)")
	request := matrixRequest("claude", t.TempDir())
	request.Evidence = []EvidenceFile{{Path: file}}
	for _, build := range []func(Request) (Invocation, error){BuildStart, BuildResume} {
		invocation, err := build(request)
		if err != nil {
			t.Fatal(err)
		}
		i := slices.Index(invocation.Args, "--settings")
		if i < 0 || i+1 >= len(invocation.Args) {
			t.Fatal("evidence read grants missing")
		}
		if invocation.Args[len(invocation.Args)-1] != request.Prompt || i+2 >= len(invocation.Args) {
			t.Fatal("permission options must precede the positional prompt")
		}
		var settings map[string]map[string][]string
		if err := json.Unmarshal([]byte(invocation.Args[i+1]), &settings); err != nil {
			t.Fatal(err)
		}
		if len(settings) != 1 || len(settings["permissions"]) != 1 || !reflect.DeepEqual(settings["permissions"]["allow"], want) {
			t.Fatalf("settings = %#v, want only exact Read grants %v", settings, want)
		}
		for flag, value := range map[string]string{"--tools": "Read", "--permission-mode": "plan"} {
			i := slices.Index(invocation.Args, flag)
			if i < 0 || invocation.Args[i+1] != value {
				t.Fatalf("lost %s %s", flag, value)
			}
		}
		if slices.Contains(invocation.Args, "--add-dir") || slices.Contains(invocation.Args, "--dangerously-skip-permissions") {
			t.Fatal("broadened access")
		}
	}
}

func TestW011ClaudeRejectsUnrepresentableEvidencePaths(t *testing.T) {
	for _, path := range []string{"relative.org", "/tmp/line\nbreak.org", "/tmp/.env"} {
		request := matrixRequest("claude", t.TempDir())
		request.Evidence = []EvidenceFile{{Path: path}}
		for _, build := range []func(Request) (Invocation, error){BuildStart, BuildResume} {
			if _, err := build(request); err == nil {
				t.Fatalf("accepted unsafe evidence path %q", path)
			}
		}
	}
}
