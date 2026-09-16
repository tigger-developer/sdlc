// Command sdlc-validate checks Org records and audit readiness without a model.
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tigger-developer/sdlc/internal/readiness"
	"gopkg.in/yaml.v3"
)

//go:embed help.md
var helpText string
var buildRelease = "devel"

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, out, diagnostics io.Writer) int {
	flags := flag.NewFlagSet("sdlc-validate", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	flags.Usage = func() { fmt.Fprint(diagnostics, helpText) }
	help := flags.Bool("help", false, "show help")
	flags.BoolVar(help, "h", false, "show help")
	version := flags.Bool("version", false, "show version")
	mode := flags.String("readiness-check", "", "test-code or delivery-code")
	spec := flags.String("spec", "spec.org", "specification path")
	validation := flags.String("validation", "", "validation path; defaults beside spec")
	if err := flags.Parse(args); err != nil {
		return failure(out, err)
	}
	if *help {
		fmt.Fprint(out, helpText)
		return 0
	}
	if *version {
		fmt.Fprintln(out, "sdlc-validate "+buildRelease)
		return 0
	}
	if flags.NArg() != 0 || !readiness.ValidMode(*mode) {
		return failure(out, fmt.Errorf("require --readiness-check=test-code or delivery-code; no positional arguments"))
	}
	if *validation == "" {
		*validation = filepath.Join(filepath.Dir(*spec), "validation.org")
	}
	report, err := readiness.Check(*spec, *validation, *mode)
	if err != nil {
		return failure(out, err)
	}
	if err := writeYAML(out, report); err != nil {
		fmt.Fprintln(diagnostics, err)
		return 3
	}
	return report.ExitCode()
}

func failure(out io.Writer, err error) int {
	_ = writeYAML(out, map[string]any{"audit-readiness": map[string]any{"error": map[string]string{"code": "EXECUTION_ERROR", "message": err.Error()}}})
	return 3
}

func writeYAML(out io.Writer, value any) error {
	encoder := yaml.NewEncoder(out)
	encoder.SetIndent(2)
	return encoder.Encode(value)
}
