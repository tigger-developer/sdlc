// ABOUTME: Verifies private diagnostic bounds and anonymous caller liveness.
// ABOUTME: Uses local files and synthetic heartbeat messages only.
package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

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
	if output.String() != "Audit in progress.\nAudit in progress.\n" {
		t.Fatalf("missing anonymous liveness: %q", &output)
	}
	info, err := os.Stat(log.file.Name())
	if err != nil || info.Size() != 4 || info.Mode().Perm() != 0o600 {
		t.Fatalf("private capped log: %v %v", info, err)
	}
	if strings.Contains(output.String(), "private") {
		t.Fatal("session leaked")
	}
}
