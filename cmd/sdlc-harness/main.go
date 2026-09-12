// Command sdlc-harness is the internal fixed adapter used by SDLC workflows.
package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/tigger-developer/sdlc/internal/harness"
)

var buildRelease string

//go:embed help.md
var helpText string

type inputList []string

func (values *inputList) String() string { return strings.Join(*values, ",") }
func (values *inputList) Set(value string) error {
	*values = append(*values, value)
	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "sdlc-harness: %v\n", err)
		os.Exit(1)
	}
}

func run(arguments []string, input io.Reader, output, errorOutput io.Writer) (returnErr error) {
	if len(arguments) == 1 && (arguments[0] == "-h" || arguments[0] == "--help") {
		printTopLevelHelp(output)
		return nil
	}
	if len(arguments) == 1 && (arguments[0] == "--version" || arguments[0] == "-version") {
		version := buildRelease
		if version == "" {
			version = "devel"
		}
		fmt.Fprintf(output, "sdlc-harness %s\n", version)
		return nil
	}
	if len(arguments) == 0 {
		return errors.New("usage: sdlc-harness start|resume|goal-config [options]")
	}
	action := arguments[0]
	if action == "goal-config" {
		return runGoalConfig(arguments[1:], output, errorOutput)
	}
	if action != "start" && action != "resume" {
		return fmt.Errorf("unknown operation %q; use start, resume or goal-config", action)
	}
	flags := flag.NewFlagSet("sdlc-harness "+action, flag.ContinueOnError)
	flags.SetOutput(errorOutput)
	project := flags.String("project", ".", "project root")
	globalConfig := flags.String("global-config", "", "global SDLC YAML configuration")
	phase := flags.String("phase", "audit", "SDLC phase: definition, build, or audit")
	gate := flags.String("gate", "", "composite audit gate: definition or implementation")
	auditPrompts := flags.String("audit-prompts", "", "audit prompt YAML (defaults to the installed prompts/audits.yaml)")
	auditRecord := flags.String("audit-record", "", "YAML audit session record to update")
	workItem := flags.String("work-item", "", "stable work-item identifier for the audit record")
	harnessName := flags.String("harness", "", "harness override")
	provider := flags.String("provider", "", "provider override where supported")
	model := flags.String("model", "", "model override")
	timeout := flags.Duration("timeout", 0, "execution timeout")
	session := flags.String("session", "", "stable session identity for resume")
	var inputs inputList
	flags.Var(&inputs, "input", "exact evidence file to include; repeat as needed")
	flags.Usage = func() {
		printTopLevelHelp(errorOutput)
		flags.PrintDefaults()
	}
	if err := flags.Parse(arguments[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	if action == "start" && strings.TrimSpace(*session) != "" {
		return errors.New("start must not receive --session; start creates a fresh external session identity")
	}
	normalizedPhase := strings.ToLower(strings.TrimSpace(*phase))
	normalizedGate := strings.ToLower(strings.TrimSpace(*gate))
	if normalizedPhase == "audit" && normalizedGate != "definition" && normalizedGate != "implementation" {
		return errors.New("audit phase requires --gate definition or --gate implementation")
	}
	projectRoot, err := filepath.Abs(*project)
	if err != nil {
		return err
	}
	config, err := harness.ResolveConfig(harness.ConfigOptions{
		ProjectRoot: projectRoot, GlobalPath: *globalConfig, Phase: normalizedPhase,
		Harness: *harnessName, Provider: *provider, Model: *model, Timeout: *timeout,
	})
	if err != nil {
		return err
	}
	prompt, err := io.ReadAll(io.LimitReader(input, 2*1024*1024+1))
	if err != nil {
		return fmt.Errorf("reading prompt: %w", err)
	}
	if len(prompt) == 0 || len(prompt) > 2*1024*1024 {
		if normalizedPhase != "audit" {
			return errors.New("prompt must contain between 1 byte and 2 MiB")
		}
	}
	var registry auditPromptDocument
	if normalizedPhase == "audit" {
		var loadErr error
		registry, loadErr = readAuditPrompts(*auditPrompts)
		if loadErr != nil {
			return loadErr
		}
		base, loadErr := registry.prompt(normalizedGate)
		if loadErr != nil {
			return loadErr
		}
		prompt = append([]byte(base+"\n\nOperator-supplied audit context:\n"), prompt...)
	}
	evidence, err := harness.CaptureEvidence(projectRoot, inputs)
	if err != nil {
		return err
	}
	if normalizedPhase == "audit" && *auditRecord != "" {
		recordPath, pathErr := filepath.Abs(*auditRecord)
		if pathErr != nil {
			return pathErr
		}
		for _, file := range evidence.Files {
			if file.Path == recordPath {
				return errors.New("--audit-record is harness-owned output; do not also supply it as --input")
			}
		}
	}
	resultFile, err := os.CreateTemp(os.TempDir(), "sdlc-harness-result-")
	if err != nil {
		return err
	}
	resultPath := resultFile.Name()
	if err := resultFile.Close(); err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(resultPath); err != nil && !errors.Is(err, os.ErrNotExist) && returnErr == nil {
			returnErr = fmt.Errorf("removing harness result file: %w", err)
		}
	}()
	identity := *session
	var entry harness.AuditEntry
	cacheKey := ""
	if normalizedPhase == "audit" && strings.TrimSpace(*auditRecord) != "" {
		if strings.TrimSpace(*workItem) == "" {
			return errors.New("--work-item is required with --audit-record")
		}
		if err := harness.MigrateLegacyAudit(*auditRecord); err != nil {
			return err
		}
		var found bool
		var recordErr error
		entry, found, recordErr = harness.ReadAuditEntry(*auditRecord, *workItem, normalizedGate)
		if recordErr != nil {
			return recordErr
		}
		cacheKey, err = auditCacheKey(*workItem, normalizedGate, string(prompt), registry, evidence, config)
		if err != nil {
			return err
		}
		if identity != "" && identity != entry.SessionID {
			return fmt.Errorf("supplied session does not match recorded audit session %s", entry.SessionID)
		}
		if cached, ok := cachedAudit(entry, cacheKey); ok {
			if err := evidence.Verify(); err != nil {
				return err
			}
			fmt.Fprintf(errorOutput, "AUDIT CACHE HIT: %s/%s round=%d; no provider invoked; no round consumed; record=%s\n", *workItem, normalizedGate, cached.Round, *auditRecord)
			_, err = fmt.Fprintln(output, cached.Response)
			return err
		}
		if action == "start" && found && (entry.SessionID != "" || entry.ExternalRound > 0) {
			if entry.SessionID == "" {
				return errors.New("prior audit attempt has no native session identity; operator recovery required, not a replacement session")
			}
			return fmt.Errorf("audit session already exists for %s/%s; resume session %s", *workItem, normalizedGate, entry.SessionID)
		}
		if action == "resume" {
			if entry.Status == "running" {
				return errors.New("recorded audit is running or was interrupted before recording its outcome; inspect it before resuming")
			}
			if !found || entry.SessionID == "" {
				return fmt.Errorf("no resumable native audit session for %s/%s; inspect the record before starting or recovering an audit", *workItem, normalizedGate)
			}
			if entry.ExternalRound >= config.MaxRounds {
				return fmt.Errorf("audit session for %s/%s has reached max_rounds=%d; operator decision required", *workItem, normalizedGate, config.MaxRounds)
			}
			if identity == "" {
				identity = entry.SessionID
			} else if identity != entry.SessionID {
				return fmt.Errorf("supplied session does not match recorded audit session %s", entry.SessionID)
			}
		}
	}
	request := harness.Request{
		Harness: config.Harness, Provider: config.Provider, Model: config.Model,
		Prompt: string(prompt), Directory: projectRoot, ResultFile: resultPath, SessionID: identity,
		Evidence: evidence.Files,
	}
	recordPath := ""
	if normalizedPhase == "audit" {
		recordPath = strings.TrimSpace(*auditRecord)
	}
	runner := recoveryRun{config: config, request: request, evidence: evidence, registry: registry,
		path: recordPath, work: *workItem, gate: normalizedGate, cacheKey: cacheKey, audit: normalizedPhase == "audit", diagnostics: errorOutput}
	result, err := runner.execute(entry, action == "resume")
	if err != nil {
		return err
	}
	fmt.Fprintf(errorOutput, "SESSION_ID: %s\n", result.SessionID)
	_, err = fmt.Fprintln(output, result.Response)
	return err
}

func auditField(response, name string) string {
	for _, line := range strings.Split(response, "\n") {
		if strings.HasPrefix(line, name+":") {
			return strings.TrimSpace(strings.TrimPrefix(line, name+":"))
		}
	}
	return ""
}

type auditPromptDocument struct {
	Version                     int    `yaml:"version"`
	EvidenceInstructions        string `yaml:"evidence_instructions"`
	TimeoutResumeInstructions   string `yaml:"timeout_resume_instructions"`
	TimeoutMessage              string `yaml:"timeout_message"`
	TimeoutBlockedMessage       string `yaml:"timeout_blocked_message"`
	SessionRecoveryInstructions string `yaml:"session_recovery_instructions"`
	Gates                       map[string]struct {
		Prompt string `yaml:"prompt"`
	} `yaml:"gates"`
}

func readAuditPrompts(explicit string) (auditPromptDocument, error) {
	var document auditPromptDocument
	path := explicit
	if path == "" {
		if executable, err := os.Executable(); err == nil {
			path = filepath.Join(filepath.Dir(executable), "..", "prompts", "audits.yaml")
		}
	}
	if path == "" {
		return document, errors.New("audit prompt registry path is unavailable")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return document, fmt.Errorf("reading audit prompt registry %s: %w", path, err)
	}
	if err := yaml.Unmarshal(contents, &document); err != nil {
		return document, fmt.Errorf("parsing audit prompt registry %s: %w", path, err)
	}
	return document, nil
}

func (document auditPromptDocument) prompt(gate string) (string, error) {
	value, found := document.Gates[gate]
	if !found || strings.TrimSpace(value.Prompt) == "" {
		return "", fmt.Errorf("audit prompt registry has no %s gate prompt", gate)
	}
	return strings.TrimSpace(value.Prompt + "\n\n" + document.EvidenceInstructions), nil
}

func printTopLevelHelp(output io.Writer) {
	fmt.Fprint(output, helpText)
}
