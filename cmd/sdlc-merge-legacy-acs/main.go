package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/tigger-developer/sdlc/internal/projectinit"
)

var buildRelease string

func main() {
	if status := commandStatus(os.Args[1:], os.Stdout, os.Stderr); status != 0 {
		os.Exit(status)
	}
}

func commandStatus(arguments []string, output, errorOutput io.Writer) int {
	if err := runCommand(arguments, output, errorOutput); err != nil {
		fmt.Fprintf(errorOutput, "sdlc-merge-legacy-acs: %v\n", err)
		return 1
	}
	return 0
}

func runCommand(arguments []string, output, errorOutput io.Writer) error {
	flags := flag.NewFlagSet("sdlc-merge-legacy-acs", flag.ContinueOnError)
	flags.SetOutput(errorOutput)
	project := flags.String("project", ".", "project root")
	version := flags.Bool("version", false, "print the command version")
	flags.Usage = func() {
		fmt.Fprintln(errorOutput, "usage: sdlc-merge-legacy-acs [--project PATH]")
		fmt.Fprintln(errorOutput, "Merge docs/ACs.org into the folded legacy section of docs/work.org.")
		flags.PrintDefaults()
	}
	if err := flags.Parse(arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	if *version {
		fmt.Fprintf(output, "sdlc-merge-legacy-acs %s\n", sourceRevision())
		return nil
	}
	result, err := projectinit.MergeLegacyAcceptanceCriteria(*project)
	if err != nil {
		return err
	}
	if result.LedgerChanged || result.SourceRemoved {
		fmt.Fprintf(output, "Merged %d legacy acceptance criteria into docs/work.org.\n", result.AcceptanceCriteria)
	} else {
		fmt.Fprintln(output, "Legacy acceptance criteria are already current in docs/work.org.")
	}
	if result.ProfileChanged {
		fmt.Fprintln(output, "Updated .sdlc/project.yaml to use docs/work.org as its sole requirement authority.")
	}
	return nil
}

func sourceRevision() string {
	if buildRelease != "" {
		return buildRelease
	}
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "devel"
	}
	for _, setting := range buildInfo.Settings {
		if setting.Key == "vcs.revision" && setting.Value != "" {
			return setting.Value
		}
	}
	return "devel"
}
