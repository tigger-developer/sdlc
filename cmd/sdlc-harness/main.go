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
		return errors.New("prompt must contain between 1 byte and 2 MiB")
	}
	bundleInputs := make([]harness.Input, 0, len(inputs))
	for index, supplied := range inputs {
		path := supplied
		if !filepath.IsAbs(path) {
			path = filepath.Join(projectRoot, path)
		}
		bundleInputs = append(bundleInputs, harness.Input{Name: fmt.Sprintf("%03d-%s", index+1, filepath.Base(path)), Source: path})
	}
	if len(bundleInputs) == 0 {
		return errors.New("at least one --input evidence file is required")
	}
	bundle, err := harness.CreateBundle(os.TempDir(), bundleInputs)
	if err != nil {
		return err
	}
	defer func() {
		if err := bundle.Cleanup(); err != nil && returnErr == nil {
			returnErr = err
		}
	}()
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
	if action == "start" && identity == "" {
		identity, err = harness.NewSessionIdentity()
		if err != nil {
			return err
		}
	}
	request := harness.Request{
		Harness: config.Harness, Provider: config.Provider, Model: config.Model,
		Prompt: string(prompt), Bundle: bundle.Path, ResultFile: resultPath, SessionID: identity,
	}
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()
	result, err := harness.Execute(ctx, request, action == "resume", &bundle, nil, errorOutput)
	if err != nil {
		return err
	}
	if config.Harness != "" && normalizedPhase == "audit" {
		if err := harness.ValidateCompositeVerdict(result.Response); err != nil {
			return err
		}
	}
	fmt.Fprintf(errorOutput, "SESSION_ID: %s\n", result.SessionID)
	_, err = fmt.Fprintln(output, result.Response)
	return err
}

func printTopLevelHelp(output io.Writer) {
	fmt.Fprint(output, helpText)
}
