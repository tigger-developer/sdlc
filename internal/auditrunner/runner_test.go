package auditrunner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestRunUsesHermesWithProjectProviderAndModelAndBoundedInputs(t *testing.T) {
	fixture := newFixture(t, "audit-design")
	mustWrite(t, fixture.userConfig, `SDLC_AUDIT_HARNESS="hermes"
SDLC_AUDIT_PROVIDER="openai-codex"
SDLC_AUDIT_MODEL="gpt-5.6-luna"
`)
	mustWrite(t, filepath.Join(fixture.project, ".env"), `SDLC_AUDIT_PROVIDER="nous"
SDLC_AUDIT_MODEL="z-ai/glm-5.3"
`)
	mustWrite(t, filepath.Join(fixture.project, "architecture.md"), "AUTHORIZED CONTEXT")
	mustWrite(t, filepath.Join(fixture.project, "secret.md"), "MUST NOT LEAK")

	var invokedArgs []string
	var invokedPrompt string
	var invokedEnvironment []string
	var invokedDirectory string
	runner := func(ctx context.Context, name string, args []string, cwd string, environment []string, stdin io.Reader, stdout, stderr io.Writer) error {
		if name != "hermes" {
			return fmt.Errorf("command = %q, want hermes", name)
		}
		if _, ok := ctx.Deadline(); !ok {
			return errors.New("audit command has no deadline")
		}
		invokedArgs = slices.Clone(args)
		invokedEnvironment = slices.Clone(environment)
		invokedDirectory = cwd
		contents, err := io.ReadAll(stdin)
		if err != nil {
			return err
		}
		invokedPrompt = string(contents)
		_, _ = io.WriteString(stderr, "session_id: suppressed")
		_, err = io.WriteString(stdout, "AUDIT: audit-design\nAUDITOR_PROVIDER: nous\nAUDITOR_MODEL: z-ai/glm-5.3\nVERDICT: PASS\n")
		return err
	}

	var output bytes.Buffer
	var diagnostics bytes.Buffer
	options := fixture.options(runner)
	options.Context = []string{"architecture.md"}
	options.Output = &output
	options.ErrorOutput = &diagnostics
	options.LookupEnv = mapLookup(map[string]string{"PATH": "/bin", "HOME": "/home/test", "SECRET_TOKEN": "do-not-pass"})
	if err := Run(options); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	wantArgs := []string{"chat", "--quiet", "--query-file", "-", "--in", invokedDirectory, "--provider", "nous", "--model", "z-ai/glm-5.3", "--toolsets", "", "--ignore-rules", "--source", "tool", "--run-budget", "60"}
	if !slices.Equal(invokedArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", invokedArgs, wantArgs)
	}
	if strings.Contains(strings.Join(invokedArgs, " "), "--resume") || strings.Contains(strings.Join(invokedArgs, " "), "--continue") {
		t.Fatalf("Hermes invocation resumes prior context: %#v", invokedArgs)
	}
	if !slices.Contains(invokedEnvironment, "PATH=/bin") || !slices.Contains(invokedEnvironment, "HOME=/home/test") {
		t.Fatalf("minimal environment lacks required entries: %#v", invokedEnvironment)
	}
	if slices.Contains(invokedEnvironment, "SECRET_TOKEN=do-not-pass") {
		t.Fatalf("unrelated secret entered child environment: %#v", invokedEnvironment)
	}
	for _, expected := range []string{"AUDIT PROMPT", "COMMON VERDICT CONTRACT", "CANDIDATE", "AUTHORIZED CONTEXT"} {
		if !strings.Contains(invokedPrompt, expected) {
			t.Errorf("prompt lacks supplied content %q", expected)
		}
	}
	if strings.Contains(invokedPrompt, "MUST NOT LEAK") {
		t.Error("prompt contains content from an unlisted file")
	}
	if diagnostics.Len() != 0 {
		t.Fatalf("diagnostics = %q", diagnostics.String())
	}
	if output.String() != "AUDIT: audit-design\nAUDITOR_PROVIDER: nous\nAUDITOR_MODEL: z-ai/glm-5.3\nVERDICT: PASS\n" {
		t.Fatalf("output = %q", output.String())
	}
}

func TestRunUsesCodexAndIgnoresConfiguredProvider(t *testing.T) {
	fixture := newFixture(t, "audit-design")
	mustWrite(t, fixture.userConfig, "SDLC_AUDIT_HARNESS=codex\nSDLC_AUDIT_PROVIDER=nous\nSDLC_AUDIT_MODEL=gpt-5.6-luna\n")
	var invokedArgs []string
	runner := func(_ context.Context, name string, args []string, _ string, _ []string, _ io.Reader, stdout, _ io.Writer) error {
		if name != "codex" {
			return fmt.Errorf("command = %q, want codex", name)
		}
		invokedArgs = slices.Clone(args)
		_, err := io.WriteString(stdout, "AUDIT: audit-design\nAUDITOR_PROVIDER: openai-codex\nAUDITOR_MODEL: gpt-5.6-luna\nVERDICT: PASS\n")
		return err
	}
	if err := Run(fixture.options(runner)); err != nil {
		t.Fatal(err)
	}
	want := []string{"exec", "--ephemeral", "--ignore-user-config", "--ignore-rules", "--skip-git-repo-check", "--sandbox", "read-only", "--model", "gpt-5.6-luna", "-"}
	if !slices.Equal(invokedArgs, want) || slices.Contains(invokedArgs, "nous") || slices.Contains(invokedArgs, "--provider") {
		t.Fatalf("Codex args = %#v, want %#v without provider", invokedArgs, want)
	}
}

func TestRunUsesClaudeAndIgnoresConfiguredProvider(t *testing.T) {
	fixture := newFixture(t, "audit-spec")
	mustWrite(t, fixture.userConfig, "SDLC_AUDIT_HARNESS=claude\nSDLC_AUDIT_PROVIDER=nous\nSDLC_AUDIT_MODEL=sonnet\n")
	var invokedArgs []string
	runner := func(_ context.Context, name string, args []string, _ string, _ []string, _ io.Reader, stdout, _ io.Writer) error {
		if name != "claude" {
			return fmt.Errorf("command = %q, want claude", name)
		}
		invokedArgs = slices.Clone(args)
		_, err := io.WriteString(stdout, "AUDIT: audit-spec\nAUDITOR_PROVIDER: anthropic\nAUDITOR_MODEL: sonnet\nVERDICT: PASS\n")
		return err
	}
	if err := Run(fixture.options(runner)); err != nil {
		t.Fatal(err)
	}
	want := []string{"--print", "--output-format", "text", "--model", "sonnet", "--no-session-persistence", "--safe-mode", "--permission-mode", "dontAsk", "--tools", ""}
	if !slices.Equal(invokedArgs, want) || slices.Contains(invokedArgs, "nous") || slices.Contains(invokedArgs, "--provider") {
		t.Fatalf("Claude args = %#v, want %#v without provider", invokedArgs, want)
	}
}

func TestRunUsesHermesWhenHarnessIsUnset(t *testing.T) {
	fixture := newFixture(t, "audit-spec")
	mustWrite(t, fixture.userConfig, "SDLC_AUDIT_PROVIDER=openai-codex\nSDLC_AUDIT_MODEL=gpt-5.6-luna\n")
	var invoked string
	runner := func(ctx context.Context, name string, args []string, cwd string, environment []string, stdin io.Reader, stdout, stderr io.Writer) error {
		invoked = name
		_, err := io.WriteString(stdout, "AUDIT: audit-spec\nAUDITOR_PROVIDER: openai-codex\nAUDITOR_MODEL: gpt-5.6-luna\nVERDICT: PASS\n")
		return err
	}
	options := fixture.options(runner)
	if err := Run(options); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if invoked != "hermes" {
		t.Fatalf("command = %q, want hermes", invoked)
	}
}

func TestResolveConfigurationUsesEnvironmentBeforeProjectAndAuditHarnessFallsBack(t *testing.T) {
	fixture := newFixture(t, "audit-spec")
	mustWrite(t, fixture.userConfig, "SDLC_AGENT_HARNESS=codex\nSDLC_AUDIT_PROVIDER=user\nSDLC_AUDIT_MODEL=user-model\n")
	mustWrite(t, filepath.Join(fixture.project, ".env"), "SDLC_AUDIT_PROVIDER=project\nSDLC_AUDIT_MODEL=project-model\n")
	options := defaults(fixture.options(passingRunner("audit-spec", "environment", "environment-model")))
	options.LookupEnv = mapLookup(map[string]string{
		"SDLC_AUDIT_PROVIDER": "environment",
		"SDLC_AUDIT_MODEL":    "environment-model",
	})
	config, err := resolveConfiguration(options, fixture.project)
	if err != nil {
		t.Fatal(err)
	}
	if config.harness != "codex" || config.provider != "environment" || config.model != "environment-model" {
		t.Fatalf("resolved audit configuration = %#v", config)
	}
}

func TestEffectiveConfigurationRequiresProviderOnlyForHermes(t *testing.T) {
	if _, err := effectiveConfiguration(configuration{harness: "hermes", model: "model"}); err == nil || !strings.Contains(err.Error(), "SDLC_AUDIT_PROVIDER") {
		t.Fatalf("Hermes without provider error = %v", err)
	}
	codex, err := effectiveConfiguration(configuration{harness: "codex", provider: "ignored", model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if codex.provider != "openai-codex" {
		t.Fatalf("Codex effective provider = %q", codex.provider)
	}
	claude, err := effectiveConfiguration(configuration{harness: "claude", provider: "ignored", model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if claude.provider != "anthropic" {
		t.Fatalf("Claude effective provider = %q", claude.provider)
	}
}

func TestRunAllowsExactExternalContextAndRejectsOtherExternalFiles(t *testing.T) {
	fixture := newFixture(t, "audit-spec")
	mustWrite(t, fixture.userConfig, auditEnvironment())
	external := t.TempDir()
	denied := t.TempDir()
	contract := filepath.Join(external, "contract.md")
	sibling := filepath.Join(external, "unrequested.md")
	mustWrite(t, contract, "EXTERNAL CONTRACT")
	mustWrite(t, sibling, "UNREQUESTED")
	mustWrite(t, filepath.Join(denied, "secret.md"), "SECRET")

	var prompt string
	options := fixture.options(func(_ context.Context, _ string, _ []string, _ string, _ []string, stdin io.Reader, stdout, _ io.Writer) error {
		contents, err := io.ReadAll(stdin)
		if err != nil {
			return err
		}
		prompt = string(contents)
		_, err = io.WriteString(stdout, "AUDIT: audit-spec\nAUDITOR_PROVIDER: openai-codex\nAUDITOR_MODEL: gpt-5.6-luna\nVERDICT: PASS\n")
		return err
	})
	options.ExternalContext = []string{contract}
	if err := Run(options); err != nil {
		t.Fatalf("Run() with exact external context error = %v", err)
	}
	if !strings.Contains(prompt, "EXTERNAL CONTRACT") || strings.Contains(prompt, "UNREQUESTED") {
		t.Fatalf("prompt did not remain bounded to the exact external file: %q", prompt)
	}
	options.Context = []string{filepath.Join(denied, "secret.md")}
	options.ExternalContext = nil
	if err := Run(options); err == nil || !strings.Contains(err.Error(), "outside authorized directories") {
		t.Fatalf("Run() error = %v, want unauthorized-root error", err)
	}
}

func TestRunRejectsSymlinkInput(t *testing.T) {
	fixture := newFixture(t, "audit-spec")
	mustWrite(t, fixture.userConfig, auditEnvironment())
	target := filepath.Join(fixture.project, "target.md")
	mustWrite(t, target, "TARGET")
	link := filepath.Join(fixture.project, "link.md")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	options := fixture.options(nil)
	options.Context = []string{"link.md"}
	if err := Run(options); err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("Run() error = %v, want symbolic-link error", err)
	}
}

func TestRunReportsCleanupFailure(t *testing.T) {
	fixture := newFixture(t, "audit-code")
	mustWrite(t, fixture.userConfig, auditEnvironment())
	options := fixture.options(passingRunner("audit-code", "openai-codex", "gpt-5.6-luna"))
	options.RemoveAll = func(string) error { return errors.New("cleanup denied") }
	if err := Run(options); err == nil || !strings.Contains(err.Error(), "cleaning isolated audit directory") {
		t.Fatalf("Run() error = %v, want cleanup error", err)
	}
}

func TestRunRejectsReportedIdentityMismatch(t *testing.T) {
	fixture := newFixture(t, "audit-code")
	mustWrite(t, fixture.userConfig, auditEnvironment())
	err := Run(fixture.options(passingRunner("audit-code", "openai-codex", "gpt-5")))
	if err == nil || !strings.Contains(err.Error(), "reported model") {
		t.Fatalf("Run() error = %v, want reported model mismatch", err)
	}
}

func TestRunReturnsValidFailVerdict(t *testing.T) {
	fixture := newFixture(t, "audit-tests")
	mustWrite(t, fixture.userConfig, auditEnvironment())
	runner := func(ctx context.Context, name string, args []string, cwd string, environment []string, stdin io.Reader, stdout, stderr io.Writer) error {
		_, err := io.WriteString(stdout, "AUDIT: audit-tests\nAUDITOR_PROVIDER: openai-codex\nAUDITOR_MODEL: gpt-5.6-luna\nVERDICT: FAIL\n1. [BLOCKING] Missing boundary test\n")
		return err
	}
	if err := Run(fixture.options(runner)); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunDiscardsHermesReasoningPrefixAndEmitsOnlyVerdict(t *testing.T) {
	fixture := newFixture(t, "audit-code")
	mustWrite(t, fixture.userConfig, auditEnvironment())
	runner := func(ctx context.Context, name string, args []string, cwd string, environment []string, stdin io.Reader, stdout, stderr io.Writer) error {
		_, err := io.WriteString(stdout, "┌─ Reasoning ─┐\nI considered AUDIT: audit-code but this is not the report.\n└─────────────┘\nAUDIT: audit-code\nAUDITOR_PROVIDER: openai-codex\nAUDITOR_MODEL: gpt-5.6-luna\nVERDICT: PASS\n")
		return err
	}
	var output bytes.Buffer
	options := fixture.options(runner)
	options.Output = &output
	if err := Run(options); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	want := "AUDIT: audit-code\nAUDITOR_PROVIDER: openai-codex\nAUDITOR_MODEL: gpt-5.6-luna\nVERDICT: PASS\n"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}

type fixture struct {
	project       string
	sdlcRoot      string
	userConfig    string
	auditName     string
	temporaryRoot string
}

func newFixture(t *testing.T, auditName string) fixture {
	t.Helper()
	project := t.TempDir()
	home := t.TempDir()
	sdlcRoot := filepath.Join(home, ".agents", "sdlc")
	temporaryRoot := filepath.Join(t.TempDir(), "audit-temp")
	mustWrite(t, filepath.Join(sdlcRoot, "prompts", "audits", auditName+".md"), "AUDIT PROMPT")
	mustWrite(t, filepath.Join(sdlcRoot, "AUDITS.md"), "COMMON VERDICT CONTRACT")
	loader, err := os.ReadFile(filepath.Join("..", "..", "src", "libexec", "load-sdlc-env.sh"))
	if err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(sdlcRoot, "libexec", "load-sdlc-env.sh"), string(loader))
	mustWrite(t, filepath.Join(project, "artifact.md"), "CANDIDATE")
	if err := os.MkdirAll(temporaryRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	return fixture{project: project, sdlcRoot: sdlcRoot, userConfig: filepath.Join(home, ".agents", ".env"), auditName: auditName, temporaryRoot: temporaryRoot}
}

func (f fixture) options(run RunCommand) Options {
	return Options{
		ProjectRoot: f.project, SDLCRoot: f.sdlcRoot, UserConfigPath: f.userConfig,
		AuditName: f.auditName, Artifacts: []string{"artifact.md"},
		Output: io.Discard, ErrorOutput: io.Discard, RunCommand: run,
		TemporaryRoot: f.temporaryRoot, Timeout: time.Minute,
	}
}

func passingRunner(auditName, provider, model string) RunCommand {
	return func(ctx context.Context, name string, args []string, cwd string, environment []string, stdin io.Reader, stdout, stderr io.Writer) error {
		_, err := fmt.Fprintf(stdout, "AUDIT: %s\nAUDITOR_PROVIDER: %s\nAUDITOR_MODEL: %s\nVERDICT: PASS\n", auditName, provider, model)
		return err
	}
}

func auditEnvironment() string {
	return "SDLC_AUDIT_HARNESS=hermes\nSDLC_AUDIT_PROVIDER=openai-codex\nSDLC_AUDIT_MODEL=gpt-5.6-luna\n"
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) { value, ok := values[key]; return value, ok }
}

func mustWrite(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
