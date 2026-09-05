package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/tigger-developer/sdlc/internal/auditrunner"
)

var buildVersion = "devel"

type repeatedFlag []string

func (values *repeatedFlag) String() string { return fmt.Sprint([]string(*values)) }
func (values *repeatedFlag) Set(value string) error {
	*values = append(*values, value)
	return nil
}

func main() {
	os.Exit(execute(os.Args[1:], os.Stdout, os.Stderr, auditrunner.Run))
}

func execute(arguments []string, stdout, stderr io.Writer, runAudit func(auditrunner.Options) error) int {
	if len(arguments) == 1 && (arguments[0] == "--help" || arguments[0] == "-h") {
		printUsage(stdout)
		return 0
	}
	if len(arguments) == 1 && arguments[0] == "--version" {
		fmt.Fprintf(stdout, "sdlc-audit %s\n", buildVersion)
		return 0
	}
	options, err := parseOptionsTo(arguments, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "sdlc-audit: %v\n", err)
		printUsage(stderr)
		return 2
	}
	if runAudit == nil {
		runAudit = auditrunner.Run
	}
	if err := runAudit(options); err != nil {
		fmt.Fprintf(stderr, "sdlc-audit: %v\n", err)
		return 1
	}
	return 0
}

func parseOptions(arguments []string) (auditrunner.Options, error) {
	return parseOptionsTo(arguments, io.Discard)
}

func parseOptionsTo(arguments []string, output io.Writer) (auditrunner.Options, error) {
	if len(arguments) == 0 || arguments[0] == "" || arguments[0][0] == '-' {
		return auditrunner.Options{}, errors.New("an audit name must be the first argument")
	}
	options := auditrunner.Options{AuditName: arguments[0]}
	flags := flag.NewFlagSet("sdlc-audit "+options.AuditName, flag.ContinueOnError)
	flags.SetOutput(output)
	flags.StringVar(&options.ProjectRoot, "project", ".", "project root")
	flags.StringVar(&options.SDLCRoot, "sdlc-root", "", "canonical SDLC root (default ~/.agents/sdlc)")
	flags.StringVar(&options.UserConfigPath, "user-config", "", "user SDLC environment file (default ~/.agents/.env)")
	flags.StringVar(&options.Harness, "harness", "", "override audit harness")
	flags.StringVar(&options.Provider, "provider", "", "override provider for harnesses that accept one")
	flags.StringVar(&options.Model, "model", "", "override audit model")
	flags.DurationVar(&options.Timeout, "timeout", 0, "override audit timeout with a whole-second duration (for example 4m)")
	flags.Var((*repeatedFlag)(&options.Artifacts), "artifact", "artefact to audit; repeat for multiple files")
	flags.Var((*repeatedFlag)(&options.Context), "context", "authorized context file; repeat for multiple files")
	flags.Var((*repeatedFlag)(&options.ExternalContext), "external-context", "exact external context file; repeat for multiple files")
	if err := flags.Parse(arguments[1:]); err != nil {
		return auditrunner.Options{}, err
	}
	flags.Visit(func(selected *flag.Flag) {
		if selected.Name == "timeout" {
			options.TimeoutSet = true
		}
	})
	if options.TimeoutSet && (options.Timeout < time.Second || options.Timeout%time.Second != 0) {
		return auditrunner.Options{}, errors.New("--timeout must be a whole-second duration of at least one second")
	}
	if flags.NArg() != 0 {
		return auditrunner.Options{}, fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	return options, nil
}

func printUsage(output io.Writer) {
	fmt.Fprintln(output, "usage: sdlc-audit AUDIT --artifact FILE [--artifact FILE ...] [--context FILE ...] [--external-context FILE ...]")
	fmt.Fprintln(output, "audits: audit-spec, audit-design, audit-tests, audit-code")
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "inputs:")
	fmt.Fprintln(output, "  --artifact FILE   artefact to audit; repeat for multiple files")
	fmt.Fprintln(output, "  --context FILE    authorized context file; repeat for multiple files")
	fmt.Fprintln(output, "  --external-context FILE  exact external context file; repeat for multiple files")
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "configuration:")
	fmt.Fprintln(output, "  --project DIR       project root (default .)")
	fmt.Fprintln(output, "  --sdlc-root DIR     canonical SDLC root (default ~/.agents/sdlc)")
	fmt.Fprintln(output, "  --user-config FILE  user SDLC environment file (default ~/.agents/.env)")
	fmt.Fprintln(output, "  --harness NAME      override audit harness: codex, claude, or hermes")
	fmt.Fprintln(output, "  --provider NAME     override provider for harnesses that accept one")
	fmt.Fprintln(output, "  --model NAME        override audit model")
	fmt.Fprintf(output, "  --timeout DURATION  override audit timeout with a whole-second duration (default %s)\n", 5*time.Minute)
}
