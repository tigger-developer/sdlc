// ABOUTME: Verifies private diagnostic bounds and anonymous caller liveness.
// ABOUTME: Uses local files and synthetic heartbeat messages only.
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiagnosticsRejectSymlinkedDestinations(t *testing.T) {
	for _, kind := range []string{"standalone", "managed-parent", "managed-leaf"} {
		t.Run(kind, func(t *testing.T) {
			root, outside := t.TempDir(), t.TempDir()
			alias := filepath.Join(root, ".audit-diagnostics")
			if kind == "managed-parent" {
				alias = filepath.Join(root, ".sdlc")
			}
			if kind == "managed-leaf" {
				if err := os.Mkdir(filepath.Join(root, ".sdlc"), 0700); err != nil {
					t.Fatal(err)
				}
				alias = filepath.Join(root, ".sdlc", "audit-diagnostics")
			}
			if err := os.Symlink(outside, alias); err != nil {
				t.Fatal(err)
			}
			var log *auditDiagnostics
			var err error
			if kind == "standalone" {
				log, err = openStandaloneDiagnostics(root, &bytes.Buffer{})
			} else {
				log, err = openAuditDiagnostics(root, &bytes.Buffer{})
			}
			if log != nil {
				_ = log.file.Close()
			}
			if err == nil {
				t.Error("accepted symlinked diagnostic destination")
			}
			entries, readErr := os.ReadDir(outside)
			if readErr != nil || len(entries) != 0 {
				t.Fatalf("wrote outside project: %v %v", entries, readErr)
			}
		})
	}
}

func TestDiagnosticsRejectUnsafePermissionsWithoutChangingThem(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, ".audit-diagnostics")
	// #nosec G301 -- deliberately public fixture verifies refusal without chmod.
	if err := os.Mkdir(directory, 0755); err != nil {
		t.Fatal(err)
	}
	log, err := openStandaloneDiagnostics(root, &bytes.Buffer{})
	if log != nil {
		_ = log.file.Close()
	}
	if err == nil {
		t.Fatal("accepted public diagnostic directory")
	}
	info, statErr := os.Stat(directory)
	if statErr != nil || info.Mode().Perm() != 0755 {
		t.Fatalf("existing permissions changed: %v %v", info, statErr)
	}
	entries, readErr := os.ReadDir(directory)
	if readErr != nil || len(entries) != 0 {
		t.Fatalf("unsafe destination received logs: %v %v", entries, readErr)
	}
}

func TestDiagnosticsRejectWritableParentsBeforeCreatingLogs(t *testing.T) {
	for _, kind := range []string{"project-root", "runtime-parent"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			parent := root
			if kind == "runtime-parent" {
				parent = filepath.Join(root, ".sdlc")
				if err := os.Mkdir(parent, 0700); err != nil {
					t.Fatal(err)
				}
			}
			// #nosec G302 -- deliberately unsafe temporary directory verifies refusal.
			if err := os.Chmod(parent, 0777); err != nil {
				t.Fatal(err)
			}
			log, err := openAuditDiagnostics(root, &bytes.Buffer{})
			if log != nil {
				_ = log.file.Close()
			}
			if err == nil {
				t.Fatal("accepted parent writable by another local user")
			}
			if _, err := os.Stat(filepath.Join(root, ".sdlc", "audit-diagnostics")); !os.IsNotExist(err) {
				t.Fatalf("created logs below unsafe parent: %v", err)
			}
		})
	}
}

func TestAuditDiagnosticsKeepsHeartbeatAfterLogCap(t *testing.T) {
	var output bytes.Buffer
	log, err := openAuditDiagnostics(t.TempDir(), &output)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = log.file.Close() })
	log.remaining = 4
	for _, body := range []string{
		"sdlc-harness: external context started (harness=fixture session=private); waiting for final response\n",
		"sdlc-harness: external context still running (harness=fixture session=private elapsed=30s)\n",
	} {
		if n, err := log.Write([]byte(body)); err != nil || n != len(body) {
			t.Fatalf("write=%d, %v", n, err)
		}
	}
	if strings.Count(output.String(), "Audit in progress.\n") != 2 {
		t.Fatalf("missing anonymous liveness: %q", &output)
	}
	if strings.Count(output.String(), "AUDIT INPUTS FROZEN:") != 1 || !strings.HasPrefix(output.String(), "AUDIT INPUTS FROZEN:") {
		t.Fatalf("startup must warn once before liveness, even after log cap: %q", &output)
	}
	info, err := os.Stat(log.file.Name())
	if err != nil || info.Size() != 4 || info.Mode().Perm() != 0o600 {
		t.Fatalf("private capped log: %v %v", info, err)
	}
	if strings.Contains(output.String(), "private") {
		t.Fatal("session leaked")
	}
}

func TestAuditStatusRemainsVisibleAfterLogCap(t *testing.T) {
	var output bytes.Buffer
	log, err := openAuditDiagnostics(t.TempDir(), &output)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = log.file.Close() })
	log.remaining = 0
	if _, err := log.Write([]byte("Audit primary failed: secret-body\nTrying the configured fallback.\n")); err != nil {
		t.Fatal(err)
	}
	if output.Len() != 0 {
		t.Fatal("native output forged a public status")
	}
	if err := reportAuditStatus(log, "Audit fallback failed: provider usage quota is exhausted."); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "usage quota") || strings.Contains(output.String(), "secret-body") {
		t.Fatalf("unsafe or missing status: %q", output.String())
	}
}
