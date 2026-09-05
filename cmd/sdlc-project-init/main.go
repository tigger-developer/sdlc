package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/tigger-developer/sdlc/internal/projectinit"
)

var buildRelease string

func main() {
	command := filepath.Base(os.Args[0])
	if command != "sdlc-project-update" {
		command = "sdlc-project-init"
	}
	if status := commandStatus(command, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); status != 0 {
		os.Exit(status)
	}
}

type usageError struct {
	err error
}

func (e *usageError) Error() string {
	return e.err.Error()
}

func commandStatus(command string, arguments []string, input io.Reader, output, errorOutput io.Writer) int {
	err := runCommand(command, arguments, input, output, errorOutput)
	if err == nil {
		return 0
	}
	fmt.Fprintf(errorOutput, "%s: %v\n", command, err)
	var invalid *usageError
	if errors.As(err, &invalid) {
		return 2
	}
	return 1
}

func runCommand(command string, arguments []string, input io.Reader, output, errorOutput io.Writer) error {
	if len(arguments) == 1 && (arguments[0] == "-version" || arguments[0] == "--version") {
		version := buildRelease
		if version == "" {
			version = "devel"
		}
		fmt.Fprintf(output, "%s %s\n", command, version)
		return nil
	}
	configuredRoot, err := bootstrapSDLCRoot(arguments)
	if err != nil {
		return err
	}
	schema, err := projectinit.LoadConfigSchema(configuredRoot)
	if err != nil {
		return err
	}
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(errorOutput)
	flags.Usage = func() {
		fmt.Fprintf(errorOutput, "usage: %s [options]\n", command)
		if command == "sdlc-project-update" {
			fmt.Fprintln(errorOutput, "Refresh project SDLC infrastructure and adopted revision without launching an agent.")
		}
		fmt.Fprintln(errorOutput, "Options accept one or two leading hyphens (for example, -project or --project).")
		flags.PrintDefaults()
		fmt.Fprintln(errorOutput, "  --version")
		fmt.Fprintln(errorOutput, "    \tprint the command version")
	}
	project := flags.String("project", ".", "project root")
	sdlcRoot := flags.String("sdlc-root", "", "canonical SDLC root (default ~/.agents/sdlc)")
	userConfig := flags.String("user-config", "", "user SDLC environment file")
	infra := flags.String("infra", "", "deprecated external infrastructure ownership: yes or no")
	noLaunch := flags.Bool("no-launch", false, "render the scaffold without invoking an agent harness")
	configured := make(map[string]*string, len(schema.Fields))
	for _, field := range schema.Fields {
		configured[field.Key] = flags.String(field.Flag, "", field.Help)
	}
	if err := flags.Parse(arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return &usageError{err: err}
	}
	if flags.NArg() != 0 {
		return &usageError{err: fmt.Errorf("unexpected positional arguments: %v", flags.Args())}
	}
	overrides := map[string]string{}
	for key, value := range configured {
		if *value != "" {
			overrides[key] = *value
		}
	}
	if *infra != "" {
		if overrides["SDLC_INFRA_ROLE"] != "" {
			return &usageError{err: errors.New("--infra and --infra-role cannot be used together")}
		}
		value, err := parseYesNo(*infra)
		if err != nil {
			return &usageError{err: err}
		}
		if value {
			overrides["SDLC_INFRA_ROLE"] = "consumer"
		} else {
			overrides["SDLC_INFRA_ROLE"] = "none"
		}
	}
	return projectinit.Run(projectinit.Options{
		ProjectRoot: *project, SDLCRoot: *sdlcRoot, UserConfigPath: *userConfig,
		Overrides:                  overrides,
		SDLCRevision:               sourceRevisionForCommand(command),
		NoLaunch:                   resolveNoLaunch(command, *noLaunch),
		UpdateConstitutionRevision: command == "sdlc-project-update",
		OfferLegacyMigration:       command == "sdlc-project-init",
		Input:                      input, Output: output, ErrorOutput: errorOutput,
	})
}

func resolveNoLaunch(command string, requested bool) bool {
	return requested || command == "sdlc-project-update"
}

func bootstrapSDLCRoot(arguments []string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving user home: %w", err)
	}
	root := filepath.Join(home, ".agents", "sdlc")
	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		if argument == "--sdlc-root" || argument == "-sdlc-root" {
			if index+1 >= len(arguments) {
				return "", &usageError{err: errors.New("--sdlc-root requires a value")}
			}
			root = arguments[index+1]
			index++
			continue
		}
		if strings.HasPrefix(argument, "--sdlc-root=") {
			root = strings.TrimPrefix(argument, "--sdlc-root=")
		} else if strings.HasPrefix(argument, "-sdlc-root=") {
			root = strings.TrimPrefix(argument, "-sdlc-root=")
		}
	}
	if root == "" {
		return "", &usageError{err: errors.New("--sdlc-root requires a non-empty value")}
	}
	return filepath.Abs(root)
}

func sourceRevisionForCommand(command string) string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	return sourceRevisionForCommandBuildInfo(command, buildInfo, buildRelease)
}

func sourceRevisionForCommandBuildInfo(command string, buildInfo *debug.BuildInfo, release string) string {
	if command == "sdlc-project-update" && release == "" {
		return ""
	}
	return sourceRevisionForBuildInfo(buildInfo, release)
}

func sourceRevisionFromBuildInfo(buildInfo *debug.BuildInfo) string {
	return sourceRevisionForBuildInfo(buildInfo, "")
}

func sourceRevisionForBuildInfo(buildInfo *debug.BuildInfo, release string) string {
	var revision string
	modified := false
	for _, setting := range buildInfo.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}
	if modified {
		return ""
	}
	if release != "" {
		return release
	}
	if revision != "" {
		return revision
	}
	if buildInfo.Main.Version != "" && buildInfo.Main.Version != "(devel)" {
		return buildInfo.Main.Version
	}
	return ""
}

func parseYesNo(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "yes", "true", "1":
		return true, nil
	case "no", "false", "0":
		return false, nil
	default:
		return false, fmt.Errorf("--infra expects yes or no, got %q", value)
	}
}
