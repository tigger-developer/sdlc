package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
)

func TestSourceRevisionFromBuildInfo(t *testing.T) {
	clean := &debug.BuildInfo{
		Main: debug.Module{Version: "(devel)"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "0123456789abcdef"},
			{Key: "vcs.modified", Value: "false"},
		},
	}
	if got := sourceRevisionFromBuildInfo(clean); got != "0123456789abcdef" {
		t.Fatalf("clean source revision = %q", got)
	}

	dirty := *clean
	dirty.Settings = append([]debug.BuildSetting(nil), clean.Settings...)
	dirty.Settings[1].Value = "true"
	if got := sourceRevisionFromBuildInfo(&dirty); got != "" {
		t.Fatalf("dirty source revision = %q, want unresolved", got)
	}

	release := &debug.BuildInfo{Main: debug.Module{Version: "v2.0.0"}}
	if got := sourceRevisionFromBuildInfo(release); got != "v2.0.0" {
		t.Fatalf("release source revision = %q", got)
	}
}

func TestSourceRevisionPrefersInjectedRelease(t *testing.T) {
	clean := &debug.BuildInfo{
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "0123456789abcdef"},
			{Key: "vcs.modified", Value: "false"},
		},
	}
	if got := sourceRevisionForBuildInfo(clean, "v2.0.1"); got != "v2.0.1" {
		t.Fatalf("source revision = %q, want injected release", got)
	}

	clean.Settings[1].Value = "true"
	if got := sourceRevisionForBuildInfo(clean, "v2.0.1"); got != "" {
		t.Fatalf("dirty source revision = %q, want unresolved", got)
	}
}

func TestProjectUpdateRequiresInjectedRelease(t *testing.T) {
	clean := &debug.BuildInfo{
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "0123456789abcdef"},
			{Key: "vcs.modified", Value: "false"},
		},
	}
	if got := sourceRevisionForCommandBuildInfo("sdlc-project-update", clean, ""); got != "" {
		t.Fatalf("untagged updater revision = %q, want unresolved", got)
	}
	if got := sourceRevisionForCommandBuildInfo("sdlc-project-init", clean, ""); got != "0123456789abcdef" {
		t.Fatalf("untagged initializer revision = %q, want clean Git revision", got)
	}
	if got := sourceRevisionForCommandBuildInfo("sdlc-project-update", clean, "v2.1.0"); got != "v2.1.0" {
		t.Fatalf("released updater revision = %q, want v2.1.0", got)
	}
}

func TestResolveNoLaunch(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		requested bool
		want      bool
	}{
		{name: "initializer default", command: "sdlc-project-init", want: false},
		{name: "initializer flag", command: "sdlc-project-init", requested: true, want: true},
		{name: "updater default", command: "sdlc-project-update", want: true},
		{name: "updater cannot re-enable launch", command: "sdlc-project-update", requested: false, want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := resolveNoLaunch(test.command, test.requested); got != test.want {
				t.Fatalf("resolveNoLaunch(%q, %t) = %t, want %t", test.command, test.requested, got, test.want)
			}
		})
	}
}

func TestHelpDoesNotRequireSDLCRoot(t *testing.T) {
	schemaRoot := t.TempDir()
	installSchemaForCommandTest(t, schemaRoot)
	outputs := make([]string, 0, 2)
	for _, helpFlag := range []string{"-h", "--help"} {
		var output bytes.Buffer
		var errorOutput bytes.Buffer
		status := commandStatus("sdlc-project-update", []string{"--sdlc-root", schemaRoot, helpFlag}, strings.NewReader(""), &output, &errorOutput)
		if status != 0 {
			t.Fatalf("%s status = %d, stderr = %q", helpFlag, status, errorOutput.String())
		}
		if !strings.Contains(errorOutput.String(), "--project") || !strings.Contains(errorOutput.String(), "--version") {
			t.Fatalf("%s help is incomplete: %q", helpFlag, errorOutput.String())
		}
		outputs = append(outputs, errorOutput.String())
	}
	if outputs[0] != outputs[1] {
		t.Fatalf("-h and --help differ:\n-h:\n%s\n--help:\n%s", outputs[0], outputs[1])
	}
}

func installSchemaForCommandTest(t *testing.T, root string) {
	t.Helper()
	target := filepath.Join(root, "config", "project-init.schema.yaml")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(filepath.Join("..", "..", "src", "config", "project-init.schema.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, contents, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestCommandStatusDistinguishesUsageAndOperationalFailure(t *testing.T) {
	schemaRoot := t.TempDir()
	schemaTarget := filepath.Join(schemaRoot, "config", "project-init.schema.yaml")
	if err := os.MkdirAll(filepath.Dir(schemaTarget), 0o755); err != nil {
		t.Fatal(err)
	}
	schema, err := os.ReadFile(filepath.Join("..", "..", "src", "config", "project-init.schema.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(schemaTarget, schema, 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name      string
		arguments []string
		want      int
	}{
		{name: "missing flag value", arguments: []string{"--sdlc-root"}, want: 2},
		{name: "invalid flag", arguments: []string{"--sdlc-root", schemaRoot, "--not-a-flag"}, want: 2},
		{name: "unexpected positional argument", arguments: []string{"--sdlc-root", schemaRoot, "unexpected"}, want: 2},
		{name: "help does not hide positional argument", arguments: []string{"unexpected", "--help"}, want: 2},
		{name: "missing installed root", arguments: []string{"--sdlc-root", filepath.Join(t.TempDir(), "missing")}, want: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := commandStatus("sdlc-project-update", test.arguments, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
			if status != test.want {
				t.Fatalf("commandStatus() = %d, want %d", status, test.want)
			}
		})
	}
}

func TestUntaggedProjectUpdateRefusesBeforeProjectProcessing(t *testing.T) {
	schemaRoot := t.TempDir()
	schemaTarget := filepath.Join(schemaRoot, "config", "project-init.schema.yaml")
	if err := os.MkdirAll(filepath.Dir(schemaTarget), 0o755); err != nil {
		t.Fatal(err)
	}
	schema, err := os.ReadFile(filepath.Join("..", "..", "src", "config", "project-init.schema.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(schemaTarget, schema, 0o600); err != nil {
		t.Fatal(err)
	}
	var diagnostics bytes.Buffer
	status := commandStatus(
		"sdlc-project-update",
		[]string{"--sdlc-root", schemaRoot, "--project", t.TempDir()},
		strings.NewReader(""),
		&bytes.Buffer{},
		&diagnostics,
	)
	if status != 1 || !strings.Contains(diagnostics.String(), "requires a versioned SDLC build") {
		t.Fatalf("status = %d, stderr = %q", status, diagnostics.String())
	}
}
