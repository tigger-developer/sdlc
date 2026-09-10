// Command sdlc-harness is the internal fixed adapter used by SDLC workflows.
package main

import (
	"context"
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
		return errors.New("usage: sdlc-harness start|resume [options]")
	}
	action := arguments[0]
	if action != "start" && action != "resume" {
		return fmt.Errorf("unknown operation %q; use start or resume", action)
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
	if normalizedPhase == "audit" {
		base, loadErr := loadAuditPrompt(*auditPrompts, normalizedGate)
		if loadErr != nil {
			return loadErr
		}
		prompt = append([]byte(base+"\n\nOperator-supplied audit context:\n"), prompt...)
	}
	evidence, err := harness.CaptureEvidence(projectRoot, inputs)
	if err != nil {
		return err
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
	var previousEvidence []harness.EvidenceFile
	if normalizedPhase == "audit" && strings.TrimSpace(*auditRecord) != "" {
		if strings.TrimSpace(*workItem) == "" {
			return errors.New("--work-item is required with --audit-record")
		}
		if err := harness.MigrateLegacyAudit(*auditRecord); err != nil {
			return err
		}
		entry, found, recordErr := harness.ReadAuditEntry(*auditRecord, *workItem, normalizedGate)
		if recordErr != nil {
			return recordErr
		}
		previousEvidence = entry.LatestEvidence()
		if action == "start" && found && entry.SessionID != "" {
			return fmt.Errorf("audit session already exists for %s/%s; resume session %s", *workItem, normalizedGate, entry.SessionID)
		}
		if action == "resume" {
			if !found || entry.SessionID == "" {
				return fmt.Errorf("no recorded audit session for %s/%s; start a new audit first", *workItem, normalizedGate)
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
	if action == "start" && identity == "" {
		identity, err = harness.NewSessionIdentity()
		if err != nil {
			return err
		}
	}
	evidencePrompt, err := evidence.Prompt(previousEvidence)
	if err != nil {
		return err
	}
	request := harness.Request{
		Harness: config.Harness, Provider: config.Provider, Model: config.Model,
		Prompt: string(prompt) + "\n\n" + evidencePrompt, Directory: projectRoot, ResultFile: resultPath, SessionID: identity,
	}
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()
	result, err := harness.Execute(ctx, request, action == "resume", &evidence, nil, errorOutput)
	if err != nil {
		return err
	}
	if config.Harness != "" && normalizedPhase == "audit" {
		if err := harness.ValidateCompositeVerdict(result.Response); err != nil {
			return err
		}
		if !strings.Contains(result.Response, "GATE: "+normalizedGate) {
			return fmt.Errorf("audit response gate does not match requested %s gate", normalizedGate)
		}
	}
	if normalizedPhase == "audit" && strings.TrimSpace(*auditRecord) != "" {
		status := "active"
		if strings.Contains(result.Response, "VERDICT: PASS") || strings.Contains(result.Response, "VERDICT: PROVISIONAL PASS") {
			status = "passed"
		}
		entry, _, recordErr := harness.ReadAuditEntry(*auditRecord, *workItem, normalizedGate)
		if recordErr != nil {
			return recordErr
		}
		if entry.ExternalRound < config.MaxRounds {
			entry.ExternalRound++
		}
		entry.WorkItem, entry.Gate, entry.SessionID, entry.Status = *workItem, normalizedGate, result.SessionID, status
		entry.Revision = auditField(result.Response, "REVISION")
		entry.Verdict = auditField(result.Response, "VERDICT")
		entry.Response = result.Response
		entry.History = append(entry.History, harness.AuditRound{Round: entry.ExternalRound, Revision: entry.Revision, Verdict: entry.Verdict, Response: result.Response, Evidence: evidence.Files})
		if entry.ExternalRound >= config.MaxRounds && status == "active" {
			entry.Status = "exhausted"
		}
		if err := harness.WriteAuditEntry(*auditRecord, entry); err != nil {
			return err
		}
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
	Version              int    `yaml:"version"`
	EvidenceInstructions string `yaml:"evidence_instructions"`
	Gates                map[string]struct {
		Prompt string `yaml:"prompt"`
	} `yaml:"gates"`
}

func loadAuditPrompt(explicit, gate string) (string, error) {
	path := explicit
	if path == "" {
		if executable, err := os.Executable(); err == nil {
			path = filepath.Join(filepath.Dir(executable), "..", "prompts", "audits.yaml")
		}
	}
	if path == "" {
		return "", errors.New("audit prompt registry path is unavailable")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading audit prompt registry %s: %w", path, err)
	}
	var document auditPromptDocument
	if err := yaml.Unmarshal(contents, &document); err != nil {
		return "", fmt.Errorf("parsing audit prompt registry %s: %w", path, err)
	}
	value, found := document.Gates[gate]
	if !found || strings.TrimSpace(value.Prompt) == "" {
		return "", fmt.Errorf("audit prompt registry has no %s gate prompt", gate)
	}
	return strings.TrimSpace(value.Prompt + "\n\n" + document.EvidenceInstructions), nil
}

func printTopLevelHelp(output io.Writer) {
	fmt.Fprint(output, helpText)
}
