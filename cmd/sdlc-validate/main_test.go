package main

import (
	"bytes"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestHelpAndInvalidInvocation(t *testing.T) {
	for _, option := range []string{"-h", "--help", "--version"} {
		var out, diagnostics bytes.Buffer
		if code := run([]string{option}, &out, &diagnostics); code != 0 || out.Len() == 0 {
			t.Fatalf("%s: %d %s", option, code, out.String())
		}
	}
	for _, args := range [][]string{nil, {"--readiness-check=test-spec"}, {"--unknown"}, {"--readiness-check=test-code", "--spec=/absent/spec.org"}} {
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code != 3 {
			t.Fatalf("%v: %d", args, code)
		}
		var doc map[string]any
		if err := yaml.Unmarshal(out.Bytes(), &doc); err != nil || doc["audit-readiness"] == nil {
			t.Fatalf("not a YAML diagnostic: %s (%v)", out.String(), err)
		}
	}
}
