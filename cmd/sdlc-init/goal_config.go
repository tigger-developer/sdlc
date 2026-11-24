package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/tigger-developer/sdlc/internal/projectinit"
)

// runGoalConfig resolves native continuation limits without initializing a project.
func runGoalConfig(arguments []string, output, diagnostics io.Writer) error {
	flags := flag.NewFlagSet("sdlc-init goal-config", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	project := flags.String("project", ".", "project root")
	global := flags.String("global-config", "/opt/sdlc/sdlc.yaml", "global SDLC YAML file")
	root := flags.String("sdlc-root", "/opt/sdlc/docs", "SDLC root containing the configuration schema")
	turns := flags.String("goal-max-turns", "", "explicit continuation limit")
	tokens := flags.String("goal-max-token-budget", "", "explicit token budget")
	flags.Usage = func() {
		fmt.Fprintln(diagnostics, "usage: sdlc-init goal-config [options]")
		fmt.Fprintln(diagnostics, "Resolve native goal limits without initializing or changing a project.")
		fmt.Fprintln(diagnostics, "Example: sdlc-init goal-config --project /path/to/project --goal-max-token-budget 100,000")
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
