// ABOUTME: Verifies gate-local verdict accounting independently of provider attempts.
// ABOUTME: Protects incident exclusion and historical reset boundaries.
package harness

import (
	"path/filepath"
	"testing"
)

func TestVerdictBudgetExcludesIncidentsAndOtherGates(t *testing.T) {
	entry := AuditEntry{Gate: "implementation", ReadinessCheck: "delivery-code", ExternalRound: 6,
		History: []AuditRound{
			{Round: 1, ReadinessCheck: "test-code", Verdict: "FAIL"},
			{Round: 2, ReadinessCheck: "test-code", Verdict: "PASS"},
			{Round: 3, ReadinessCheck: "delivery-code", Incident: "timeout"},
			{Round: 4, ReadinessCheck: "delivery-code", Incident: "response-malformed"},
			{Round: 5, ReadinessCheck: "delivery-code", Verdict: "FAIL"},
			{Round: 6, ReadinessCheck: "delivery-code", Verdict: "PROVISIONAL PASS"},
		}}
	if got := entry.RoundsUsed(); got != 2 {
		t.Fatalf("delivery verdicts = %d, want 2", got)
	}
	entry.BudgetResets = []AuditBudgetReset{{AfterRound: 5}}
	if got := entry.RoundsUsed(); got != 1 {
		t.Fatalf("verdicts since historical reset = %d, want 1", got)
	}
}

func TestLegacyVerdictRequiresMatchingGateEvidence(t *testing.T) {
	for _, selected := range []string{"test-code", "delivery-code"} {
		for _, recorded := range []string{"", "implementation", "test-code", "delivery-code"} {
			entry := AuditEntry{Gate: "implementation", ReadinessCheck: selected, ExternalRound: 1, Verdict: "FAIL"}
			if recorded != "" {
				entry.Response = "GATE: " + recorded + "\nREVISION: old\nVERDICT: FAIL"
			}
			want := 0
			if recorded == selected {
				want = 1
			}
			if got := entry.RoundsUsed(); got != want {
				t.Errorf("selected=%s recorded=%q: got %d, want %d", selected, recorded, got, want)
			}
			entry.History = []AuditRound{{Round: 1, Verdict: entry.Verdict, Response: entry.Response}}
			if got := entry.RoundsUsed(); got != want {
				t.Errorf("migration changed count: selected=%s recorded=%q got=%d want=%d", selected, recorded, got, want)
			}
		}
	}
}

func TestGateResetAndFailureStreakAreIndependent(t *testing.T) {
	entry := AuditEntry{WorkItem: "W019", Gate: "implementation", ReadinessCheck: "delivery-code", SessionID: "native", Status: "incident", ExternalRound: 7,
		History: []AuditRound{
			{Round: 1, ReadinessCheck: "test-code", Verdict: "FAIL"},
			{Round: 2, ReadinessCheck: "delivery-code", Incident: "timeout"},
			{Round: 3, ReadinessCheck: "delivery-code", Verdict: "FAIL"},
			{Round: 4, ReadinessCheck: "delivery-code", Incident: "response-malformed"},
			{Round: 5, ReadinessCheck: "delivery-code", Incident: "resume-failed"},
			{Round: 6, ReadinessCheck: "test-code", Verdict: "PASS"},
			{Round: 7, ReadinessCheck: "delivery-code", Incident: "response-empty"},
		}}
	if got := entry.ConsecutiveFailures(); got != 3 {
		t.Fatalf("failure streak = %d, want 3", got)
	}
	path := filepath.Join(t.TempDir(), "audits.yaml")
	if err := WriteAuditEntry(path, entry); err != nil {
		t.Fatal(err)
	}
	if err := ResetAuditBudget(path, "W019", "implementation", "delivery-code"); err != nil {
		t.Fatal(err)
	}
	got, _, err := ReadAuditEntry(path, "W019", "implementation")
	if err != nil || got.RoundsUsed() != 0 || got.ConsecutiveFailures() != 0 {
		t.Fatalf("reset = %#v, %v", got, err)
	}
	got.ReadinessCheck = "test-code"
	if got.RoundsUsed() != 2 {
		t.Fatal("delivery reset changed test-code allowance")
	}
}
