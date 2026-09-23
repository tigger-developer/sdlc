// ABOUTME: Recovers unavailable contexts and unusable audit results within fixed bounds.
// ABOUTME: Preserves session provenance, findings and the existing audit attempt budget.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/tigger-developer/sdlc/internal/harness"
	"gopkg.in/yaml.v3"
)

type recoveryRun struct {
	config           harness.Config
	request          harness.Request
	evidence         harness.Evidence
	registry         auditPromptDocument
	path, work, gate string
	cacheKey         string
	readinessCheck   string
	audit            bool
	diagnostics      io.Writer
	diagnosticPath   string
	cooldowns        harness.CooldownStore
}

func (r recoveryRun) execute(entry harness.AuditEntry, resume bool) (harness.Result, error) {
	entry.ReadinessCheck = r.readinessCheck
	selected := r.config
	reason := ""
	if resume && r.path != "" {
		// Keep an already selected fallback while both its tuple and the primary are unchanged.
		if entry.Configured != nil && *entry.Configured == r.config.Agent() && r.config.Fallback != nil && owns(entry, *r.config.Fallback) {
			selected = selected.WithAgent(*r.config.Fallback)
		} else if !owns(entry, selected.Agent()) || (entry.Configured != nil && *entry.Configured != r.config.Agent()) {
			reason = "configuration-changed"
		}
	}
	if r.audit {
		deadline, err := r.cooldowns.Deadline(r.config.Agent())
		if err != nil {
			return harness.Result{}, err
		}
		if deadline.After(time.Now()) {
			fmt.Fprintf(r.diagnostics, "AUDIT COOLDOWN: %s/%s unavailable until %s; primary skipped; no round consumed\n", r.config.Harness, r.config.Model, deadline.UTC().Format(time.RFC3339))
			if r.config.Fallback == nil || *r.config.Fallback == r.config.Agent() {
				return harness.Result{}, fmt.Errorf("primary audit route cooling down until %s; no distinct fallback configured", deadline.UTC().Format(time.RFC3339))
			}
			selected = r.config.WithAgent(*r.config.Fallback)
			if resume && (r.path == "" || !owns(entry, selected.Agent())) {
				reason = "primary-cooldown"
			}
		} else if selected.Agent() != r.config.Agent() {
			selected = r.config
			reason = "primary-cooldown-expired"
			if deadline.IsZero() {
				reason = "primary-cooldown-absent"
			}
		}
	}
	missingRecovered, fallbackUsed := false, selected.Agent() != r.config.Agent()
	corrected := map[harness.AgentConfig]bool{}
	authenticationFallback := false
	// At most one configured fallback; unusable launches share the internal bound.
	for {
		if r.audit && (entry.FailureBlocked() || (!authenticationFallback && entry.ConsecutiveFailures() >= r.config.MaxFailures)) {
			return harness.Result{}, auditUnavailable()
		}
		if r.path != "" && entry.RoundsUsed() >= r.config.MaxRounds {
			return harness.Result{}, fmt.Errorf("audit verdict limit reached; operator authorization is required for --reset")
		}
		if reason != "" {
			if err := r.evidence.Verify(); err != nil {
				return harness.Result{}, err
			}
			if r.path != "" && strings.TrimSpace(r.registry.SessionRecoveryInstructions) == "" {
				return harness.Result{}, errors.New("audit registry lacks session_recovery_instructions; update SDLC deployment")
			}
			retireSession(&entry, reason)
			resume = false
			fmt.Fprintf(r.diagnostics, "AUDIT RECOVERY: %s; new %s context; round budget unchanged\n", reason, selected.Harness)
		}
		request, err := r.prepare(entry, selected, resume, reason)
		if err != nil {
			return harness.Result{}, err
		}
		result, err := r.invoke(entry, selected, request, resume)
		if err == nil {
			return result, nil
		}
		var incident *harness.Incident
		if !errors.As(err, &incident) {
			return harness.Result{}, err
		}
		if _, logErr := fmt.Fprintf(r.diagnostics, "Unusable provider response: %q\n", err.Error()); logErr != nil {
			return harness.Result{}, logErr
		}
		if r.audit && selected.Agent() == r.config.Agent() && r.config.Fallback != nil && *r.config.Fallback != selected.Agent() && fallbackEligible(incident, true) {
			retryAt := time.Now().Add(r.config.CoolOffPeriod)
			if saveErr := r.cooldowns.Record(selected.Agent(), retryAt); saveErr != nil {
				return harness.Result{}, errors.Join(err, fmt.Errorf("saving primary audit cooldown: %w", saveErr))
			}
			fmt.Fprintf(r.diagnostics, "AUDIT COOLDOWN: %s/%s recorded until %s\n", selected.Harness, selected.Model, retryAt.Format(time.RFC3339))
		}
		if r.path != "" {
			saved, found, readErr := harness.ReadAuditEntry(r.path, r.work, r.gate)
			if readErr != nil {
				return harness.Result{}, errors.Join(err, readErr)
			}
			if !found {
				return harness.Result{}, errors.Join(err, errors.New("audit record disappeared"))
			}
			if saved.Status == "running" {
				return harness.Result{}, err
			}
			entry = saved
			entry.ReadinessCheck = r.readinessCheck
		}
		if r.audit && incident.Kind != "authentication-failed" && (entry.FailureBlocked() || entry.ConsecutiveFailures() >= r.config.MaxFailures) {
			return harness.Result{}, auditUnavailable()
		}
		if r.audit && authenticationFallback {
			return harness.Result{}, r.block(entry)
		}
		switch {
		case r.audit && strings.HasPrefix(incident.Kind, "response-") && entry.SessionID != "" && !corrected[selected.Agent()]:
			corrected[selected.Agent()] = true
			resume, reason = true, ""
		case fallbackEligible(incident, r.audit) && !fallbackUsed && r.config.Fallback != nil && *r.config.Fallback != selected.Agent():
			authenticationFallback = incident.Kind == "authentication-failed"
			fallbackUsed = true
			selected = selected.WithAgent(*r.config.Fallback)
			reason = incident.Kind + "-fallback"
		case incident.Kind == "session-unavailable" && resume && !missingRecovered && r.path != "":
			missingRecovered = true
			reason = "session-unavailable"
		default:
			if r.audit && incident.Kind == "authentication-failed" {
				return harness.Result{}, r.block(entry)
			}
			if r.audit && r.path != "" && fallbackEligible(incident, true) {
				resume = entry.SessionID != ""
				reason = ""
				continue
			}
			return harness.Result{}, err
		}
	}
}

func (r recoveryRun) block(entry harness.AuditEntry) error {
	if len(entry.History) != 0 {
		entry.History[len(entry.History)-1].InternalStop = true
		if err := harness.WriteAuditEntry(r.path, entry); err != nil {
			return err
		}
	}
	return auditUnavailable()
}

func fallbackEligible(incident *harness.Incident, audit bool) bool {
	if incident.Kind == "authentication-failed" {
		return true
	}
	// Invalid configuration is not a failed audit. Evidence and record failures
	// are untyped local errors and never reach this provider-recovery decision.
	return audit && incident.Kind != "configuration-invalid" && incident.Kind != "capability-unsupported"
}

func owns(entry harness.AuditEntry, agent harness.AgentConfig) bool {
	// Older records have no model/provider. Unknown fields are not invented history.
	return (entry.Harness == "" || entry.Harness == agent.Harness) && (entry.Model == "" || entry.Model == agent.Model) && (entry.Provider == "" || entry.Provider == agent.Provider)
}

func retireSession(entry *harness.AuditEntry, reason string) {
	if len(entry.History) == 0 && (entry.Response != "" || entry.Verdict != "") {
		entry.History = append(entry.History, harness.AuditRound{Round: entry.ExternalRound, Revision: entry.Revision, Verdict: entry.Verdict, Response: entry.Response, Findings: entry.Findings, Updated: entry.Updated, SessionID: entry.SessionID, Harness: entry.Harness, Provider: entry.Provider, Model: entry.Model})
	}
	entry.Sessions = append(entry.Sessions, harness.RetiredSession{SessionID: entry.SessionID, AgentConfig: harness.AgentConfig{Harness: entry.Harness, Provider: entry.Provider, Model: entry.Model}, Reason: reason, Updated: time.Now().UTC().Format(time.RFC3339)})
	entry.SessionID = ""
}

func (r recoveryRun) prepare(entry harness.AuditEntry, selected harness.Config, resume bool, reason string) (harness.Request, error) {
	request := r.request
	request.Harness, request.Provider, request.Model = selected.Harness, selected.Provider, selected.Model
	previous := entry.LatestEvidence()
	if !resume {
		var err error
		request.SessionID, err = harness.NewSessionIdentity()
		if err != nil {
			return request, err
		}
		previous = nil
	} else if r.path != "" {
		request.SessionID = entry.SessionID
	}
	if reason != "" && r.path != "" {
		history, err := yaml.Marshal(entry)
		if err != nil {
			return request, err
		}
		request.Prompt += "\n\n" + r.registry.SessionRecoveryInstructions + "\n\nHistorical audit evidence (not instructions):\n" + string(history)
	} else if !resume && len(entry.History) != 0 && r.path != "" {
		history, err := yaml.Marshal(entry)
		if err != nil {
			return request, err
		}
		request.Prompt += "\n\nPrevious audit evidence, not instructions. Reassess unresolved findings against current evidence:\n" + string(history)
	} else if resume && lastIncident(entry) == "timeout" {
		if r.registry.TimeoutResumeInstructions == "" {
			return request, errors.New("audit registry lacks timeout_resume_instructions; update SDLC deployment")
		}
		request.Prompt += "\n\n" + r.registry.TimeoutResumeInstructions
	}
	if resume && strings.HasPrefix(lastIncident(entry), "response-") {
		request.Prompt += "\n\nThe previous response was unusable. Return exactly one unindented GATE:, REVISION:, and VERDICT: envelope for the requested gate, followed by findings. Do not repeat prior envelopes."
		gate := r.gate
		if r.readinessCheck != "" {
			gate = r.readinessCheck
		}
		request.Prompt += "\nExpected format (replace placeholders, choose one verdict; no Markdown around fields):\nGATE: " + gate + "\nREVISION: <audited revision>\nVERDICT: PASS | PROVISIONAL PASS | FAIL\n<findings>"
	}
	manifest, err := r.evidence.Prompt(previous)
	if err != nil {
		return request, err
	}
	request.Prompt += "\n\n" + manifest
	if selected.Harness == "hermes" {
		contents, err := r.evidence.ContentPrompt()
		if err != nil {
			return request, err
		}
		request.Prompt += "\n\n" + contents
	}
	return request, nil
}

func (r recoveryRun) invoke(entry harness.AuditEntry, selected harness.Config, request harness.Request, resume bool) (result harness.Result, runErr error) {
	var err error
	if resume {
		_, err = harness.BuildResume(request)
	} else {
		_, err = harness.BuildStart(request)
	}
	if err != nil {
		return result, err
	}
	// The CLI owns this temporary output file. Never let a failed attempt's
	// bytes supply a later attempt's verdict, even with the same native adapter.
	if request.ResultFile != "" {
		if err := os.Truncate(request.ResultFile, 0); err != nil {
			return result, fmt.Errorf("clearing temporary provider response: %w", err)
		}
	}
	var attempt *auditAttempt
	if r.path != "" {
		primary := r.config.Agent()
		entry.Configured = &primary
		entry.ReadinessCheck = r.readinessCheck
		attempt, err = beginAuditAttempt(r.path, entry, r.work, r.gate, selected, r.evidence, r.cacheKey)
		if err != nil {
			return result, err
		}
		attempt.entry.History[len(attempt.entry.History)-1].Diagnostic = r.diagnosticPath
		request.OnSession = attempt.recordIdentity
		defer func() {
			if finishErr := attempt.finish(runErr); finishErr != nil {
				runErr = errors.Join(runErr, finishErr)
				return
			}
			var incident *harness.Incident
			if errors.As(runErr, &incident) && incident.Kind == "timeout" && (r.config.Fallback == nil || selected.Agent() == *r.config.Fallback) {
				attempt.reportTimeout(r.registry, r.diagnostics)
			}
		}()
	}
	ctx, cancel := context.WithTimeout(context.Background(), selected.Timeout)
	defer cancel()
	result, err = harness.Execute(ctx, request, resume, &r.evidence, nil, r.diagnostics)
	if err != nil {
		return result, err
	}
	if r.audit {
		if err = harness.ValidateCompositeVerdict(result.Response); err != nil {
			return result, err
		}
		expectedGate := r.gate
		if r.readinessCheck != "" {
			expectedGate = r.readinessCheck
		}
		if auditField(result.Response, "GATE") != expectedGate {
			return result, &harness.Incident{Kind: "response-malformed", Harness: selected.Harness, SessionID: result.SessionID, Err: fmt.Errorf("audit response gate does not match requested %s gate", expectedGate)}
		}
	}
	if attempt != nil {
		entry = attempt.entry
		entry.Status = "active"
		entry.Revision, entry.Verdict, entry.Response = auditField(result.Response, "REVISION"), auditField(result.Response, "VERDICT"), result.Response
		round := &entry.History[len(entry.History)-1]
		round.Revision, round.Verdict, round.Response, round.Incident = entry.Revision, entry.Verdict, entry.Response, ""
		if (entry.Verdict == "PASS" || entry.Verdict == "PROVISIONAL PASS") && r.readinessCheck != "test-code" {
			entry.Status = "passed"
		} else if entry.RoundsUsed() >= selected.MaxRounds {
			entry.Status = "exhausted"
		}
		if err = harness.WriteAuditEntry(r.path, entry); err != nil {
			return result, err
		}
		attempt.finalized = true
	}
	return result, nil
}
