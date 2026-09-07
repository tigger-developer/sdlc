package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpDescribesSupportedInput(t *testing.T) {
	var output bytes.Buffer
	if status := run([]string{"--help"}, &output, &output); status != 0 {
		t.Fatalf("status = %d", status)
	}
	if !strings.Contains(output.String(), "Markdown or Org") {
		t.Fatalf("help = %q", output.String())
	}
	if !strings.Contains(output.String(), "does not modify the source document") {
		t.Fatalf("help does not explain its source boundary: %q", output.String())
	}
}

func TestCleanupRefusesArbitraryFiles(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "notes.html")
	if err := os.WriteFile(path, []byte(previewOwnershipMarker), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := removeGeneratedPreview(path); err == nil {
		t.Fatal("cleanup accepted an arbitrary filename")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("arbitrary file was changed: %v", err)
	}
}

func TestCleanupRemovesOnlyOwnedGeneratedPreview(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), ".sdlc-preview-owned.html")
	if err := os.WriteFile(path, []byte(previewOwnershipMarker), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := removeGeneratedPreview(path); err != nil {
		t.Fatalf("removeGeneratedPreview: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("owned preview remains or stat failed unexpectedly: %v", err)
	}
}

func TestInputFormatRejectsUnsupportedDocuments(t *testing.T) {
	for path, want := range map[string]string{"spec.org": "org", "README.md": "gfm", "README.markdown": "gfm"} {
		got, err := inputFormat(path)
		if err != nil || got != want {
			t.Fatalf("inputFormat(%q) = %q, %v", path, got, err)
		}
	}
	if _, err := inputFormat("secret.env"); err == nil {
		t.Fatal("unsupported input was accepted")
	}
}
