package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/tigger-developer/sdlc/internal/projectinit"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "sdlc-project-init: %v\n", err)
		os.Exit(1)
	}
}

func run(arguments []string) error {
	flags := flag.NewFlagSet("sdlc-project-init", flag.ContinueOnError)
	project := flags.String("project", ".", "project root")
	sdlcRoot := flags.String("sdlc-root", "", "canonical SDLC root (default ~/.agents/sdlc)")
	userConfig := flags.String("user-config", "", "user SDLC environment file")
	harness := flags.String("harness", "", "agent harness: codex, claude, or hermes")
	deliveryProvider := flags.String("delivery-provider", "", "delivery model provider")
	deliveryModel := flags.String("delivery-model", "", "delivery model")
	auditProvider := flags.String("audit-provider", "", "independent audit model provider")
	auditModel := flags.String("audit-model", "", "independent audit model")
	technologies := flags.String("technologies", "", "comma-separated technology standards")
	infra := flags.String("infra", "", "external infrastructure ownership: yes or no")
	infraOwner := flags.String("infra-owner", "", "external infrastructure owner descriptor")
	infraContract := flags.String("infra-contract", "", "external infrastructure integration-contract path")
	noLaunch := flags.Bool("no-launch", false, "render the baseline without invoking an agent harness")
	if err := flags.Parse(arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	var infraEnabled *bool
	if *infra != "" {
		value, err := parseYesNo(*infra)
		if err != nil {
			return err
		}
		infraEnabled = &value
	}
	return projectinit.Run(projectinit.Options{
		ProjectRoot: *project, SDLCRoot: *sdlcRoot, UserConfigPath: *userConfig,
		Harness: *harness, DeliveryProvider: *deliveryProvider, DeliveryModel: *deliveryModel,
		AuditProvider: *auditProvider, AuditModel: *auditModel, Technologies: splitList(*technologies),
		InfraEnabled: infraEnabled, InfraOwner: *infraOwner, InfraContract: *infraContract,
		SDLCRevision: sourceRevision(),
		NoLaunch:     *noLaunch, Input: os.Stdin, Output: os.Stdout, ErrorOutput: os.Stderr,
	})
}

func sourceRevision() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	return sourceRevisionFromBuildInfo(buildInfo)
}

func sourceRevisionFromBuildInfo(buildInfo *debug.BuildInfo) string {
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
	if revision != "" {
		return revision
	}
	if buildInfo.Main.Version != "" && buildInfo.Main.Version != "(devel)" {
		return buildInfo.Main.Version
	}
	return ""
}

func splitList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	var result []string
	for _, item := range strings.Split(value, ",") {
		result = append(result, strings.TrimSpace(item))
	}
	return result
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
