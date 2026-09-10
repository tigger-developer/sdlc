package harness

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeEvidence(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestEvidenceManifestIsStableAndCreatesNoFiles(t *testing.T) {
	root := t.TempDir()
	a, b := filepath.Join(root, "a.org"), filepath.Join(root, "b.org")
	writeEvidence(t, a, "a")
	writeEvidence(t, b, "b")
	first, err := CaptureEvidence(root, []string{"b.org", "a.org"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := CaptureEvidence(root, []string{a, b})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("order changed manifest: %#v / %#v", first, second)
	}
	if first.Files[0].SHA256 != "ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb" {
		t.Fatalf("wrong SHA-256: %#v", first.Files[0])
	}
	files, err := os.ReadDir(root)
	if err != nil || len(files) != 2 {
		t.Fatalf("capture created files: %v, %v", files, err)
	}
	changes := second.Changes(first.Files)
	for _, change := range changes {
		if change.Change != "unchanged" {
			t.Fatalf("unchanged files: %#v", changes)
		}
	}
	writeEvidence(t, a, "changed")
	c := filepath.Join(root, "c.org")
	writeEvidence(t, c, "c")
	third, err := CaptureEvidence(root, []string{c, a})
	if err != nil {
		t.Fatal(err)
	}
	changes = third.Changes(first.Files)
	for i, want := range []string{"changed", "removed", "added"} {
		if changes[i].Change != want {
			t.Fatalf("changes = %#v", changes)
		}
	}
	// An omitted source need not have been deleted.
	if _, err := os.Stat(b); err != nil {
		t.Fatal(err)
	}
}

func TestEvidenceRejectsInvalidInputs(t *testing.T) {
	root := t.TempDir()
	writeEvidence(t, filepath.Join(root, "valid"), "evidence")
	if err := os.Symlink(filepath.Join(root, "valid"), filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	for _, paths := range [][]string{nil, {"missing"}, {"."}, {"link"}, {"valid", "./valid"}, {".env"}} {
		if _, err := CaptureEvidence(root, paths); err == nil {
			t.Errorf("invalid inputs accepted: %v", paths)
		}
	}
	large := filepath.Join(root, "large")
	file, err := os.Create(large)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(maximumInputBytes + 1); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := CaptureEvidence(root, []string{large}); err == nil {
		t.Fatal("oversize accepted")
	}
	var paths []string
	for _, name := range []string{"one", "two", "three", "four", "five"} {
		path := filepath.Join(root, name)
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Truncate(maximumInputBytes); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	if _, err := CaptureEvidence(root, paths); err == nil || !strings.Contains(err.Error(), "total size") {
		t.Fatalf("total size limit = %v", err)
	}
}

func TestEvidenceMutationCannotProduceAcceptedResult(t *testing.T) {
	for _, when := range []string{"before", "during", "deleted", "symlink"} {
		t.Run(when, func(t *testing.T) {
			root := t.TempDir()
			source := filepath.Join(root, "spec.org")
			writeEvidence(t, source, "original")
			evidence, err := CaptureEvidence(root, []string{source})
			if err != nil {
				t.Fatal(err)
			}
			request := Request{Harness: "codex", Model: "fixture", Directory: root, ResultFile: filepath.Join(root, "result")}
			if when == "before" {
				writeEvidence(t, source, "changed")
			}
			called := false
			result, err := Execute(context.Background(), request, false, &evidence, func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, stdout, _ io.Writer) error {
				called = true
				switch when {
				case "during":
					writeEvidence(t, source, "changed")
				case "deleted", "symlink":
					if err := os.Remove(source); err != nil {
						return err
					}
					if when == "symlink" {
						writeEvidence(t, filepath.Join(root, "replacement"), "original")
						if err := os.Symlink(filepath.Join(root, "replacement"), source); err != nil {
							return err
						}
					}
				}
				writeEvidence(t, request.ResultFile, "GATE: implementation\nREVISION: x\nVERDICT: PASS\n")
				_, err := io.WriteString(stdout, "{\"type\":\"thread.started\",\"thread_id\":\"session\"}\n")
				return err
			}, io.Discard)
			if err == nil || result.Response != "" {
				t.Fatalf("changed evidence accepted: %#v, %v", result, err)
			}
			if when == "before" && called {
				t.Fatal("provider ran after preflight failure")
			}
		})
	}
}

func TestOldAuditRecordHasNoEvidenceBaseline(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "audits.yaml")
	writeEvidence(t, path, "version: 1\naudits:\n- work_item: W001\n  gate: definition\n  session_id: retained\n  history:\n  - round: 1\n    verdict: FAIL\n")
	entry, found, err := ReadAuditEntry(path, "W001", "definition")
	if err != nil || !found || entry.SessionID != "retained" {
		t.Fatalf("old record: %#v, %v", entry, err)
	}
	evidence := Evidence{Files: []EvidenceFile{{Path: "/project/spec.org", SHA256: "hash"}}}
	if change := evidence.Changes(entry.LatestEvidence())[0].Change; change != "added" {
		t.Fatal(change)
	}
	prompt, err := evidence.Prompt(entry.LatestEvidence())
	if err != nil || !strings.Contains(prompt, "/project/spec.org") {
		t.Fatalf("prompt = %q, %v", prompt, err)
	}
}
