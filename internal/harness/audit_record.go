package harness

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type AuditRecord struct {
	Version int           `yaml:"version"`
	Audits  []AuditEntry  `yaml:"audits"`
	Legacy  []LegacyAudit `yaml:"legacy,omitempty"`
}

// LegacyAudit preserves an audits.org file during migration to YAML.
type LegacyAudit struct {
	Path     string `yaml:"path"`
	Content  string `yaml:"content"`
	Migrated string `yaml:"migrated"`
}

type AuditEntry struct {
	WorkItem      string             `yaml:"work_item"`
	Gate          string             `yaml:"gate"`
	SessionID     string             `yaml:"session_id"`
	Harness       string             `yaml:"harness,omitempty"`
	Provider      string             `yaml:"provider,omitempty"`
	Model         string             `yaml:"model,omitempty"`
	Configured    *AgentConfig       `yaml:"configured,omitempty"`
	Sessions      []RetiredSession   `yaml:"previous_sessions,omitempty"`
	ExternalRound int                `yaml:"external_round"`
	BudgetResets  []AuditBudgetReset `yaml:"budget_resets,omitempty"`
	Status        string             `yaml:"status"`
	Updated       string             `yaml:"updated"`
	Revision      string             `yaml:"revision,omitempty"`
	Verdict       string             `yaml:"verdict,omitempty"`
	Response      string             `yaml:"response,omitempty"`
	Findings      []string           `yaml:"findings,omitempty"`
	Remediation   []string           `yaml:"remediation,omitempty"`
	History       []AuditRound       `yaml:"history,omitempty"`
}

// AuditBudgetReset starts a new bounded cycle without erasing lifetime history.
type AuditBudgetReset struct {
	AfterRound int    `yaml:"after_round"`
	Updated    string `yaml:"updated"`
}

// RoundsUsed counts provider launches since the last explicit operator reset.
func (entry AuditEntry) RoundsUsed() int {
	if len(entry.BudgetResets) == 0 {
		return entry.ExternalRound
	}
	return entry.ExternalRound - entry.BudgetResets[len(entry.BudgetResets)-1].AfterRound
}

// ResetAuditBudget preserves native context, findings, cache entries and history.
func ResetAuditBudget(path, workItem, gate string) error {
	entry, found, err := ReadAuditEntry(path, workItem, gate)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("no audit record for %s/%s", workItem, gate)
	}
	if entry.Status == "running" {
		return errors.New("cannot reset a running or interrupted audit; establish that it has stopped first")
	}
	if entry.RoundsUsed() == 0 {
		return nil
	}
	entry.BudgetResets = append(entry.BudgetResets, AuditBudgetReset{AfterRound: entry.ExternalRound, Updated: time.Now().UTC().Format(time.RFC3339)})
	if entry.Status == "exhausted" {
		entry.Status = "active"
		if entry.SessionID == "" {
			entry.Status = "identity-missing"
		}
	}
	return WriteAuditEntry(path, entry)
}

type AuditRound struct {
	CacheKey  string         `yaml:"cache_key,omitempty"`
	SessionID string         `yaml:"session_id,omitempty"`
	Harness   string         `yaml:"harness,omitempty"`
	Provider  string         `yaml:"provider,omitempty"`
	Model     string         `yaml:"model,omitempty"`
	Incident  string         `yaml:"incident,omitempty"`
	Evidence  []EvidenceFile `yaml:"evidence,omitempty"`
	Round     int            `yaml:"round"`
	Revision  string         `yaml:"revision,omitempty"`
	Verdict   string         `yaml:"verdict,omitempty"`
	Response  string         `yaml:"response,omitempty"`
	Findings  []string       `yaml:"findings,omitempty"`
	Updated   string         `yaml:"updated"`
}

// RetiredSession preserves context provenance without duplicating audit findings.
type RetiredSession struct {
	SessionID   string `yaml:"session_id"`
	AgentConfig `yaml:",inline"`
	Reason      string `yaml:"reason"`
	Updated     string `yaml:"updated"`
}

// LatestEvidence belongs to the latest invocation, including an interrupted audit.
// Old records without manifests deliberately trigger a full evidence read.
func (entry AuditEntry) LatestEvidence() []EvidenceFile {
	if len(entry.History) == 0 {
		return nil
	}
	return entry.History[len(entry.History)-1].Evidence
}

func ReadAuditEntry(path, workItem, gate string) (AuditEntry, bool, error) {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return AuditEntry{}, false, nil
	}
	if err != nil {
		return AuditEntry{}, false, fmt.Errorf("reading audit record %s: %w", path, err)
	}
	var record AuditRecord
	if err := yaml.Unmarshal(contents, &record); err != nil {
		return AuditEntry{}, false, fmt.Errorf("parsing audit record %s: %w", path, err)
	}
	for _, entry := range record.Audits {
		if entry.WorkItem == workItem && entry.Gate == gate {
			return entry, true, nil
		}
	}
	return AuditEntry{}, false, nil
}

// MigrateLegacyAudit preserves a sibling audits.org verbatim in audits.yaml.
// The legacy file is removed only after the YAML record has been written.
func MigrateLegacyAudit(path string) error {
	legacyPath := filepath.Join(filepath.Dir(path), "audits.org")
	content, err := os.ReadFile(legacyPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading legacy audit record %s: %w", legacyPath, err)
	}
	var record AuditRecord
	if existing, readErr := os.ReadFile(path); readErr == nil {
		if err := yaml.Unmarshal(existing, &record); err != nil {
			return fmt.Errorf("parsing audit record %s: %w", path, err)
		}
	} else if !errors.Is(readErr, os.ErrNotExist) {
		return fmt.Errorf("reading audit record %s: %w", path, readErr)
	}
	for _, legacy := range record.Legacy {
		if legacy.Path == "audits.org" {
			return nil
		}
	}
	if record.Version == 0 {
		record.Version = 1
	}
	record.Legacy = append(record.Legacy, LegacyAudit{
		Path: "audits.org", Content: string(content), Migrated: time.Now().UTC().Format(time.RFC3339),
	})
	if err := writeAuditRecord(path, record); err != nil {
		return err
	}
	if err := os.Remove(legacyPath); err != nil {
		return fmt.Errorf("removing migrated legacy audit record %s: %w", legacyPath, err)
	}
	return nil
}

func WriteAuditEntry(path string, entry AuditEntry) error {
	if strings.TrimSpace(entry.WorkItem) == "" || strings.TrimSpace(entry.Gate) == "" || (strings.TrimSpace(entry.SessionID) == "" && entry.Status != "running" && entry.Status != "identity-missing" && entry.Status != "exhausted") {
		return errors.New("audit record requires work item, gate, and session ID")
	}
	var record AuditRecord
	if contents, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(contents, &record); err != nil {
			return fmt.Errorf("parsing audit record %s: %w", path, err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading audit record %s: %w", path, err)
	}
	if record.Version == 0 {
		record.Version = 1
	}
	entry.Updated = time.Now().UTC().Format(time.RFC3339)
	if len(entry.History) > 0 && entry.History[len(entry.History)-1].Updated == "" {
		entry.History[len(entry.History)-1].Updated = entry.Updated
	}
	updated := false
	for index := range record.Audits {
		if record.Audits[index].WorkItem == entry.WorkItem && record.Audits[index].Gate == entry.Gate {
			record.Audits[index] = entry
			updated = true
			break
		}
	}
	if !updated {
		record.Audits = append(record.Audits, entry)
	}
	return writeAuditRecord(path, record)
}

func writeAuditRecord(path string, record AuditRecord) error {
	contents, err := yaml.Marshal(record)
	if err != nil {
		return fmt.Errorf("rendering audit record: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating audit record directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".audits.yaml-*")
	if err != nil {
		return fmt.Errorf("creating audit record temporary file: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(contents); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryName, path); err != nil {
		return fmt.Errorf("installing audit record %s: %w", path, err)
	}
	return nil
}
