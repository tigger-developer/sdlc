// ABOUTME: Exposes validated goal budgets to the current delivery skill.
// ABOUTME: This read-only operation never starts a provider or claims goal activation.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tigger-developer/sdlc/internal/projectinit"
)

func runGoalConfig(arguments []string, output, diagnostics io.Writer) error {
	flags := flag.NewFlagSet("sdlc-harness goal-config", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	project := flags.String("project", ".", "project root")
	global := flags.String("global-config", "", "global SDLC YAML file")
	root := flags.String("sdlc-root", "", "SDLC root containing the configuration schema")
	turns := flags.String("goal-max-turns", "", "explicit continuation limit")
	tokens := flags.String("goal-max-token-budget", "", "explicit token budget")
	flags.Usage = func() { printTopLevelHelp(diagnostics) }
	if err := flags.Parse(arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	if *root == "" {
		executable, err := os.Executable()
		if err != nil {
			return err
		}
		*root = filepath.Join(filepath.Dir(executable), "..")
	}
	if *global == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		*global = filepath.Join(home, ".agents", "sdlc.yaml")
	}
	schema, err := projectinit.LoadConfigSchema(*root)
	if err != nil {
		return err
	}
	overrides := map[string]string{}
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "goal-max-turns":
			overrides["SDLC_GOAL_MAX_TURNS"] = *turns
		case "goal-max-token-budget":
			overrides["SDLC_GOAL_MAX_TOKEN_BUDGET"] = *tokens
		}
	})
	limits, err := projectinit.ResolveGoalLimits(schema, *project, *global, overrides)
	if err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(limits)
}
