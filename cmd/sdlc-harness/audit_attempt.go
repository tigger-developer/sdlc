// ABOUTME: Checkpoints audit invocations and bounded timeout recovery in the harness-owned record.
// ABOUTME: Preserves native session identity and evidence without inventing an audit verdict.
package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/tigger-developer/sdlc/internal/harness"
)

type auditAttempt struct {
	path      string
	entry     harness.AuditEntry
	limit     int
	finalized bool
}

func beginAuditAttempt(path string, entry harness.AuditEntry, work, gate string, config harness.Config, evidence harness.Evidence, cacheKey string) (*auditAttempt, error) {
	if len(entry.History) == 0 && (entry.Response != "" || entry.Verdict != "" || entry.Revision != "") {
		entry.History = append(entry.History, harness.AuditRound{Round: entry.ExternalRound, Revision: entry.Revision, Verdict: entry.Verdict, Response: entry.Response, Findings: entry.Findings, Updated: entry.Updated, SessionID: entry.SessionID, Harness: entry.Harness, Provider: entry.Provider, Model: entry.Model})
	}
	entry.WorkItem, entry.Gate, entry.Harness = work, gate, config.Harness
	entry.Provider, entry.Model = config.Provider, config.Model
	entry.ExternalRound++
	entry.Status = "running"
	entry.Revision, entry.Verdict, entry.Response = "", "", ""
	// Existing structured findings/remediation remain available; a timeout does
	// not resolve them. Prior response bodies remain in their historical rounds.
	entry.History = append(entry.History, harness.AuditRound{Round: entry.ExternalRound, Evidence: evidence.Files, Incident: "interrupted", Harness: config.Harness, Provider: config.Provider, Model: config.Model})
	entry.History[len(entry.History)-1].CacheKey = cacheKey
	attempt := &auditAttempt{path: path, entry: entry, limit: config.MaxRounds}
	return attempt, harness.WriteAuditEntry(path, entry)
}

func (attempt *auditAttempt) recordIdentity(identity string) error {
	if attempt.entry.SessionID != "" && identity != attempt.entry.SessionID {
		return errors.New("native session identity differs from the recorded audit session")
	}
	attempt.entry.SessionID = identity
	attempt.entry.History[len(attempt.entry.History)-1].SessionID = identity
	return harness.WriteAuditEntry(attempt.path, attempt.entry)
}

func (attempt *auditAttempt) finish(runErr error) error {
	// A validated response remains recorded even if writing stdout then fails.
	if runErr == nil || attempt.finalized {
		return nil
	}
	kind := "runner-error"
	var incident *harness.Incident
	if errors.As(runErr, &incident) {
		kind = incident.Kind
	}
	attempt.entry.Status = "incident"
	if kind == "timeout" {
		attempt.entry.Status = "timed-out"
	}
	if attempt.entry.SessionID == "" {
		attempt.entry.Status = "identity-missing"
	}
	if attempt.entry.ExternalRound >= attempt.limit {
		attempt.entry.Status = "exhausted"
	}
	round := &attempt.entry.History[len(attempt.entry.History)-1]
	round.Incident = kind
	return harness.WriteAuditEntry(attempt.path, attempt.entry)
}

func lastIncident(entry harness.AuditEntry) string {
	if len(entry.History) == 0 {
		return ""
	}
	return entry.History[len(entry.History)-1].Incident
}

func (attempt *auditAttempt) reportTimeout(registry auditPromptDocument, output io.Writer) {
	remaining := attempt.limit - attempt.entry.ExternalRound
	fmt.Fprintf(output, "AUDIT TIMEOUT: no verdict; attempt %d/%d; remaining=%d; record=%s\n", attempt.entry.ExternalRound, attempt.limit, remaining, attempt.path)
	if attempt.entry.SessionID == "" || remaining <= 0 {
		fmt.Fprintln(output, strings.TrimSpace(registry.TimeoutBlockedMessage))
		fmt.Fprintf(output, "Recovery unavailable: native_session_recorded=%t; remaining=%d\n", attempt.entry.SessionID != "", remaining)
		return
	}
	fmt.Fprintf(output, "SESSION_ID: %s\n%s\n", attempt.entry.SessionID, strings.TrimSpace(registry.TimeoutMessage))
}
