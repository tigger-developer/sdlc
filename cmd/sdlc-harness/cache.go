// ABOUTME: Reuses completed audit responses only for an identical, versioned request digest.
// ABOUTME: Cache hits launch no provider and do not mutate history or consume rounds.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/tigger-developer/sdlc/internal/harness"
)

func auditCacheKey(work, gate, prompt string, registry auditPromptDocument, evidence harness.Evidence, config harness.Config) (string, error) {
	// Execution limits do not alter a verdict. Supplied configuration files, if
	// any, remain part of Evidence and are conservatively compared byte-for-byte.
	request := struct {
		Version  int
		Work     string
		Gate     string
		Prompt   string
		Registry auditPromptDocument
		Evidence []harness.EvidenceFile
		Primary  harness.AgentConfig
		Fallback *harness.AgentConfig
	}{1, work, gate, prompt, registry, evidence.Files, config.Agent(), config.Fallback}
	encoded, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func cachedAudit(entry harness.AuditEntry, key string) (harness.AuditRound, bool) {
	if key == "" || entry.Status == "running" {
		return harness.AuditRound{}, false
	}
	for index := len(entry.History) - 1; index >= 0; index-- {
		round := entry.History[index]
		if round.CacheKey != key || round.Incident != "" || round.SessionID == "" {
			continue
		}
		if harness.ValidateCompositeVerdict(round.Response) != nil || auditField(round.Response, "GATE") != entry.Gate || auditField(round.Response, "VERDICT") != round.Verdict || auditField(round.Response, "REVISION") != round.Revision {
			continue
		}
		return round, true
	}
	return harness.AuditRound{}, false
}
