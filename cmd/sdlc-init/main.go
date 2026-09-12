package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tigger-developer/sdlc/internal/projectinit"
)

var buildRelease string

func main() {
	if status := commandStatus(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); status != 0 {
		os.Exit(status)
	}
}

func commandStatus(arguments []string, input io.Reader, output, errorOutput io.Writer) int {
	err := runCommand(arguments, input, output, errorOutput)
	if err == nil {
		return 0
	}
	fmt.Fprintf(errorOutput, "sdlc-init: %v\n", err)
	return 1
}

func runCommand(arguments []string, input io.Reader, output, errorOutput io.Writer) error {
	if len(arguments) == 1 && (arguments[0] == "-version" || arguments[0] == "--version") {
		version := buildRelease
		if version == "" {
			version = "devel"
		}
		fmt.Fprintf(output, "sdlc-init %s\n", version)
		return nil
	}
	if requestsHelp(arguments) {
		printBootstrapHelp(errorOutput)
		return nil
	}
	sdlcRoot, err := bootstrapSDLCRoot(arguments)
	if err != nil {
		return err
	}
	schema, err := projectinit.LoadConfigSchema(sdlcRoot)
	if err != nil {
		return err
	}
	flags := flag.NewFlagSet("sdlc-init", flag.ContinueOnError)
	flags.SetOutput(errorOutput)
	flags.Usage = func() {
		fmt.Fprintln(errorOutput, "usage: sdlc-init [options]")
		printInitPurpose(errorOutput)
		flags.PrintDefaults()
		fmt.Fprintln(errorOutput, "  --version\n    \tprint the command version")
	}
	project := flags.String("project", ".", "project root")
	configuredRoot := flags.String("sdlc-root", sdlcRoot, "canonical SDLC root")
	globalConfig := flags.String("global-config", "", "global SDLC YAML configuration")
	overrideGlobalConfig := flags.Bool("override-global-config", false, "prompt for project overrides to populated global defaults")
	noAgentScan := flags.Bool("no-agent-scan", false, "skip semantic classification of archived Spec Kit work")
	overrideFlags := map[string]*string{}
	for _, field := range schema.Fields {
		overrideFlags[field.Key] = flags.String(field.Flag, "", field.Help)
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
	overrides := map[string]string{}
	for key, value := range overrideFlags {
		if strings.TrimSpace(*value) != "" {
			overrides[key] = *value
		}
	}
	return projectinit.Run(projectinit.Options{
		ProjectRoot:          *project,
		SDLCRoot:             *configuredRoot,
		GlobalConfigPath:     *globalConfig,
		OverrideGlobalConfig: *overrideGlobalConfig,
		SkipAgentScans:       *noAgentScan,
		Overrides:            overrides,
		Input:                input,
		Output:               output,
		ErrorOutput:          errorOutput,
	})
}

func requestsHelp(arguments []string) bool {
	for _, argument := range arguments {
		if argument == "-h" || argument == "--help" {
			return true
		}
	}
	return false
}

func printBootstrapHelp(output io.Writer) {
	fmt.Fprintln(output, "usage: sdlc-init [options]")
	printInitPurpose(output)
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Core options:")
	fmt.Fprintln(output, "  --project PATH        project root (default: current directory)")
	fmt.Fprintln(output, "  --sdlc-root PATH      canonical SDLC root (default: ~/.agents/sdlc)")
	fmt.Fprintln(output, "  --global-config PATH  global SDLC YAML configuration")
	fmt.Fprintln(output, "  --override-global-config")
	fmt.Fprintln(output, "                        prompt for project overrides to populated global defaults")
	fmt.Fprintln(output, "  --no-agent-scan      skip semantic classification of archived Spec Kit work")
	fmt.Fprintln(output, "  --version             print the command version")
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Project-setting options are generated from config/project-init.schema.yaml.")
}

func printInitPurpose(output io.Writer) {
	fmt.Fprintln(output, "Initialize or migrate one Git project to the installed SDLC framework.")
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Run it once from the project root. It creates the project profile and work")
	fmt.Fprintln(output, "ledger, optionally migrates legacy tickets, and preserves prior SDLC state.")
	fmt.Fprintln(output, "It does not install or update the global SDLC framework; run make install")
	fmt.Fprintln(output, "from the SDLC source repository for that operation.")
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "After interruption, rerun on the migration branch. From master or main,")
	fmt.Fprintln(output, "confirm the proposed switch to the migration branch with y; n leaves it unchanged.")
	fmt.Fprintln(output, "Ambiguous branch history requires selecting the intended migration branch first.")
}

func bootstrapSDLCRoot(arguments []string) (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	root := filepath.Join(userHome, ".agents", "sdlc")
	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		if argument == "--sdlc-root" || argument == "-sdlc-root" {
			if index+1 >= len(arguments) {
				return "", errors.New("--sdlc-root requires a value")
			}
			root = arguments[index+1]
			index++
		} else if strings.HasPrefix(argument, "--sdlc-root=") {
			root = strings.TrimPrefix(argument, "--sdlc-root=")
		}
	}
	return filepath.Abs(root)
}
