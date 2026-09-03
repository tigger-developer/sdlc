package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/tigger-developer/sdlc/internal/projectinit"
)

var buildRelease string

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "sdlc-project-init: %v\n", err)
		os.Exit(1)
	}
}

func run(arguments []string) error {
	configuredRoot, err := bootstrapSDLCRoot(arguments)
	if err != nil {
		return err
	}
	schema, err := projectinit.LoadConfigSchema(configuredRoot)
	if err != nil {
		return err
	}
	flags := flag.NewFlagSet("sdlc-project-init", flag.ContinueOnError)
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
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	overrides := map[string]string{}
	for key, value := range configured {
		if *value != "" {
			overrides[key] = *value
		}
	}
	if *infra != "" {
		if overrides["SDLC_INFRA_ROLE"] != "" {
			return errors.New("--infra and --infra-role cannot be used together")
		}
		value, err := parseYesNo(*infra)
		if err != nil {
			return err
		}
		if value {
			overrides["SDLC_INFRA_ROLE"] = "consumer"
		} else {
			overrides["SDLC_INFRA_ROLE"] = "none"
		}
	}
	return projectinit.Run(projectinit.Options{
		ProjectRoot: *project, SDLCRoot: *sdlcRoot, UserConfigPath: *userConfig,
		Overrides:    overrides,
		SDLCRevision: sourceRevision(),
		NoLaunch:     *noLaunch, Input: os.Stdin, Output: os.Stdout, ErrorOutput: os.Stderr,
	})
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
				return "", errors.New("--sdlc-root requires a value")
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
		return "", errors.New("--sdlc-root requires a non-empty value")
	}
	return filepath.Abs(root)
}

func sourceRevision() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	return sourceRevisionForBuildInfo(buildInfo, buildRelease)
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
