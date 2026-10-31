// ABOUTME: Migrates an initialized v3 project's specs directory to tickets.
// ABOUTME: Preserves configuration and numbering while repairing tracked text references.
package projectinit

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type ticketTextEdit struct {
	path     string
	contents []byte
	mode     os.FileMode
}

func migrateV3Tickets(options Options, root string) error {
	profile, err := readYAMLConfig(filepath.Join(root, projectProfilePath))
	if err != nil {
		return err
	}
	if profile["version"] != 3 {
		return fmt.Errorf("%s must declare version 3 for ticket-directory migration", projectProfilePath)
	}
	source := filepath.Join(root, "specs")
	if !exists(source) {
		fmt.Fprintln(options.Output, "SDLC v3 is initialized; no specs directory requires migration.")
		return nil
	}
	info, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("specs must be a regular directory")
	}
	if exists(filepath.Join(root, "tickets")) {
		return fmt.Errorf("both specs and tickets exist; resolve the destination conflict before sdlc-init")
	}
	if err := requireCleanWorktree(options, root); err != nil {
		return err
	}
	edits, err := planTicketTextEdits(options, root)
	if err != nil {
		return err
	}
	if err := options.RunCommand("git", []string{"mv", "--", "specs", "tickets"}, root, nil, options.Output, options.ErrorOutput); err != nil {
		return fmt.Errorf("renaming specs to tickets: %w", err)
	}
	for _, edit := range edits {
		if err := os.WriteFile(filepath.Join(root, edit.path), edit.contents, edit.mode); err != nil {
			return fmt.Errorf("updating %s after directory rename; inspect the uncommitted migration before retrying: %w", edit.path, err)
		}
	}
	if err := options.RunCommand("git", []string{"add", "-u"}, root, nil, options.Output, options.ErrorOutput); err != nil {
		return fmt.Errorf("staging ticket migration: %w", err)
	}
	if err := options.RunCommand("git", []string{"commit", "-m", "chore: migrate specs directory to tickets"}, root, nil, options.Output, options.ErrorOutput); err != nil {
		return fmt.Errorf("committing ticket migration; changes remain for inspection: %w", err)
	}
	fmt.Fprintf(options.Output, "Migrated specs to tickets and updated %d tracked text files; ticket identities and project settings were retained.\n", len(edits))
	return nil
}

func planTicketTextEdits(options Options, root string) ([]ticketTextEdit, error) {
	output, err := commandOutput(options, root, "git", "ls-files", "-z")
	if err != nil {
		return nil, err
	}
	var edits []ticketTextEdit
	for _, path := range strings.Split(output, "\x00") {
		if path == "" {
			continue
		}
		switch filepath.Ext(path) {
		case ".org", ".md", ".yaml":
		default:
			continue
		}
		full := filepath.Join(root, filepath.FromSlash(path))
		info, err := os.Lstat(full)
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(full)
			if err != nil {
				return nil, err
			}
			if strings.Contains(target, "specs/") {
				return nil, fmt.Errorf("symlink %s refers to specs/; reconcile it before migration", path)
			}
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}
		contents, err := os.ReadFile(full)
		if err != nil {
			return nil, err
		}
		if !utf8.Valid(contents) || bytes.ContainsRune(contents, 0) {
			return nil, fmt.Errorf("selected document %s is not UTF-8 text; convert it before migration", path)
		}
		updated := bytes.ReplaceAll(contents, []byte("specs/"), []byte("tickets/"))
		if bytes.Equal(contents, updated) {
			continue
		}
		if strings.HasPrefix(path, "specs/") {
			path = "tickets/" + strings.TrimPrefix(path, "specs/")
		}
		edits = append(edits, ticketTextEdit{path: path, contents: updated, mode: info.Mode().Perm()})
	}
	return edits, nil
}
