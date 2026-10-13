// ABOUTME: Keeps bounded private audit diagnostics separate from agent-facing results.
// ABOUTME: Retains rejected responses without exposing provider sessions or retry counts.
package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tigger-developer/sdlc/internal/harness"
)

//go:embed audit-start.md
var auditStartNotice string

type auditDiagnostics struct {
	mu        sync.Mutex
	file      *os.File
	output    io.Writer
	remaining int
}

func openAuditDiagnostics(root string, output io.Writer) (*auditDiagnostics, error) {
	return openDiagnosticsDirectory(root, filepath.Join(".sdlc", "audit-diagnostics"), output)
}

func openStandaloneDiagnostics(root string, output io.Writer) (*auditDiagnostics, error) {
	// Runtime diagnostics must not create .sdlc and elect an SDLC workflow.
	return openDiagnosticsDirectory(root, ".audit-diagnostics", output)
}

func openDiagnosticsDirectory(root, relative string, output io.Writer) (*auditDiagnostics, error) {
	directory, err := privateDiagnosticDirectory(root, relative)
	if err != nil {
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

// Validate every runtime component before creating a child. MkdirAll would
// follow an existing parent or leaf symlink before its destination was checked.
func privateDiagnosticDirectory(root, relative string) (string, error) {
	directory, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(directory)
	if err != nil {
		return "", err
	}
	if info.Mode().Perm()&0022 != 0 {
		return "", fmt.Errorf("audit diagnostic parent %q must not be writable by other users; permissions were not changed", directory)
	}
	for _, part := range strings.Split(relative, string(filepath.Separator)) {
		directory = filepath.Join(directory, part)
		info, err = os.Lstat(directory)
		if os.IsNotExist(err) {
			// Another local opener may create it first; validate either result.
			if createErr := os.Mkdir(directory, 0700); createErr != nil && !os.IsExist(createErr) {
				return "", createErr
			}
			info, err = os.Lstat(directory)
		}
		if err != nil {
			return "", err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("audit diagnostic path %q must be a real directory, not a symlink", directory)
		}
		if info.Mode().Perm()&0022 != 0 {
			return "", fmt.Errorf("audit diagnostic parent %q must not be writable by other users; permissions were not changed", directory)
		}
	}
	if info.Mode().Perm() != 0700 {
		return "", fmt.Errorf("audit diagnostic directory %q must be private (0700); permissions were not changed", directory)
	}
	resolved, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return "", err
	}
	if resolved != directory {
		return "", fmt.Errorf("audit diagnostic directory %q changed while resolving", directory)
	}
	return directory, nil
}

func (d *auditDiagnostics) Write(data []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(data)
	started := strings.HasPrefix(string(data), "sdlc-harness: external context started (")
	liveness := started || strings.HasPrefix(string(data), "sdlc-harness: external context still running (")
	if len(data) > d.remaining {
		data = data[:d.remaining]
	}
	written, err := d.file.Write(data)
	d.remaining -= written
	if err != nil {
		return written, err
	}
	if started {
		if _, err := io.WriteString(d.output, auditStartNotice); err != nil {
			return written, err
		}
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

// report is a separate trusted path: native stderr never selects public text.
func (d *auditDiagnostics) report(message string) error {
	if _, err := d.Write([]byte(message + "\n")); err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := fmt.Fprintln(d.output, message)
	return err
}

func auditUnavailable() error {
	return &harness.Incident{Kind: "human-intervention", Harness: "audit", Err: fmt.Errorf("Audit unavailable: unable to obtain a valid result. Human intervention required; invoke --reset only after operator authorization")}
}
