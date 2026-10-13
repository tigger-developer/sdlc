// ABOUTME: Atomically claims one audit context before invoking a provider.
// ABOUTME: Uses the existing round and reset generation to reject obsolete owners.
package harness

import "errors"

var ErrAuditSuperseded = errors.New("another audit invocation owns this context; obsolete invocation cannot update it")
var ErrAuditRunning = errors.New("an audit already owns this work item and context; no second provider invoked")

// ClaimAuditEntry accepts exactly the next round under the audit-record lock.
// Provider execution never holds that lock; checkpoints retain this round owner.
func ClaimAuditEntry(path string, entry AuditEntry) error {
	return withAuditRecordLock(path, func(resolved string) error {
		current, found, err := ReadAuditEntry(resolved, entry.WorkItem, entry.Gate)
		if err != nil {
			return err
		}
		if found && current.Generation != entry.Generation {
			return ErrAuditReset
		}
		if found && current.Status == "running" {
			return ErrAuditRunning
		}
		if entry.ExternalRound <= 0 || entry.ExternalRound != current.ExternalRound+1 {
			return ErrAuditSuperseded
		}
		return writeAuditEntry(resolved, entry, true)
	})
}
