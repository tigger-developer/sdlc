package main

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tigger-developer/sdlc/internal/auditrunner"
)

func TestParseOptionsCollectsBoundedInputsAndOverrides(t *testing.T) {
	options, err := parseOptions([]string{
		"audit-design",
		"--artifact", "specs/001/plan.md",
		"--artifact", "specs/001/data-model.md",
		"--context", "docs/architecture.md",
		"--external-context", "/opt/contracts/integration.md",
		"--provider", "openai-codex",
		"--model", "gpt-5.6-luna",
		"--timeout", "4m",
	})
	if err != nil {
		t.Fatalf("parseOptions() error = %v", err)
	}
	if options.AuditName != "audit-design" || options.Provider != "openai-codex" || options.Model != "gpt-5.6-luna" || options.Timeout != 4*time.Minute || !options.TimeoutSet {
		t.Fatalf("unexpected options: %#v", options)
	}
	if !slices.Equal(options.Artifacts, []string{"specs/001/plan.md", "specs/001/data-model.md"}) {
		t.Fatalf("artifacts = %#v", options.Artifacts)
	}
	if !slices.Equal(options.Context, []string{"docs/architecture.md"}) || !slices.Equal(options.ExternalContext, []string{"/opt/contracts/integration.md"}) {
		t.Fatalf("bounded context = %#v; external context = %#v", options.Context, options.ExternalContext)
	}
}

func TestParseOptionsRejectsUnsupportedTimeouts(t *testing.T) {
	for _, value := range []string{"0", "-1s", "500ms", "1500ms"} {
		t.Run(value, func(t *testing.T) {
			if _, err := parseOptions([]string{"audit-code", "--artifact", "change.diff", "--timeout", value}); err == nil || !strings.Contains(err.Error(), "whole-second duration") {
				t.Fatalf("timeout %q error = %v", value, err)
			}
		})
	}
}

func TestHelpDocumentsBoundedInputFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if got := execute([]string{"--help"}, &stdout, &stderr, nil); got != 0 {
		t.Fatalf("execute() = %d; stderr=%q", got, stderr.String())
	}
	for _, flag := range []string{"--artifact", "--context", "--external-context", "--timeout"} {
		if !strings.Contains(stdout.String(), flag) {
			t.Errorf("help does not document %s:\n%s", flag, stdout.String())
		}
	}
}

func TestExecuteProvidesHelpVersionAndExitClasses(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
		run  func(auditrunner.Options) error
		want int
	}{
		{name: "help", args: []string{"--help"}, want: 0},
		{name: "version", args: []string{"--version"}, want: 0},
		{name: "invalid", args: []string{"--artifact", "spec.md"}, want: 2},
		{name: "operational", args: []string{"audit-spec", "--artifact", "spec.md"}, run: func(auditrunner.Options) error { return errors.New("failed") }, want: 1},
		{name: "success", args: []string{"audit-spec", "--artifact", "spec.md"}, run: func(auditrunner.Options) error { return nil }, want: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if got := execute(test.args, &stdout, &stderr, test.run); got != test.want {
				t.Fatalf("execute() = %d, want %d; stdout=%q stderr=%q", got, test.want, stdout.String(), stderr.String())
			}
		})
	}
}
