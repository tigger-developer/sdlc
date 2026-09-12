// ABOUTME: Captures bounded audit evidence as paths and hashes, without copying files.
// ABOUTME: Compares retained manifests and rejects evidence changes during review.
package harness

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"unicode/utf8"
)

const maximumInputBytes = 16 * 1024 * 1024
const maximumEvidenceBytes = 64 * 1024 * 1024

// EvidenceFile identifies the bytes at an original, absolute file path.
type EvidenceFile struct {
	Path   string `yaml:"path" json:"path"`
	SHA256 string `yaml:"sha256" json:"sha256"`
	Bytes  int64  `yaml:"bytes" json:"bytes"`
}

// Evidence contains metadata only; source files are never copied or modified.
type Evidence struct {
	Files []EvidenceFile
}

// CaptureEvidence validates and hashes a complete input list in stable path order.
func CaptureEvidence(project string, paths []string) (Evidence, error) {
	if len(paths) == 0 {
		return Evidence{}, errors.New("at least one --input evidence file is required")
	}
	evidence := Evidence{}
	seen := map[string]bool{}
	var total int64
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			path = filepath.Join(project, path)
		}
		path, err := filepath.Abs(path)
		if err != nil {
			return Evidence{}, err
		}
		if seen[path] {
			return Evidence{}, fmt.Errorf("duplicate evidence path %q", path)
		}
		seen[path] = true
		file, err := hashEvidenceFile(path)
		if err != nil {
			return Evidence{}, err
		}
		total += file.Bytes
		if total > maximumEvidenceBytes {
			return Evidence{}, errors.New("evidence exceeds the 64 MiB total size limit")
		}
		evidence.Files = append(evidence.Files, file)
	}
	sort.Slice(evidence.Files, func(i, j int) bool { return evidence.Files[i].Path < evidence.Files[j].Path })
	return evidence, nil
}

func hashEvidenceFile(path string) (result EvidenceFile, returnErr error) {
	return readEvidenceFile(path, nil)
}

func readEvidenceFile(path string, contents io.Writer) (result EvidenceFile, returnErr error) {
	if filepath.Base(path) == ".env" {
		return result, fmt.Errorf("evidence input %q is a prohibited .env file", path)
	}
	info, err := os.Lstat(path)
	if err != nil {
		return result, fmt.Errorf("inspecting evidence %q: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return result, fmt.Errorf("evidence %q must be a regular non-symlink file", path)
	}
	if info.Size() > maximumInputBytes {
		return result, fmt.Errorf("evidence %q exceeds the 16 MiB size limit", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return result, fmt.Errorf("opening evidence %q: %w", path, err)
	}
	defer func() {
		if err := file.Close(); err != nil && returnErr == nil {
			returnErr = err
		}
	}()
	opened, err := file.Stat()
	if err != nil {
		return result, err
	}
	if !os.SameFile(info, opened) {
		return result, fmt.Errorf("evidence %q changed while opening", path)
	}
	hash := sha256.New()
	var destination io.Writer = hash
	if contents != nil {
		destination = io.MultiWriter(hash, contents)
	}
	n, err := io.Copy(destination, io.LimitReader(file, maximumInputBytes+1))
	if err != nil {
		return result, fmt.Errorf("hashing evidence %q: %w", path, err)
	}
	if n > maximumInputBytes {
		return result, fmt.Errorf("evidence %q exceeds the 16 MiB size limit", path)
	}
	return EvidenceFile{Path: path, SHA256: hex.EncodeToString(hash.Sum(nil)), Bytes: n}, nil
}

// ContentPrompt supplies verified text to a tools-disabled adapter over stdin.
// Nothing is written to disk or persisted in the audit manifest.
func (evidence Evidence) ContentPrompt() (string, error) {
	type document struct {
		EvidenceFile
		Content string `json:"content"`
	}
	documents := make([]document, 0, len(evidence.Files))
	for _, want := range evidence.Files {
		var contents bytes.Buffer
		got, err := readEvidenceFile(want.Path, &contents)
		if err != nil {
			return "", err
		}
		if got != want {
			return "", fmt.Errorf("audit evidence %q changed before transport", want.Path)
		}
		if !utf8.Valid(contents.Bytes()) {
			return "", fmt.Errorf("inline audit evidence %q must be UTF-8 text", want.Path)
		}
		documents = append(documents, document{EvidenceFile: want, Content: contents.String()})
	}
	encoded, err := json.Marshal(documents)
	if err != nil {
		return "", err
	}
	return "Evidence contents (tools-disabled adapter):\n" + string(encoded) + "\n", nil
}

// Verify checks originals immediately before and after provider execution.
func (evidence Evidence) Verify() error {
	for _, want := range evidence.Files {
		got, err := hashEvidenceFile(want.Path)
		if err != nil {
			return err
		}
		if got != want {
			return fmt.Errorf("audit evidence %q changed; result cannot be accepted", want.Path)
		}
	}
	return nil
}

// EvidenceChange compares this input list with the previous accepted round.
// Removed means no longer supplied, not necessarily deleted from the repository.
type EvidenceChange struct {
	EvidenceFile
	Change string `json:"change"`
}

func (evidence Evidence) Changes(previous []EvidenceFile) []EvidenceChange {
	prior := map[string]EvidenceFile{}
	for _, file := range previous {
		prior[file.Path] = file
	}
	changes := make([]EvidenceChange, 0, len(evidence.Files))
	for _, file := range evidence.Files {
		change := "added"
		if old, found := prior[file.Path]; found {
			change = "unchanged"
			if old.SHA256 != file.SHA256 {
				change = "changed"
			}
		}
		changes = append(changes, EvidenceChange{EvidenceFile: file, Change: change})
		delete(prior, file.Path)
	}
	for _, file := range prior {
		changes = append(changes, EvidenceChange{EvidenceFile: file, Change: "removed"})
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	return changes
}

// Prompt supplies instructions and metadata, never document contents.
func (evidence Evidence) Prompt(previous []EvidenceFile) (string, error) {
	manifest, err := json.MarshalIndent(evidence.Changes(previous), "", "  ")
	if err != nil {
		return "", fmt.Errorf("rendering evidence manifest: %w", err)
	}
	return "Evidence manifest:\n```json\n" + string(manifest) + "\n```\n", nil
}
