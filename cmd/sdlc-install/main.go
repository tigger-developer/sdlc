package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tigger-developer/sdlc/internal/installer"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "sdlc-install: %v\n", err)
		os.Exit(1)
	}
}

func run(arguments []string, input io.Reader, output io.Writer) error {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolving current directory: %w", err)
	}
	flags := flag.NewFlagSet("sdlc-install", flag.ContinueOnError)
	flags.SetOutput(output)
	agent := flags.String("agent", "auto", "target agent: auto, claude, codex, copilot, hermes, or custom")
	agentHome := flags.String("agent-home", "", "provider home containing the sdlc link")
	source := flags.String("source", workingDirectory, "canonical SDLC clone")
	apply := flags.Bool("apply", false, "create the provider-home sdlc symlink")
	configure := flags.Bool("configure", false, "offer supported provider configuration changes for confirmation")
	if err := flags.Parse(arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	return installer.Run(installer.Options{
		Agent:     *agent,
		AgentHome: *agentHome,
		Source:    *source,
		Apply:     *apply,
		Configure: *configure,
		Input:     input,
		Output:    output,
	})
}
