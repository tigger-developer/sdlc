package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateLegacyAuditPreservesContentAndRemovesSource(t *testing.T) {
	dir := t.TempDir()
	record := filepath.Join(dir, "audits.yaml")
	legacy := filepath.Join(dir, "audits.org")
	content := "* Historical audit\n- Findings: preserved\n"
	if err := os.WriteFile(legacy, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := MigrateLegacyAudit(record); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatalf("legacy audit still exists: %v", err)
	}
	contents, err := os.ReadFile(record)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), "Findings: preserved") {
		t.Fatalf("legacy content was not preserved: %s", contents)
	}
	if err := MigrateLegacyAudit(record); err != nil {
		t.Fatalf("second migration: %v", err)
	}
}
