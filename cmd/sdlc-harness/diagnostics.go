// ABOUTME: Keeps bounded private audit diagnostics separate from agent-facing results.
// ABOUTME: Retains rejected responses without exposing provider sessions or retry counts.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tigger-developer/sdlc/internal/harness"
)

type auditDiagnostics struct {
	mu        sync.Mutex
	file      *os.File
	output    io.Writer
	remaining int
}

func openAuditDiagnostics(root string, output io.Writer) (*auditDiagnostics, error) {
	directory := filepath.Join(root, ".sdlc", "audit-diagnostics")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, err
	}
	ignore, err := os.OpenFile(filepath.Join(directory, ".gitignore"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err == nil {
		_, writeErr := ignore.WriteString("*\n")
		closeErr := ignore.Close()
		if writeErr != nil {
			return nil, writeErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
	} else if !os.IsExist(err) {
		return nil, err
	}
	file, err := os.CreateTemp(directory, "audit-*.log")
	if err != nil {
		return nil, err
	}
	return &auditDiagnostics{file: file, output: output, remaining: 16 * 1024 * 1024}, nil
}

func (d *auditDiagnostics) Write(data []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(data)
	liveness := strings.HasPrefix(string(data), "sdlc-harness: external context started (") || strings.HasPrefix(string(data), "sdlc-harness: external context still running (")
	if len(data) > d.remaining {
		data = data[:d.remaining]
	}
	written, err := d.file.Write(data)
	d.remaining -= written
	if err != nil {
		return written, err
	}
	if liveness {
		if _, err := fmt.Fprintln(d.output, "Audit in progress."); err != nil {
			return written, err
		}
	}
	// This bounded sink accepts excess bytes into a discard sink, like io.Discard;
	// reaching the retention cap must not interrupt a running provider or heartbeat.
	return n, nil
}

func (d *auditDiagnostics) capture(body []byte) error {
	if _, err := fmt.Fprintln(d, "\nProvider output (unvalidated; not an audit verdict):"); err != nil {
		return err
	}
	_, err := d.Write(body)
	return err
}

func auditUnavailable() error {
	return &harness.Incident{Kind: "human-intervention", Harness: "audit", Err: fmt.Errorf("Audit unavailable: unable to obtain a valid result. Human intervention required; invoke --reset only after operator authorization")}
}
