package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tigger-developer/sdlc/internal/installer"
)

var buildRelease string

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "sdlc-install: %v\n", err)
		os.Exit(1)
	}
}

func run(arguments []string, input io.Reader, output io.Writer) error {
	return runAtRoot(arguments, input, output, "")
}

func runAtRoot(arguments []string, input io.Reader, output io.Writer, installRoot string) error {
	if err := rejectPrivilegedInstall(os.Geteuid()); err != nil {
		return err
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolving current directory: %w", err)
	}
	flags := flag.NewFlagSet("sdlc-install", flag.ContinueOnError)
	flags.SetOutput(output)
	agent := flags.String("agent", "", "target one agent: auto, claude, codex, copilot, hermes, goose, or custom; omit for interactive detection")
	agentHome := flags.String("agent-home", "", "provider home receiving native SDLC adapters")
	source := flags.String("source", workingDirectory, "staging SDLC clone")
	apply := flags.Bool("apply", false, "synchronize SDLC-owned copies for one provider")
	configure := flags.Bool("configure", false, "apply supported provider configuration changes")
	bare := flags.Bool("bare", false, "install /opt/sdlc only; no prompts, home links, discovery, or harness configuration")
	dryRun := flags.Bool("dry-run", false, "with --bare, preview canonical deployment without writing")
	version := flags.Bool("version", false, "print the command version")
	flags.Usage = func() {
		fmt.Fprintln(output, "usage: sdlc-install [options]")
		fmt.Fprintln(output, "Install or update the global SDLC framework from this source repository.")
		fmt.Fprintln(output, "")
		fmt.Fprintln(output, "This repository-internal command is invoked by make install. It deploys the")
		fmt.Fprintln(output, "shared standards, global skills, command guard, and supported provider")
		fmt.Fprintln(output, "adapters. It does not initialize a software project; run sdlc-init from")
		fmt.Fprintln(output, "that project's root instead.")
		fmt.Fprintln(output, "Without --agent, it asks which existing provider homes to use unless the")
		fmt.Fprintln(output, "top-level active-harness list in /opt/sdlc/sdlc.yaml supplies the selection.")
		fmt.Fprintln(output, "Use --bare for unattended canonical installation only; Nix can manage all")
		fmt.Fprintln(output, "links and provider configuration. No existing agent home is required.")
		fmt.Fprintln(output, "Pre-create writable /opt/sdlc if unprivileged creation is unavailable.")
		fmt.Fprintln(output, "Bare mode preserves existing global configuration, updates its release,")
		fmt.Fprintln(output, "and does not read legacy home configuration. Use --bare --dry-run to preview.")
		fmt.Fprintln(output, "")
		fmt.Fprintln(output, "Options:")
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
		release := buildRelease
		if release == "" {
			release = "devel"
		}
		fmt.Fprintf(output, "sdlc-install %s\n", release)
		return nil
	}
	if *bare {
		if *agent != "" || *agentHome != "" || *configure {
			return errors.New("--bare cannot be combined with provider selection or --configure")
		}
		return installer.RunBare(installer.Options{Source: *source, Apply: !*dryRun, Release: buildRelease, InstallRoot: installRoot, Input: input, Output: output})
	}
	if *dryRun {
		return errors.New("--dry-run requires --bare; provider mode previews by default without --apply")
	}
	if *agent == "" && *agentHome == "" {
		if *apply || *configure {
			return errors.New("--apply and --configure require an explicit --agent or --agent-home; omit them for interactive installation")
		}
		if installRoot != "" {
			return installer.RunInteractiveAtRoot(*source, "", installRoot, buildRelease, input, output)
		}
		return installer.RunInteractive(*source, "", buildRelease, input, output)
	}
	return installer.Run(installer.Options{
		Agent:       *agent,
		AgentHome:   *agentHome,
		Source:      *source,
		Apply:       *apply,
		Configure:   *configure,
		Release:     buildRelease,
		InstallRoot: installRoot,
		Input:       input,
		Output:      output,
	})
}

func rejectPrivilegedInstall(effectiveUID int) error {
	if effectiveUID == 0 {
		return errors.New("do not run the SDLC installer with sudo or as root; run it as the target user")
	}
	return nil
}
